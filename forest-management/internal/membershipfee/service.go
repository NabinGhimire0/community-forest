package membershipfee

import (
	"errors"
	"fmt"
	"time"

	"forest-management/internal/models"
	"forest-management/internal/notifications"

	"gorm.io/gorm"
)

type MembershipFeeService struct {
	db    *gorm.DB
	notif *notifications.NotificationService
}

func NewMembershipFeeService(db *gorm.DB) *MembershipFeeService {
	return &MembershipFeeService{
		db:    db,
		notif: notifications.NewNotificationService(db),
	}
}

type CollectFeeInput struct {
	MemberID       uint    `json:"member_id" binding:"required"`
	FiscalYearID   uint    `json:"fiscal_year_id" binding:"required"`
	AmountPaid     float64 `json:"amount_paid" binding:"required"`
	PaymentMethod  string  `json:"payment_method" binding:"required"` // cash, bank, online
	TransactionRef *string `json:"transaction_ref"`
	Remarks        *string `json:"remarks"`
}

type CollectFeeResponse struct {
	Transaction models.Transaction `json:"transaction"`
	ReceiptNo   string             `json:"receipt_no"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// CollectFee records a membership fee payment and creates a transaction
//
// FLOW:
// 1. Validate member exists and is active
// 2. Get the fee setting for the fiscal year
// 3. Check if member already paid for this fiscal year
// 4. Create Transaction (type: membership_fee)
// 5. Generate receipt number
// 6. Send notification to member
func (s *MembershipFeeService) CollectFee(adminUserID uint, input CollectFeeInput) (*CollectFeeResponse, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Validate member
	var member models.Member
	if err := tx.Preload("User").First(&member, input.MemberID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("member not found")
	}
	if member.Status != "active" {
		tx.Rollback()
		return nil, errors.New("member is not active")
	}

	// 2. Get fee setting
	var feeSetting models.FeeSetting
	if err := tx.Where("fiscal_year_id = ?", input.FiscalYearID).First(&feeSetting).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("membership fee not configured for this fiscal year")
	}

	// 3. Check if already paid
	var existingCount int64
	tx.Model(&models.Transaction{}).
		Where("member_id = ? AND fiscal_year_id = ? AND type = ?", input.MemberID, input.FiscalYearID, "membership_fee").
		Count(&existingCount)
	if existingCount > 0 {
		tx.Rollback()
		return nil, errors.New("member has already paid membership fee for this fiscal year")
	}

	// 4. Calculate amounts
	totalAmount := feeSetting.MembershipFee
	amountRemaining := totalAmount - input.AmountPaid
	if amountRemaining < 0 {
		amountRemaining = 0
	}

	// 5. Generate receipt number
	now := time.Now()
	var count int64
	tx.Model(&models.Transaction{}).Where("type = ?", "membership_fee").Count(&count)
	receiptNo := fmt.Sprintf("MEM-REC-%d-%04d", now.Year(), count+1)

	// 6. Create transaction
	transaction := models.Transaction{
		MemberID:        input.MemberID,
		FiscalYearID:    input.FiscalYearID,
		Type:            "membership_fee",
		TotalAmount:     totalAmount,
		AmountPaid:      input.AmountPaid,
		AmountRemaining: amountRemaining,
		ReceiptNo:       receiptNo,
		Date:            now,
		Remarks:         input.Remarks,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload with relations
	s.db.Preload("Member").Preload("FiscalYear").First(&transaction, transaction.ID)

	// 7. Send notification to member (async)
	if member.UserID != nil {
		go s.notif.NotifyUser(
			*member.UserID,
			"Membership Fee Received",
			fmt.Sprintf("Your membership fee of Rs. %.2f for this fiscal year has been received. Receipt: %s", input.AmountPaid, receiptNo),
			"payment",
			stringPtr("transaction"),
			&transaction.ID,
		)
	}

	return &CollectFeeResponse{
		Transaction: transaction,
		ReceiptNo:   receiptNo,
	}, nil
}

// BulkCollectFee collects membership fees for all active members who haven't paid
func (s *MembershipFeeService) BulkCollectFee(adminUserID uint, fiscalYearID uint, paymentMethod string) ([]CollectFeeResponse, error) {
	// Get fee setting
	var feeSetting models.FeeSetting
	if err := s.db.Where("fiscal_year_id = ?", fiscalYearID).First(&feeSetting).Error; err != nil {
		return nil, errors.New("membership fee not configured for this fiscal year")
	}

	// Get all active members
	var members []models.Member
	s.db.Where("status = ?", "active").Find(&members)

	var results []CollectFeeResponse
	var errors_list []string

	for _, member := range members {
		// Check if already paid
		var existingCount int64
		s.db.Model(&models.Transaction{}).
			Where("member_id = ? AND fiscal_year_id = ? AND type = ?", member.ID, fiscalYearID, "membership_fee").
			Count(&existingCount)

		if existingCount > 0 {
			continue // Skip — already paid
		}

		input := CollectFeeInput{
			MemberID:      member.ID,
			FiscalYearID:  fiscalYearID,
			AmountPaid:    feeSetting.MembershipFee,
			PaymentMethod: paymentMethod,
		}

		result, err := s.CollectFee(adminUserID, input)
		if err != nil {
			errors_list = append(errors_list, fmt.Sprintf("Member %d: %s", member.ID, err.Error()))
			continue
		}

		results = append(results, *result)
	}

	if len(errors_list) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all fee collections failed: %v", errors_list)
	}

	return results, nil
}

// GetMemberFeeStatus returns fee payment status for a member across fiscal years
func (s *MembershipFeeService) GetMemberFeeStatus(memberID uint) ([]map[string]interface{}, error) {
	var fiscalYears []models.FiscalYear
	s.db.Order("start_date DESC").Find(&fiscalYears)

	var result []map[string]interface{}
	for _, fy := range fiscalYears {
		var feeSetting models.FeeSetting
		feeErr := s.db.Where("fiscal_year_id = ?", fy.ID).First(&feeSetting).Error

		var transaction models.Transaction
		txnErr := s.db.Where("member_id = ? AND fiscal_year_id = ? AND type = ?", memberID, fy.ID, "membership_fee").First(&transaction).Error

		entry := map[string]interface{}{
			"fiscal_year_id":   fy.ID,
			"fiscal_year_name": fy.Name,
			"is_active":        fy.IsActive,
		}

		if feeErr != nil {
			entry["fee_configured"] = false
		} else {
			entry["fee_configured"] = true
			entry["membership_fee"] = feeSetting.MembershipFee
		}

		if txnErr != nil {
			entry["paid"] = false
		} else {
			entry["paid"] = true
			entry["amount_paid"] = transaction.AmountPaid
			entry["amount_remaining"] = transaction.AmountRemaining
			entry["receipt_no"] = transaction.ReceiptNo
			entry["paid_date"] = transaction.Date
		}

		result = append(result, entry)
	}

	return result, nil
}

// ListFeeCollections returns all membership fee transactions with filters
func (s *MembershipFeeService) ListFeeCollections(page, perPage int, fiscalYearID, memberID string) ([]models.Transaction, *PaginationMeta, error) {
	var txns []models.Transaction
	var total int64

	query := s.db.Model(&models.Transaction{}).Where("type = ?", "membership_fee")

	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	if memberID != "" {
		query = query.Where("member_id = ?", memberID)
	}

	query.Count(&total)
	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("Member").
		Preload("FiscalYear").
		Order("date DESC").
		Offset(offset).
		Limit(perPage).
		Find(&txns).Error

	return txns, &PaginationMeta{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

func stringPtr(s string) *string {
	return &s
}
