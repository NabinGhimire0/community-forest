package requests

import (
	"errors"
	"fmt"
	"time"

	"forest-management/internal/models"
	"forest-management/internal/notifications"
	"forest-management/pkg/response"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RequestService struct {
	db *gorm.DB
}

var (
	ErrNoActiveFiscalYear   = errors.New("no active fiscal year is configured")
	ErrMemberActiveYearOnly = errors.New(
		"members can submit resource requests only for the active fiscal year",
	)
)

func NewRequestService(db *gorm.DB) *RequestService {
	return &RequestService{db: db}
}

func (s *RequestService) CreateRequest(userID uint, input CreateRequestInput) (*models.Request, error) {
	// Get the user to check role
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("user not found")
	}

	var member models.Member

	// Determine which member is making the request
	if user.Role == "admin" || user.Role == "staff" {
		// Admin/Staff: must specify a member_id
		if input.MemberID == nil {
			return nil, errors.New("member_id is required when creating request as admin/staff")
		}
		if err := s.db.First(&member, *input.MemberID).Error; err != nil {
			return nil, errors.New("specified member not found")
		}
	} else {
		// Regular member: find by user_id
		if err := s.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
			return nil, errors.New("member profile not found. Please contact administrator")
		}
	}

	// Validate resource item
	var resourceItem models.ResourceItem
	if err := s.db.Preload("Type").First(&resourceItem, input.ResourceItemID).Error; err != nil {
		return nil, errors.New("invalid resource item")
	}

	// Members are never allowed to choose a previous or future fiscal year.
	// The backend enforces this rule even when a client tampers with the form.
	var fiscalYear models.FiscalYear
	if user.Role == "member" {
		if err := s.db.Where("is_active = ?", true).First(&fiscalYear).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNoActiveFiscalYear
			}
			return nil, fmt.Errorf("failed to load active fiscal year: %w", err)
		}
		if input.FiscalYearID != fiscalYear.ID {
			return nil, ErrMemberActiveYearOnly
		}
		input.FiscalYearID = fiscalYear.ID
	} else if err := s.db.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		return nil, errors.New("invalid fiscal year")
	}

	// Check stock availability
	var stock models.Stock
	err := s.db.Where(
		"resource_item_id = ? AND fiscal_year_id = ?",
		input.ResourceItemID, input.FiscalYearID,
	).First(&stock).Error

	if err != nil {
		return nil, errors.New("no stock found for this resource in the selected fiscal year")
	}
	availableQuantity := stock.RemainingQuantity - stock.ReservedQuantity
	if availableQuantity < input.QuantityRequested {
		return nil, fmt.Errorf("insufficient available stock. Available: %.2f %s, Requested: %.2f %s",
			availableQuantity, resourceItem.Type.Unit, input.QuantityRequested, resourceItem.Type.Unit)
	}

	// Create request
	req := models.Request{
		MemberID:          member.ID,
		FiscalYearID:      input.FiscalYearID,
		ResourceItemID:    input.ResourceItemID,
		QuantityRequested: input.QuantityRequested,
		Status:            "pending",
		RequestedAt:       time.Now(),
		Remarks:           input.Remarks,
	}

	if err := s.db.Create(&req).Error; err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Preload relations for response
	s.db.Preload("Member").Preload("ResourceItem.Type").Preload("FiscalYear").First(&req, req.ID)

	// Notify admins about new request
	notifService := notifications.NewNotificationService(s.db)
	notifService.NotifyRole(
		"admin",
		"New Resource Request",
		fmt.Sprintf("Member %s requested %.2f %s of %s",
			member.Name, input.QuantityRequested, resourceItem.Type.Unit, resourceItem.Name),
		"request",
		stringPtr("request"),
		&req.ID,
	)

	return &req, nil
}

func (s *RequestService) ListRequests(page, perPage int, status, fiscalYearID, memberID, search string) ([]models.Request, *response.Pagination, error) {
	var requests []models.Request
	var total int64

	query := s.db.Model(&models.Request{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	if memberID != "" {
		query = query.Where("member_id = ?", memberID)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("JOIN members ON members.id = requests.member_id").
			Where("members.name ILIKE ? OR members.membership_no ILIKE ?", searchPattern, searchPattern)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("Member").
		Preload("ResourceItem.Type").
		Preload("FiscalYear").
		Preload("Approver").
		Preload("Payments").
		Order("requested_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&requests).Error

	meta := &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}

	return requests, meta, err
}

func (s *RequestService) GetRequestByID(id, userID uint, role string) (*models.Request, error) {
	var req models.Request
	query := s.db.Model(&models.Request{})

	// Members may only view requests that belong to their own member profile.
	if role == "member" {
		query = query.
			Joins("JOIN members ON members.id = requests.member_id").
			Where("members.user_id = ?", userID)
	}

	err := query.
		Preload("Member").
		Preload("ResourceItem.Type").
		Preload("FiscalYear").
		Preload("Approver").
		Preload("Payments").
		First(&req, id).Error
	return &req, err
}

func (s *RequestService) GetMemberRequests(userID uint, page, perPage int, status string) ([]models.Request, *response.Pagination, error) {
	var member models.Member
	if err := s.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, nil, errors.New("member not found")
	}

	var requests []models.Request
	var total int64

	query := s.db.Model(&models.Request{}).Where("member_id = ?", member.ID)
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
		Preload("ResourceItem.Type").
		Preload("FiscalYear").
		Preload("Payments").
		Order("requested_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&requests).Error

	meta := &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}

	return requests, meta, err
}

func (s *RequestService) ApproveRequest(requestID, adminUserID uint, input ApproveRequestInput) (*models.Request, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var req models.Request
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("ResourceItem.Type").First(&req, requestID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("request not found")
	}
	if req.Status != "pending" {
		tx.Rollback()
		return nil, errors.New("only pending requests can be approved")
	}

	quantityApproved := req.QuantityRequested
	if input.QuantityApproved != nil && *input.QuantityApproved > 0 {
		if *input.QuantityApproved > req.QuantityRequested {
			tx.Rollback()
			return nil, errors.New("approved quantity cannot exceed requested quantity")
		}
		quantityApproved = *input.QuantityApproved
	}

	var rate models.ResourceRate
	if err := tx.Where("resource_item_id = ? AND fiscal_year_id = ?", req.ResourceItemID, req.FiscalYearID).First(&rate).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("resource rate not configured for this fiscal year")
	}

	// Reserve stock during approval. This prevents two approved requests from
	// selling the same stock while the members are still paying.
	result := tx.Model(&models.Stock{}).
		Where("resource_item_id = ? AND fiscal_year_id = ? AND remaining_quantity - reserved_quantity >= ?", req.ResourceItemID, req.FiscalYearID, quantityApproved).
		UpdateColumn("reserved_quantity", gorm.Expr("reserved_quantity + ?", quantityApproved))
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, errors.New("insufficient available stock to approve this request")
	}

	totalAmount := quantityApproved * rate.RatePerUnit
	now := time.Now()
	updates := map[string]interface{}{
		"status": "approved", "quantity_approved": quantityApproved,
		"rate_per_unit": rate.RatePerUnit, "total_amount": totalAmount,
		"approved_by": adminUserID, "approved_at": &now,
	}
	if input.Remarks != nil {
		updates["remarks"] = *input.Remarks
	}
	if err := tx.Model(&req).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to approve request: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.db.Preload("Member").Preload("Member.User").Preload("ResourceItem.Type").Preload("FiscalYear").Preload("Approver").Preload("Payments").First(&req, requestID)
	if req.Member != nil && req.Member.UserID != nil {
		notifications.NewNotificationService(s.db).NotifyUser(
			*req.Member.UserID, "Request Approved",
			fmt.Sprintf("Your request for %.2f %s of %s has been approved. Total Amount: Rs. %.2f", quantityApproved, req.ResourceItem.Type.Unit, req.ResourceItem.Name, totalAmount),
			"success", stringPtr("request"), &req.ID,
		)
	}
	return &req, nil
}

func (s *RequestService) RejectRequest(requestID, adminUserID uint, remarks *string) (*models.Request, error) {
	var req models.Request

	if err := s.db.First(&req, requestID).Error; err != nil {
		return nil, errors.New("request not found")
	}

	if req.Status != "pending" {
		return nil, errors.New("only pending requests can be rejected")
	}

	now := time.Now()

	updates := map[string]interface{}{
		"status":      "rejected",
		"approved_by": adminUserID,
		"approved_at": &now,
	}

	if remarks != nil {
		updates["remarks"] = *remarks
	}

	if err := s.db.Model(&req).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to reject request: %w", err)
	}

	s.db.
		Preload("Member").
		Preload("ResourceItem.Type").
		Preload("FiscalYear").
		Preload("Approver").
		First(&req, requestID)

	notifService := notifications.NewNotificationService(s.db)

	var member models.Member
	s.db.Preload("User").First(&member, req.MemberID)

	if member.UserID != nil {
		notifService.NotifyUser(
			*member.UserID,
			"Request Rejected",
			fmt.Sprintf("Your request for %s has been rejected.", req.ResourceItem.Name),
			"warning",
			stringPtr("request"),
			&req.ID,
		)
	}

	return &req, nil
}

func (s *RequestService) GetRequestStatistics(fiscalYearID string) (map[string]interface{}, error) {
	var stats struct {
		Total     int64 `json:"total"`
		Pending   int64 `json:"pending"`
		Approved  int64 `json:"approved"`
		Rejected  int64 `json:"rejected"`
		Completed int64 `json:"completed"`
	}

	query := s.db.Model(&models.Request{})
	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}

	query.Count(&stats.Total)
	query.Where("status = ?", "pending").Count(&stats.Pending)
	query.Where("status = ?", "approved").Count(&stats.Approved)
	query.Where("status = ?", "rejected").Count(&stats.Rejected)
	query.Where("status = ?", "completed").Count(&stats.Completed)

	return map[string]interface{}{
		"total":     stats.Total,
		"pending":   stats.Pending,
		"approved":  stats.Approved,
		"rejected":  stats.Rejected,
		"completed": stats.Completed,
	}, nil
}

func stringPtr(s string) *string {
	return &s
}
