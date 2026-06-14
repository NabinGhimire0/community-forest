package members

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"forest-management/internal/audit"
	"forest-management/internal/membershipfees"
	"forest-management/internal/models"
	"forest-management/pkg/fileutil"
	"forest-management/pkg/response"
	"forest-management/pkg/security"
	"forest-management/pkg/utils"

	"gorm.io/gorm"
)

type MemberService struct {
	db *gorm.DB
}

var ErrMemberUpdateForbidden = errors.New("member update is not allowed for this account")

func NewMemberService(db *gorm.DB) *MemberService {
	return &MemberService{db: db}
}

// Credentials holds the plain-text credentials (only shown once to admin)
type Credentials struct {
	Phone         string
	PlainPassword string
}

// CreateMember creates the official member record and a one-time login credential.
// Only an administrator may call the route because the temporary password is
// displayed exactly once and must be handed to the member through a trusted channel.
func (s *MemberService) CreateMember(req CreateMemberRequest, actorUserID *uint) (*models.Member, *Credentials, error) {
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

	// 2. Generate a strong temporary password.
	plainPassword, err := security.GenerateStrongPassword(16)
	if err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to generate temporary password: %w", err)
	}

	// 3. Hash the password
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Normalize and validate the phone number used as the login identity.
	if req.Phone == nil {
		tx.Rollback()
		return nil, nil, errors.New("phone number is required")
	}
	phoneStr, err := security.NormalizeNepalMobile(*req.Phone)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	req.Phone = &phoneStr
	user := models.User{
		Name:               req.Name,
		Phone:              phoneStr,
		Password:           hashedPassword,
		Role:               "member",
		Status:             "active",
		MustChangePassword: true,
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

	// New members must receive the active fiscal year's Gasti/Membership fee
	// exactly once when a fee setting is available.
	if _, err := membershipfees.AssignActiveYearForMember(tx, member); err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf("failed to assign active fiscal-year membership fee: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	audit.CreateAuditEntry(
		s.db,
		actorUserID,
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

	actorIsAdmin := false
	if reqUserID != nil {
		var actor models.User
		if err := s.db.Select("id", "role").First(&actor, *reqUserID).Error; err != nil {
			return nil, ErrMemberUpdateForbidden
		}
		actorIsAdmin = actor.Role == "admin"
	}
	var linkedUser *models.User
	if member.UserID != nil {
		var target models.User
		if err := s.db.Select("id", "phone", "role", "status").First(&target, *member.UserID).Error; err == nil {
			linkedUser = &target
		}
	}
	if !actorIsAdmin && linkedUser != nil && (linkedUser.Role == "admin" || linkedUser.Role == "staff") {
		return nil, fmt.Errorf("%w: staff cannot edit a privileged account through the member register", ErrMemberUpdateForbidden)
	}
	if !actorIsAdmin && req.Status != nil && *req.Status != member.Status {
		return nil, fmt.Errorf("%w: only an administrator can change member status", ErrMemberUpdateForbidden)
	}
	if !actorIsAdmin && req.Phone != nil {
		normalizedPhone, err := security.NormalizeNepalMobile(*req.Phone)
		if err != nil {
			return nil, err
		}
		currentPhone := ""
		if member.Phone != nil {
			currentPhone, _ = security.NormalizeNepalMobile(*member.Phone)
		}
		if normalizedPhone != currentPhone {
			return nil, fmt.Errorf("%w: only an administrator can change a login phone number", ErrMemberUpdateForbidden)
		}
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
		normalizedPhone, err := security.NormalizeNepalMobile(*req.Phone)
		if err != nil {
			return nil, err
		}
		req.Phone = &normalizedPhone
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

	// Keep the linked login identity in sync with the official member record.
	// Status changes must immediately affect authentication as well.
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin member update: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	if member.UserID != nil {
		userUpdates := map[string]interface{}{
			"name":   req.Name,
			"status": member.Status,
		}
		if req.Phone != nil {
			userUpdates["phone"] = *req.Phone
		}
		if err := tx.Model(&models.User{}).Where("id = ?", *member.UserID).Updates(userUpdates).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update linked user: %w", err)
		}
		if member.Status != "active" {
			now := time.Now()
			if err := tx.Model(&models.UserSession{}).
				Where("user_id = ? AND revoked_at IS NULL", *member.UserID).
				Update("revoked_at", now).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to revoke member sessions: %w", err)
			}
		}
	}

	if err := tx.Save(&member).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update member: %w", err)
	}

	// Update family members in the same transaction so the member record cannot
	// be partially updated when a family-row operation fails.
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

		association := tx.Model(&member).Association("FamilyMembers")
		if len(familyMembers) > 0 {
			if err := association.Replace(familyMembers); err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update family members: %w", err)
			}
		} else if err := association.Clear(); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to clear family members: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit member update: %w", err)
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

// DeleteMember deactivates a member and their login while preserving the
// official register, financial ledger, documents, and audit history. Permanent
// deletion is intentionally not supported through the API.
func (s *MemberService) DeleteMember(id uint, actorUserID *uint) error {
	var member models.Member
	if err := s.db.First(&member, id).Error; err != nil {
		return errors.New("member not found")
	}
	if member.Status == "inactive" {
		return errors.New("member is already inactive")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	member.Status = "inactive"
	if err := tx.Save(&member).Error; err != nil {
		tx.Rollback()
		return err
	}
	if member.UserID != nil {
		if err := tx.Model(&models.User{}).Where("id = ?", *member.UserID).
			Update("status", "inactive").Error; err != nil {
			tx.Rollback()
			return err
		}
		now := time.Now()
		if err := tx.Model(&models.UserSession{}).
			Where("user_id = ? AND revoked_at IS NULL", *member.UserID).
			Update("revoked_at", now).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return err
	}

	audit.CreateAuditEntry(
		s.db,
		actorUserID,
		"deactivate",
		"member",
		&member.ID,
		map[string]interface{}{"status": "active"},
		map[string]interface{}{"status": "inactive"},
		"",
		"",
		"Member and linked login deactivated; historical records preserved",
	)
	return nil
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

// UpdateFamilyMember updates a family member while enforcing parent-member ownership.
func (s *MemberService) UpdateFamilyMember(memberID, familyID uint, req FamilyMemberRequest) (*models.FamilyMember, error) {
	var familyMember models.FamilyMember
	if err := s.db.Where("id = ? AND member_id = ?", familyID, memberID).First(&familyMember).Error; err != nil {
		return nil, errors.New("family member not found")
	}

	familyMember.Name = req.Name
	familyMember.Relation = req.Relation
	familyMember.Age = req.Age
	familyMember.Gender = req.Gender
	familyMember.CitizenshipNo = req.CitizenshipNo
	familyMember.Remarks = req.Remarks

	if err := s.db.Save(&familyMember).Error; err != nil {
		return nil, fmt.Errorf("failed to update family member: %w", err)
	}

	return &familyMember, nil
}

// DeleteFamilyMember deletes a family member only from the specified parent member.
func (s *MemberService) DeleteFamilyMember(memberID, familyID uint) error {
	result := s.db.Where("id = ? AND member_id = ?", familyID, memberID).Delete(&models.FamilyMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("family member not found")
	}
	return nil
}

// ResetCredentials generates a one-time temporary password and revokes all existing sessions
func (s *MemberService) ResetCredentials(memberID uint) (*Credentials, error) {
	var member models.Member
	if err := s.db.Preload("User").First(&member, memberID).Error; err != nil {
		return nil, errors.New("member not found")
	}

	if member.User == nil {
		return nil, errors.New("member has no linked user account")
	}

	// Generate a strong one-time temporary password.
	plainPassword, err := security.GenerateStrongPassword(16)
	if err != nil {
		return nil, err
	}
	hashedPassword, err := utils.HashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	// Update the password, force a change at next login, and revoke old sessions.
	now := time.Now().UTC()
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.User{}).Where("id = ?", member.User.ID).Updates(map[string]interface{}{
			"password": hashedPassword, "must_change_password": true,
			"password_changed_at": now, "failed_login_attempts": 0, "locked_until": nil,
		}).Error; err != nil {
			return err
		}
		return tx.Model(&models.UserSession{}).Where("user_id = ? AND revoked_at IS NULL", member.User.ID).Update("revoked_at", now).Error
	}); err != nil {
		return nil, err
	}

	return &Credentials{
		Phone:         member.User.Phone,
		PlainPassword: plainPassword,
	}, nil
}

// UploadMemberPhoto securely stores a member image outside the public web root.
func (s *MemberService) UploadMemberPhoto(memberID uint, file io.Reader, filename string) (string, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return "", errors.New("member not found")
	}
	saved, err := fileutil.Save(file, "members", fmt.Sprintf("member-%d-photo", memberID), fileutil.ImagePolicy)
	if err != nil {
		return "", err
	}
	if err := s.db.Model(&member).Update("photo", saved.URL).Error; err != nil {
		_ = os.Remove(saved.Path)
		return "", fmt.Errorf("failed to update member: %w", err)
	}
	return saved.URL, nil
}

// UploadAssistantPhoto securely stores an assistant-household-head image.
func (s *MemberService) UploadAssistantPhoto(memberID uint, file io.Reader, filename string) (string, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return "", errors.New("member not found")
	}
	saved, err := fileutil.Save(file, "members", fmt.Sprintf("member-%d-assistant", memberID), fileutil.ImagePolicy)
	if err != nil {
		return "", err
	}
	if err := s.db.Model(&member).Update("assistant_photo", saved.URL).Error; err != nil {
		_ = os.Remove(saved.Path)
		return "", fmt.Errorf("failed to update member: %w", err)
	}
	return saved.URL, nil
}

// GetMemberFinancialSummary returns a single, non-duplicated ledger summary.
// Draft and reversed legacy entries remain visible in detail screens but do not
// affect financial totals until they are verified.
func (s *MemberService) GetMemberFinancialSummary(memberID uint) (map[string]interface{}, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return nil, errors.New("member not found")
	}

	var transactions []models.Transaction
	if err := s.db.Preload("FiscalYear").Preload("ResourceItem.Type").
		Where("member_id = ? AND (record_status = ? OR record_status = '')", memberID, "verified").
		Order("date DESC, id DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}

	var timberPaid, firewoodPaid, otherSalesPaid, feePaid, finePaid float64
	var totalPaid, totalDue, historicalOutstanding, historicalCollected float64
	for _, txn := range transactions {
		totalPaid += txn.AmountPaid
		totalDue += txn.AmountRemaining
		if strings.HasPrefix(txn.Type, "legacy_") {
			historicalOutstanding += txn.AmountRemaining
			historicalCollected += txn.AmountPaid
		}
		switch txn.Type {
		case "membership_fee", historicalFeeType:
			feePaid += txn.AmountPaid
		case "fine":
			finePaid += txn.AmountPaid
		case historicalTimberSaleType:
			timberPaid += txn.AmountPaid
		case historicalFirewoodSaleType:
			firewoodPaid += txn.AmountPaid
		case historicalOtherSaleType:
			otherSalesPaid += txn.AmountPaid
		default:
			if txn.Type == "resource_sale" && txn.ResourceItem != nil && txn.ResourceItem.Type != nil {
				switch strings.ToLower(txn.ResourceItem.Type.Name) {
				case "timber", "काठ":
					timberPaid += txn.AmountPaid
				case "firewood", "दाउरा":
					firewoodPaid += txn.AmountPaid
				default:
					otherSalesPaid += txn.AmountPaid
				}
			}
		}
	}

	return map[string]interface{}{
		"member_name": member.Name, "membership_no": member.MembershipNo,
		"total_timber_sales": timberPaid, "total_firewood_sales": firewoodPaid,
		"total_other_sales": otherSalesPaid, "total_membership_fees": feePaid,
		"total_fines": finePaid, "total_paid": totalPaid, "total_due": totalDue,
		"historical_outstanding": historicalOutstanding,
		"historical_collected":   historicalCollected,
		"transactions":           transactions,
	}, nil
}

func (s *MemberService) attachTransactionDocuments(rows []models.Transaction) {
	for i := range rows {
		var docs []models.FileUpload
		if s.db.Where("entity = ? AND entity_id = ?", "transaction", rows[i].ID).
			Order("created_at DESC").Find(&docs).Error == nil {
			rows[i].Documents = docs
		}
	}
}

// GetMemberFeeDetails returns current membership fees and historical Gasti fee balances.
func (s *MemberService) GetMemberFeeDetails(memberID uint) (map[string]interface{}, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return nil, errors.New("member not found")
	}

	var rows []models.Transaction
	if err := s.db.Preload("FiscalYear").Preload("EnteredByUser").Preload("VerifiedByUser").
		Where("member_id = ? AND type IN ?", memberID, []string{"membership_fee", historicalFeeType}).
		Order("date DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	s.attachTransactionDocuments(rows)

	var totalAmount, totalPaid, totalRemaining float64
	for _, txn := range rows {
		if txn.RecordStatus == "verified" || txn.RecordStatus == "" {
			totalAmount += txn.TotalAmount
			totalPaid += txn.AmountPaid
			totalRemaining += txn.AmountRemaining
		}
	}
	return map[string]interface{}{
		"member_name": member.Name, "membership_no": member.MembershipNo,
		"transactions": rows, "total_amount": totalAmount,
		"total_paid": totalPaid, "total_remaining": totalRemaining,
	}, nil
}

// GetMemberSalesDetails returns current resource sales and historical sales balances.
func (s *MemberService) GetMemberSalesDetails(memberID uint) (map[string]interface{}, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return nil, errors.New("member not found")
	}

	var rows []models.Transaction
	if err := s.db.Preload("FiscalYear").Preload("ResourceItem.Type").Preload("EnteredByUser").Preload("VerifiedByUser").
		Where("member_id = ? AND type IN ?", memberID, []string{"resource_sale", historicalTimberSaleType, historicalFirewoodSaleType, historicalOtherSaleType}).
		Order("date DESC, id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	s.attachTransactionDocuments(rows)

	var timberTotal, firewoodTotal, otherTotal, received, remaining float64
	list := make([]map[string]interface{}, 0, len(rows))
	for _, txn := range rows {
		category, itemName := "अन्य / Other", "-"
		switch txn.Type {
		case historicalTimberSaleType:
			category = "काठ / Timber"
		case historicalFirewoodSaleType:
			category = "दाउरा / Firewood"
		case historicalOtherSaleType:
			category = "अन्य / Other"
		default:
			if txn.ResourceItem != nil {
				itemName = txn.ResourceItem.Name
			}
			if txn.ResourceItem != nil && txn.ResourceItem.Type != nil {
				switch strings.ToLower(txn.ResourceItem.Type.Name) {
				case "timber", "काठ":
					category = "काठ / Timber"
				case "firewood", "दाउरा":
					category = "दाउरा / Firewood"
				default:
					category = txn.ResourceItem.Type.Name
				}
			}
		}

		countInTotals := txn.RecordStatus == "verified" || txn.RecordStatus == ""
		if countInTotals {
			switch category {
			case "काठ / Timber":
				timberTotal += txn.TotalAmount
			case "दाउरा / Firewood":
				firewoodTotal += txn.TotalAmount
			default:
				otherTotal += txn.TotalAmount
			}
			received += txn.AmountPaid
			remaining += txn.AmountRemaining
		}
		fiscalYearName := "-"
		if txn.FiscalYear != nil {
			fiscalYearName = txn.FiscalYear.Name
		}
		list = append(list, map[string]interface{}{
			"id": txn.ID, "type": txn.Type, "is_legacy": strings.HasPrefix(txn.Type, "legacy_"),
			"source": txn.Source, "record_status": txn.RecordStatus,
			"date": txn.Date, "receipt_no": txn.ReceiptNo,
			"physical_reference": txn.PhysicalReference,
			"description":        category, "item_name": itemName,
			"quantity": txn.Quantity, "rate": txn.RatePerUnit,
			"total_amount": txn.TotalAmount, "paid_amount": txn.AmountPaid,
			"amount_paid": txn.AmountPaid, "remaining": txn.AmountRemaining,
			"amount_remaining": txn.AmountRemaining, "fiscal_year": fiscalYearName,
			"fiscal_year_object": txn.FiscalYear, "remarks": txn.Remarks,
			"documents": txn.Documents, "entered_by_user": txn.EnteredByUser,
			"verified_by_user": txn.VerifiedByUser, "verified_at": txn.VerifiedAt,
			"reversal_reason": txn.ReversalReason,
		})
	}
	return map[string]interface{}{
		"member_name": member.Name, "membership_no": member.MembershipNo,
		"transactions": list, "timber_total": timberTotal, "firewood_total": firewoodTotal,
		"other_total": otherTotal, "total_received": received, "total_remaining": remaining,
	}, nil
}
