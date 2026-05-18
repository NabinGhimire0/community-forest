package members

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"forest-management/internal/audit"
	"forest-management/internal/models"
	"forest-management/pkg/response"
	"forest-management/pkg/utils"

	"gorm.io/gorm"
)

type MemberService struct {
	db  *gorm.DB
	sms utils.SMSService
}

func NewMemberService(db *gorm.DB) *MemberService {
	return &MemberService{
		db:  db,
		sms: utils.GetSMSService(),
	}
}

// Credentials holds the plain-text credentials (only shown once to admin)
type Credentials struct {
	Phone         string
	PlainPassword string
}

// generateRandomPassword creates a cryptographically secure random password
func generateRandomPassword(length int) string {
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789!@#$%"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[n.Int64()]
	}
	return string(result)
}

// CreateMember creates a member AND user credentials, then sends SMS
func (s *MemberService) CreateMember(req CreateMemberRequest) (*models.Member, *Credentials, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Determine membership number
	membershipNo := req.MembershipNo
	if membershipNo == "" {
		var err error
		membershipNo, err = s.generateMembershipNo(tx)
		if err != nil {
			tx.Rollback()
			return nil, nil, fmt.Errorf("failed to generate membership number: %w", err)
		}
	}

	// 2. Generate a random password (8 characters)
	plainPassword := generateRandomPassword(8)

	// 3. Hash the password
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Create the User record
	phoneStr := ""
	if req.Phone != nil {
		phoneStr = *req.Phone
	}
	user := models.User{
		Name:     req.Name,
		Phone:    phoneStr,
		Password: hashedPassword,
		Role:     "member",
		Status:   "active",
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create user: phone number may already be registered")
	}

	// 5. Parse joined date if provided
	var joinedDate *time.Time
	if req.JoinedDate != nil {
		t, err := time.Parse("2006-01-02", *req.JoinedDate)
		if err == nil {
			joinedDate = &t
		}
	}

	// 6. Create the Member record
	member := models.Member{
		UserID:         &user.ID,
		MembershipNo:   membershipNo,
		Name:           req.Name,
		AssistantName:  req.AssistantName,
		FatherName:     req.FatherName,
		WardNo:         req.WardNo,
		Tole:           req.Tole,
		Phone:          req.Phone,
		Photo:          req.Photo,
		AssistantPhoto: req.AssistantPhoto,
		JoinedDate:     joinedDate,
		Status:         "active",
		Remarks:        req.Remarks,
	}

	// Add family members if provided
	for _, fmReq := range req.FamilyMembers {
		member.FamilyMembers = append(member.FamilyMembers, models.FamilyMember{
			Name:          fmReq.Name,
			Relation:      fmReq.Relation,
			Age:           fmReq.Age,
			Gender:        fmReq.Gender,
			CitizenshipNo: fmReq.CitizenshipNo,
			Remarks:       fmReq.Remarks,
		})
	}

	if err := tx.Create(&member).Error; err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to create member: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	audit.CreateAuditEntry(
		s.db,
		nil,
		"create",
		"member",
		&member.ID,
		nil,
		map[string]interface{}{
			"id":            member.ID,
			"membership_no": member.MembershipNo,
			"name":          member.Name,
			"phone":         phoneStr,
		},
		"",
		"",
		"Member created with auto-generated credentials",
	)

	// Send SMS with credentials (non-blocking)
	go s.sendCredentialsSMS(phoneStr, plainPassword)

	credentials := &Credentials{
		Phone:         phoneStr,
		PlainPassword: plainPassword,
	}

	return &member, credentials, nil
}

// generateMembershipNo creates the next sequential membership number
func (s *MemberService) generateMembershipNo(tx *gorm.DB) (string, error) {
	var lastMember models.Member
	err := tx.Order("id DESC").First(&lastMember).Error

	var nextNum int
	if err != nil {
		nextNum = 1
	} else {
		var lastNum int
		_, err := fmt.Sscanf(lastMember.MembershipNo, "MEM-%04d", &lastNum)
		if err != nil {
			var count int64
			tx.Model(&models.Member{}).Count(&count)
			nextNum = int(count) + 1
		} else {
			nextNum = lastNum + 1
		}
	}

	return fmt.Sprintf("MEM-%04d", nextNum), nil
}

// sendCredentialsSMS sends login credentials to the member via SMS
func (s *MemberService) sendCredentialsSMS(phone, password string) {
	message := fmt.Sprintf(
		"Ban Samiti: Your account has been created.\nPhone: %s\nPassword: %s\nLogin at your portal.",
		phone, password,
	)
	if err := s.sms.SendSMS(phone, message); err != nil {
		fmt.Printf("⚠️ Failed to send SMS to %s: %v\n", phone, err)
	}
}

// ListMembers returns paginated members with optional search
func (s *MemberService) ListMembers(page, perPage int, search, status string) ([]models.Member, *response.Pagination, error) {
	var members []models.Member
	var total int64

	query := s.db.Model(&models.Member{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR membership_no ILIKE ? OR phone ILIKE ? OR father_name ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("User").
		Preload("FamilyMembers").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&members).Error

	meta := &response.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}

	return members, meta, err
}

// GetMemberByID returns a single member with all relations
func (s *MemberService) GetMemberByID(id uint) (*models.Member, error) {
	var member models.Member
	err := s.db.
		Preload("User").
		Preload("FamilyMembers").
		First(&member, id).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// GetMemberByUserID finds a member by their linked user ID
func (s *MemberService) GetMemberByUserID(userID uint) (*models.Member, error) {
	var member models.Member
	err := s.db.
		Preload("FamilyMembers").
		Where("user_id = ?", userID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// UpdateMember updates a member's details
func (s *MemberService) UpdateMember(id uint, req UpdateMemberRequest, reqUserID *uint) (*models.Member, error) {
	var member models.Member

	if err := s.db.First(&member, id).Error; err != nil {
		return nil, errors.New("member not found")
	}

	// Store old values for audit log
	oldValues := map[string]interface{}{
		"name":           member.Name,
		"assistant_name": member.AssistantName,
		"father_name":    member.FatherName,
		"ward_no":        member.WardNo,
		"tole":           member.Tole,
		"phone":          member.Phone,
		"status":         member.Status,
	}

	// Update fields
	member.Name = req.Name
	member.AssistantName = req.AssistantName
	member.FatherName = req.FatherName
	member.WardNo = req.WardNo
	member.Tole = req.Tole

	if req.Phone != nil {
		member.Phone = req.Phone
	}

	if req.Photo != nil {
		member.Photo = req.Photo
	}

	if req.AssistantPhoto != nil {
		member.AssistantPhoto = req.AssistantPhoto
	}

	if req.Status != nil {
		member.Status = *req.Status
	}

	if req.Remarks != nil {
		member.Remarks = req.Remarks
	}

	if req.MembershipNo != "" {
		member.MembershipNo = req.MembershipNo
	}

	// Also update the linked User's name and phone
	if member.UserID != nil && req.Phone != nil {
		s.db.Model(&models.User{}).
			Where("id = ?", member.UserID).
			Updates(map[string]interface{}{
				"name":  req.Name,
				"phone": *req.Phone,
			})
	}

	// Save member
	if err := s.db.Save(&member).Error; err != nil {
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	// Update family members if provided
	if req.FamilyMembers != nil {
		var familyMembers []models.FamilyMember
		for _, fmReq := range req.FamilyMembers {
			familyMembers = append(familyMembers, models.FamilyMember{
				MemberID:      member.ID,
				Name:          fmReq.Name,
				Relation:      fmReq.Relation,
				Age:           fmReq.Age,
				Gender:        fmReq.Gender,
				CitizenshipNo: fmReq.CitizenshipNo,
				Remarks:       fmReq.Remarks,
			})
		}

		if len(familyMembers) > 0 {
			s.db.Model(&member).Association("FamilyMembers").Replace(familyMembers)
		} else {
			s.db.Model(&member).Association("FamilyMembers").Clear()
		}
	}

	// Audit log
	audit.CreateAuditEntry(
		s.db,
		reqUserID,
		"update",
		"member",
		&member.ID,
		oldValues,
		map[string]interface{}{
			"name":           member.Name,
			"assistant_name": member.AssistantName,
			"father_name":    member.FatherName,
			"ward_no":        member.WardNo,
			"tole":           member.Tole,
			"phone":          member.Phone,
			"status":         member.Status,
		},
		"",
		"",
		"Member updated",
	)

	return &member, nil
}

// DeleteMember soft-deletes a member (and their user account)
func (s *MemberService) DeleteMember(id uint) error {
	var member models.Member
	if err := s.db.First(&member, id).Error; err != nil {
		return errors.New("member not found")
	}

	tx := s.db.Begin()

	// Delete family members first
	if err := tx.Where("member_id = ?", id).Delete(&models.FamilyMember{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Delete the user account
	if member.UserID != nil {
		if err := tx.Delete(&models.User{}, member.UserID).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Delete the member
	if err := tx.Delete(&member).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// AddFamilyMember adds a family member to a member's family
func (s *MemberService) AddFamilyMember(memberID uint, req FamilyMemberRequest) (*models.FamilyMember, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return nil, errors.New("member not found")
	}

	familyMember := models.FamilyMember{
		MemberID:      memberID,
		Name:          req.Name,
		Relation:      req.Relation,
		Age:           req.Age,
		Gender:        req.Gender,
		CitizenshipNo: req.CitizenshipNo,
		Remarks:       req.Remarks,
	}

	if err := s.db.Create(&familyMember).Error; err != nil {
		return nil, err
	}

	return &familyMember, nil
}

// ListFamilyMembers returns all family members of a member
func (s *MemberService) ListFamilyMembers(memberID uint) ([]models.FamilyMember, error) {
	var family []models.FamilyMember
	err := s.db.Where("member_id = ?", memberID).Order("created_at ASC").Find(&family).Error
	return family, err
}

// ResetCredentials generates a new password and sends it via SMS
func (s *MemberService) ResetCredentials(memberID uint) (*Credentials, error) {
	var member models.Member
	if err := s.db.Preload("User").First(&member, memberID).Error; err != nil {
		return nil, errors.New("member not found")
	}

	if member.User == nil {
		return nil, errors.New("member has no linked user account")
	}

	// Generate new password
	plainPassword := generateRandomPassword(8)
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	// Update user's password
	if err := s.db.Model(&models.User{}).Where("id = ?", member.User.ID).
		Update("password", hashedPassword).Error; err != nil {
		return nil, err
	}

	// Send new credentials via SMS
	go s.sendCredentialsSMS(member.User.Phone, plainPassword)

	return &Credentials{
		Phone:         member.User.Phone,
		PlainPassword: plainPassword,
	}, nil
}

// UploadMemberPhoto saves the member photo and returns the URL
func (s *MemberService) UploadMemberPhoto(memberID uint, file io.Reader, filename string) (string, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return "", errors.New("member not found")
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/members"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	uniqueName := fmt.Sprintf("member_%d_photo_%d%s", memberID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	// Use full URL with /api prefix since we proxy
	fileURL := fmt.Sprintf("/uploads/members/%s", uniqueName)

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

	// Update member with photo URL
	if err := s.db.Model(&member).Update("photo", fileURL).Error; err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to update member: %w", err)
	}

	// Fetch updated member to return
	var updatedMember models.Member
	s.db.First(&updatedMember, memberID)

	return fileURL, nil
}

// UploadAssistantPhoto saves the assistant photo and returns the URL
func (s *MemberService) UploadAssistantPhoto(memberID uint, file io.Reader, filename string) (string, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return "", errors.New("member not found")
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/members"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	uniqueName := fmt.Sprintf("member_%d_assistant_%d%s", memberID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	fileURL := fmt.Sprintf("/uploads/members/%s", uniqueName)

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

	// Update member with assistant photo URL
	if err := s.db.Model(&member).Update("assistant_photo", fileURL).Error; err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to update member: %w", err)
	}

	return fileURL, nil
}
