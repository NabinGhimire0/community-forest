package samiti

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"forest-management/internal/models"
	"forest-management/pkg/fileutil"
	"forest-management/pkg/security"
	"forest-management/pkg/utils"

	"gorm.io/gorm"
)

// PublicSamitiHead is the deliberately limited representation exposed on the
// unauthenticated landing page. Private contact details, internal remarks and
// linked login identifiers are never serialized publicly.
type PublicSamitiHead struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Post        string     `json:"post"`
	Phone       *string    `json:"phone,omitempty"`
	Photo       *string    `json:"photo,omitempty"`
	TenureStart *time.Time `json:"tenure_start,omitempty"`
	TenureEnd   *time.Time `json:"tenure_end,omitempty"`
	IsActive    bool       `json:"is_active"`
}

func publicHead(head models.SamitiHead) PublicSamitiHead {
	return PublicSamitiHead{
		ID: head.ID, Name: head.Name, Post: head.Post, Phone: head.Phone,
		Photo: head.Photo, TenureStart: head.TenureStart, TenureEnd: head.TenureEnd, IsActive: head.IsActive,
	}
}

type SamitiService struct {
	db *gorm.DB
}

func NewSamitiService(db *gorm.DB) *SamitiService {
	return &SamitiService{db: db}
}

// ==================== Samiti Settings ====================

func (s *SamitiService) GetSettings() (*models.SamitiSetting, error) {
	var settings models.SamitiSetting
	err := s.db.First(&settings).Error
	if err != nil {
		// Return default settings if not found
		return &models.SamitiSetting{
			Name:         "Community Forestry Samiti",
			Address:      "Not Set",
			WardNo:       1,
			Municipality: "Not Set",
			District:     "Not Set",
			Province:     "Not Set",
		}, nil
	}
	return &settings, err
}

func (s *SamitiService) UpdateSettings(input UpdateSettingsInput) (*models.SamitiSetting, error) {
	var settings models.SamitiSetting

	// FirstOrCreate - creates if not exists
	result := s.db.First(&settings)
	if result.Error != nil {
		// Create initial settings
		settings = models.SamitiSetting{
			Name:         "Community Forestry Samiti",
			Address:      "Not Set",
			WardNo:       1,
			Municipality: "Not Set",
			District:     "Not Set",
			Province:     "Not Set",
		}
		s.db.Create(&settings)
	}

	// Build updates map
	updates := make(map[string]interface{})
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.RegistrationNo != nil {
		updates["registration_no"] = *input.RegistrationNo
	}
	if input.Address != nil {
		updates["address"] = *input.Address
	}
	if input.WardNo != nil {
		updates["ward_no"] = *input.WardNo
	}
	if input.Municipality != nil {
		updates["municipality"] = *input.Municipality
	}
	if input.District != nil {
		updates["district"] = *input.District
	}
	if input.Province != nil {
		updates["province"] = *input.Province
	}
	if input.ContactPhone != nil {
		updates["contact_phone"] = *input.ContactPhone
	}
	if input.ContactEmail != nil {
		updates["contact_email"] = *input.ContactEmail
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Logo != nil {
		updates["logo"] = *input.Logo
	}
	if input.MapImage != nil {
		updates["map_image"] = *input.MapImage
	}
	if input.Latitude != nil {
		updates["latitude"] = *input.Latitude
	}
	if input.Longitude != nil {
		updates["longitude"] = *input.Longitude
	}
	if input.EstablishedDate != nil {
		t, err := time.Parse("2006-01-02", *input.EstablishedDate)
		if err == nil {
			updates["established_date"] = t
		}
	}

	if len(updates) > 0 {
		if err := s.db.Model(&settings).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update settings: %w", err)
		}
	}

	s.db.First(&settings, settings.ID)
	return &settings, nil
}

// ==================== Samiti Heads ====================

func (s *SamitiService) CreateHead(input CreateHeadInput) (*models.SamitiHead, error) {
	var tenureStart, tenureEnd *time.Time
	if input.TenureStart != nil && strings.TrimSpace(*input.TenureStart) != "" {
		t, err := time.Parse("2006-01-02", *input.TenureStart)
		if err != nil {
			return nil, errors.New("invalid tenure start date")
		}
		tenureStart = &t
	}
	if input.TenureEnd != nil && strings.TrimSpace(*input.TenureEnd) != "" {
		t, err := time.Parse("2006-01-02", *input.TenureEnd)
		if err != nil {
			return nil, errors.New("invalid tenure end date")
		}
		tenureEnd = &t
	}
	if tenureStart != nil && tenureEnd != nil && tenureEnd.Before(*tenureStart) {
		return nil, errors.New("tenure end date must be after tenure start date")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			tx.Rollback()
			panic(recovered)
		}
	}()

	var userID *uint
	bootstrapDeactivated := false
	existingAccountPromoted := false
	if input.CreateLogin {
		role := strings.ToLower(strings.TrimSpace(input.AccountRole))
		if role != "admin" && role != "staff" {
			tx.Rollback()
			return nil, errors.New("committee login role must be admin or staff")
		}
		phone := ""
		if input.Phone != nil {
			phone = strings.TrimSpace(*input.Phone)
		}
		if phone == "" {
			tx.Rollback()
			return nil, errors.New("phone is required when creating login credentials")
		}
		normalizedPhone, err := security.NormalizeNepalMobile(phone)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		phone = normalizedPhone
		if err := security.ValidatePassword(input.AccountPassword); err != nil {
			tx.Rollback()
			return nil, err
		}

		hashedPassword, err := utils.HashPassword(input.AccountPassword)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to secure account password: %w", err)
		}

		var email *string
		if input.Email != nil && strings.TrimSpace(*input.Email) != "" {
			normalized := strings.TrimSpace(*input.Email)
			email = &normalized
		}

		// Committee heads are often already registered as ordinary members. When
		// the phone belongs to an existing user, promote that account instead of
		// creating a duplicate phone/login. The member row remains linked to the
		// same user; the selected operational role controls the dashboard.
		var user models.User
		findUser := tx.Where("phone = ?", phone).First(&user)
		if findUser.Error == nil {
			var linkedHeadCount int64
			if err := tx.Model(&models.SamitiHead{}).Where("user_id = ?", user.ID).Count(&linkedHeadCount).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			if linkedHeadCount > 0 {
				tx.Rollback()
				return nil, errors.New("this login account is already linked to a committee member")
			}

			now := time.Now().UTC()
			updates := map[string]interface{}{
				"name":                  strings.TrimSpace(input.Name),
				"password":              hashedPassword,
				"role":                  role,
				"status":                "active",
				"is_bootstrap_admin":    false,
				"must_change_password":  true,
				"password_changed_at":   now,
				"failed_login_attempts": 0,
				"locked_until":          nil,
			}
			if email != nil {
				updates["email"] = *email
			}
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				tx.Rollback()
				return nil, errors.New("failed to update existing login account: email may already be registered")
			}
			if err := tx.Model(&models.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", user.ID).Update("revoked_at", now).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			existingAccountPromoted = true
		} else if errors.Is(findUser.Error, gorm.ErrRecordNotFound) {
			user = models.User{
				Name:               strings.TrimSpace(input.Name),
				Phone:              phone,
				Email:              email,
				Password:           hashedPassword,
				Role:               role,
				Status:             "active",
				MustChangePassword: true,
			}
			if err := tx.Create(&user).Error; err != nil {
				tx.Rollback()
				return nil, errors.New("failed to create login account: phone or email may already be registered")
			}
		} else {
			tx.Rollback()
			return nil, findUser.Error
		}
		userID = &user.ID
		input.Phone = &phone
		if email != nil {
			input.Email = email
		}

		// The environment-seeded account is only a bootstrap account. Once a real
		// committee administrator exists, deactivate every other bootstrap admin
		// while preserving its audit history.
		if role == "admin" {
			result := tx.Model(&models.User{}).
				Where("is_bootstrap_admin = ? AND id <> ? AND status = ?", true, user.ID, "active").
				Update("status", "inactive")
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			bootstrapDeactivated = result.RowsAffected > 0
		}
	}

	head := models.SamitiHead{
		UserID:      userID,
		Name:        strings.TrimSpace(input.Name),
		Post:        input.Post,
		Phone:       input.Phone,
		Email:       input.Email,
		Address:     input.Address,
		Photo:       input.Photo,
		TenureStart: tenureStart,
		TenureEnd:   tenureEnd,
		IsActive:    true,
		Remarks:     input.Remarks,
	}
	if input.IsActive != nil {
		head.IsActive = *input.IsActive
	}

	if err := tx.Create(&head).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create committee member: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.db.Preload("User").First(&head, head.ID).Error; err != nil {
		return nil, err
	}
	head.BootstrapAdminDeactivated = bootstrapDeactivated
	head.ExistingAccountPromoted = existingAccountPromoted
	return &head, nil
}

func (s *SamitiService) ListHeads() ([]PublicSamitiHead, error) {
	var heads []models.SamitiHead
	err := s.db.Where("is_active = ?", true).Order(`
        CASE post
            WHEN 'chairperson' THEN 1
            WHEN 'secretary'   THEN 2
            WHEN 'treasurer'   THEN 3
            WHEN 'member'      THEN 4
            ELSE 5
        END ASC
    `).Find(&heads).Error
	if err != nil {
		return nil, err
	}
	result := make([]PublicSamitiHead, 0, len(heads))
	for _, head := range heads {
		result = append(result, publicHead(head))
	}
	return result, nil
}

func (s *SamitiService) GetHeadByID(id uint) (*PublicSamitiHead, error) {
	var head models.SamitiHead
	if err := s.db.Where("is_active = ?", true).First(&head, id).Error; err != nil {
		return nil, err
	}
	result := publicHead(head)
	return &result, nil
}

func (s *SamitiService) ListHeadsForAdmin() ([]models.SamitiHead, error) {
	var heads []models.SamitiHead
	err := s.db.Preload("User").Order(`
        CASE post
            WHEN 'chairperson' THEN 1
            WHEN 'secretary'   THEN 2
            WHEN 'treasurer'   THEN 3
            WHEN 'member'      THEN 4
            ELSE 5
        END ASC
    `).Find(&heads).Error
	return heads, err
}

func (s *SamitiService) GetHeadForAdmin(id uint) (*models.SamitiHead, error) {
	var head models.SamitiHead
	err := s.db.Preload("User").First(&head, id).Error
	return &head, err
}

func (s *SamitiService) UpdateHead(id uint, input UpdateHeadInput) (*models.SamitiHead, error) {
	var head models.SamitiHead
	if err := s.db.First(&head, id).Error; err != nil {
		return nil, errors.New("committee member not found")
	}

	updates := make(map[string]interface{})

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Post != nil {
		updates["post"] = *input.Post
	}
	if input.Phone != nil {
		updates["phone"] = *input.Phone
	}
	if input.Email != nil {
		updates["email"] = *input.Email
	}
	if input.Address != nil {
		updates["address"] = *input.Address
	}
	if input.Photo != nil {
		updates["photo"] = *input.Photo
	}
	if input.TenureStart != nil {
		var tenureStart *time.Time
		if *input.TenureStart != "" {
			t, err := time.Parse("2006-01-02", *input.TenureStart)
			if err == nil {
				tenureStart = &t
			}
		}
		updates["tenure_start"] = tenureStart
	}
	if input.TenureEnd != nil {
		var tenureEnd *time.Time
		if *input.TenureEnd != "" {
			t, err := time.Parse("2006-01-02", *input.TenureEnd)
			if err == nil {
				tenureEnd = &t
			}
		}
		updates["tenure_end"] = tenureEnd
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}
	if input.Remarks != nil {
		updates["remarks"] = *input.Remarks
	}

	if len(updates) > 0 {
		if err := s.db.Model(&head).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update committee member: %w", err)
		}
	}

	if input.IsActive != nil && head.UserID != nil {
		status := "inactive"
		if *input.IsActive {
			status = "active"
		}
		if err := s.db.Model(&models.User{}).Where("id = ?", *head.UserID).Update("status", status).Error; err != nil {
			return nil, err
		}
	}

	s.db.Preload("User").First(&head, id)
	return &head, nil
}

func (s *SamitiService) DeleteHead(id uint) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	var head models.SamitiHead
	if err := tx.First(&head, id).Error; err != nil {
		tx.Rollback()
		return errors.New("committee member not found")
	}
	if !head.IsActive {
		tx.Rollback()
		return errors.New("committee member is already inactive")
	}
	if head.UserID != nil {
		if err := tx.Model(&models.User{}).Where("id = ?", *head.UserID).Updates(map[string]interface{}{
			"status": "inactive",
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
		now := time.Now().UTC()
		if err := tx.Model(&models.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", *head.UserID).Update("revoked_at", now).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	// Preserve committee history and its linked account for audit/reference.
	if err := tx.Model(&head).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// ==================== Logo Upload ====================

func (s *SamitiService) UploadLogo(file io.Reader, filename string) (string, error) {
	saved, err := fileutil.Save(file, "samiti", "logo", fileutil.ImagePolicy)
	if err != nil {
		return "", err
	}
	fileURL := saved.URL
	var settings models.SamitiSetting
	if err := s.db.First(&settings).Error; err != nil {
		settings = models.SamitiSetting{Name: "Community Forestry Samiti", Address: "Not Set", WardNo: 1, Municipality: "Not Set", District: "Not Set", Province: "Not Set", Logo: &fileURL}
		if err := s.db.Create(&settings).Error; err != nil {
			_ = os.Remove(saved.Path)
			return "", fmt.Errorf("failed to create settings: %w", err)
		}
	} else if err := s.db.Model(&settings).Update("logo", fileURL).Error; err != nil {
		_ = os.Remove(saved.Path)
		return "", fmt.Errorf("failed to update settings: %w", err)
	}
	return fileURL, nil
}

// ==================== Head Photo Upload ====================

func (s *SamitiService) UploadHeadPhoto(headID uint, file io.Reader, filename string) (string, error) {
	var head models.SamitiHead
	if err := s.db.First(&head, headID).Error; err != nil {
		return "", errors.New("committee member not found")
	}
	saved, err := fileutil.Save(file, "heads", fmt.Sprintf("head-%d", headID), fileutil.ImagePolicy)
	if err != nil {
		return "", err
	}
	if err := s.db.Model(&head).Update("photo", saved.URL).Error; err != nil {
		_ = os.Remove(saved.Path)
		return "", fmt.Errorf("failed to update head: %w", err)
	}
	return saved.URL, nil
}
