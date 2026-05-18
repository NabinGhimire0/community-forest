package samiti

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"forest-management/internal/models"

	"gorm.io/gorm"
)

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
	if input.TenureStart != nil {
		t, err := time.Parse("2006-01-02", *input.TenureStart)
		if err == nil {
			tenureStart = &t
		}
	}
	if input.TenureEnd != nil {
		t, err := time.Parse("2006-01-02", *input.TenureEnd)
		if err == nil {
			tenureEnd = &t
		}
	}

	head := models.SamitiHead{
		Name:        input.Name,
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

	if err := s.db.Create(&head).Error; err != nil {
		return nil, fmt.Errorf("failed to create committee member: %w", err)
	}

	return &head, nil
}

func (s *SamitiService) ListHeads() ([]models.SamitiHead, error) {
	var heads []models.SamitiHead
	// Order by post priority: chairperson first, then secretary, treasurer, member
	err := s.db.Order(`
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

func (s *SamitiService) GetHeadByID(id uint) (*models.SamitiHead, error) {
	var head models.SamitiHead
	err := s.db.First(&head, id).Error
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

	s.db.First(&head, id)
	return &head, nil
}

func (s *SamitiService) DeleteHead(id uint) error {
	return s.db.Delete(&models.SamitiHead{}, id).Error
}

// ==================== Logo Upload ====================

func (s *SamitiService) UploadLogo(file io.Reader, filename string) (string, error) {
	// Create uploads directory if not exists
	uploadDir := "./uploads/samiti"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".png"
	}
	uniqueName := fmt.Sprintf("logo_%d%s", time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	fileURL := fmt.Sprintf("/uploads/samiti/%s", uniqueName)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Update settings with logo URL
	var settings models.SamitiSetting
	if err := s.db.First(&settings).Error; err != nil {
		// Create settings if not exists
		settings = models.SamitiSetting{
			Name:         "Community Forestry Samiti",
			Address:      "Not Set",
			WardNo:       1,
			Municipality: "Not Set",
			District:     "Not Set",
			Province:     "Not Set",
			Logo:         &fileURL,
		}
		if err := s.db.Create(&settings).Error; err != nil {
			os.Remove(filePath)
			return "", fmt.Errorf("failed to create settings: %w", err)
		}
	} else {
		if err := s.db.Model(&settings).Update("logo", fileURL).Error; err != nil {
			os.Remove(filePath)
			return "", fmt.Errorf("failed to update settings: %w", err)
		}
	}

	return fileURL, nil
}

// ==================== Head Photo Upload ====================

func (s *SamitiService) UploadHeadPhoto(headID uint, file io.Reader, filename string) (string, error) {
	var head models.SamitiHead
	if err := s.db.First(&head, headID).Error; err != nil {
		return "", errors.New("committee member not found")
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/heads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	uniqueName := fmt.Sprintf("head_%d_%d%s", headID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	fileURL := fmt.Sprintf("/uploads/heads/%s", uniqueName)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Update head with photo URL
	if err := s.db.Model(&head).Update("photo", fileURL).Error; err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to update head: %w", err)
	}

	return fileURL, nil
}
