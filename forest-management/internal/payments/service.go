package payments

import (
	"errors"
	"fmt"
	"time"

	"forest-management/internal/models"
	"forest-management/internal/notifications"
	"forest-management/pkg/response"

	"gorm.io/gorm"
)

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db}
}

func (s *PaymentService) CreatePayment(userID uint, input CreatePaymentInput) (*models.Payment, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get the user to check role
	var user models.User
	if err := tx.First(&user, userID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("user not found")
	}

	var member models.Member

	// Determine which member is making the payment
	if user.Role == "admin" || user.Role == "staff" {
		// Admin/Staff: must specify a member_id
		if input.MemberID == nil {
			tx.Rollback()
			return nil, errors.New("member_id is required when recording payment as admin/staff")
		}
		if err := tx.Preload("User").First(&member, *input.MemberID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("specified member not found")
		}
	} else {
		// Regular member: find by user_id
		if err := tx.Preload("User").Where("user_id = ?", userID).First(&member).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("member profile not found. Please contact administrator")
		}
	}

	var request *models.Request
	var totalAmount float64

	// Validate linked request if provided
	if input.RequestID != nil && *input.RequestID != 0 {
		var req models.Request
		if err := tx.Preload("ResourceItem.Type").First(&req, *input.RequestID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("request not found")
		}
		if req.MemberID != member.ID {
			tx.Rollback()
			return nil, errors.New("this request does not belong to the selected member")
		}
		if req.Status != "approved" {
			tx.Rollback()
			return nil, errors.New("only approved requests can be paid for")
		}
		if req.TotalAmount == nil {
			tx.Rollback()
			return nil, errors.New("request total amount not calculated")
		}
		request = &req
		totalAmount = *req.TotalAmount
	}

	// Create payment
	now := time.Now()
	payment := models.Payment{
		MemberID:      member.ID,
		RequestID:     input.RequestID,
		Amount:        input.Amount,
		PaymentMethod: input.PaymentMethod,
		TransactionID: input.TransactionID,
		Status:        "pending",
	}

	// Online payments are auto-verified
	if input.PaymentMethod == "esewa" || input.PaymentMethod == "khalti" {
		payment.Status = "paid"
		payment.PaidAt = &now
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// If payment is successful AND linked to a request, create transaction and update stock
	if payment.Status == "paid" && request != nil {
		// Check if total paid equals or exceeds total amount
		var totalPaid float64
		tx.Model(&models.Payment{}).
			Where("request_id = ? AND status = ?", request.ID, "paid").
			Select("COALESCE(SUM(amount), 0)").Scan(&totalPaid)

		newTotalPaid := totalPaid + input.Amount
		amountRemaining := totalAmount - newTotalPaid
		if amountRemaining < 0 {
			amountRemaining = 0
		}

		// Generate receipt number
		receiptNo := fmt.Sprintf("REC-%d-%d", now.Year(), payment.ID)

		// Create Transaction (ledger entry)
		transaction := models.Transaction{
			MemberID:        member.ID,
			FiscalYearID:    request.FiscalYearID,
			ResourceItemID:  &request.ResourceItemID,
			Type:            "resource_sale",
			Quantity:        request.QuantityApproved,
			RatePerUnit:     request.RatePerUnit,
			TotalAmount:     totalAmount,
			AmountPaid:      newTotalPaid,
			AmountRemaining: amountRemaining,
			ReceiptNo:       receiptNo,
			Date:            now,
		}

		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create transaction: %w", err)
		}

		// Mark request as completed if fully paid
		if amountRemaining <= 0 {
			if err := tx.Model(request).Update("status", "completed").Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update request status: %w", err)
			}
		}

		// Deduct from stock (only once per request)
		// Use a different approach to check if transaction exists for this request
		// Since transactions don't have a request_id field, we check by the receipt pattern or just deduct once
		// For now, we'll deduct if this is the first payment that makes it paid
		if newTotalPaid == input.Amount && totalPaid == 0 {
			if err := tx.Model(&models.Stock{}).
				Where("resource_item_id = ? AND fiscal_year_id = ?",
					request.ResourceItemID, request.FiscalYearID).
				UpdateColumn("remaining_quantity", gorm.Expr("remaining_quantity - ?", *request.QuantityApproved)).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update stock: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Reload payment with relations
	if err := s.db.Preload("Member").Preload("Member.User").Preload("Request.ResourceItem.Type").First(&payment, payment.ID).Error; err != nil {
		return nil, fmt.Errorf("failed to load payment: %w", err)
	}

	// Send notification (only if member has a user)
	if member.User != nil && member.UserID != nil {
		notifService := notifications.NewNotificationService(s.db)
		notifService.NotifyUser(
			*member.UserID,
			"Payment Successful",
			fmt.Sprintf("Payment of Rs. %.2f has been received.", input.Amount),
			"payment",
			stringPtr("payment"),
			&payment.ID,
		)
	}

	// Notify admins for online payments
	if input.PaymentMethod != "cash" && payment.Status == "paid" {
		notifService := notifications.NewNotificationService(s.db)
		notifService.NotifyRole(
			"admin",
			"New Online Payment",
			fmt.Sprintf("Payment of Rs. %.2f received via %s from member %s",
				input.Amount, input.PaymentMethod, member.Name),
			"payment",
			stringPtr("payment"),
			&payment.ID,
		)
	}

	return &payment, nil
}

func (s *PaymentService) ListPayments(page, perPage int, fiscalYearID, status, memberID string) ([]models.Payment, *response.Pagination, error) {
	var payments []models.Payment
	var total int64

	query := s.db.Model(&models.Payment{})

	if fiscalYearID != "" {
		query = query.Joins("JOIN requests ON requests.id = payments.request_id").
			Where("requests.fiscal_year_id = ?", fiscalYearID)
	}
	if status != "" {
		query = query.Where("payments.status = ?", status)
	}
	if memberID != "" {
		query = query.Where("payments.member_id = ?", memberID)
	}

	query.Count(&total)

	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("Member").
		Preload("Member.User").
		Preload("Request.ResourceItem.Type").
		Order("created_at DESC").
		Offset(offset).
		Limit(perPage).
		Find(&payments).Error

	return payments, &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

func (s *PaymentService) GetPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := s.db.Preload("Member").Preload("Member.User").Preload("Request.ResourceItem.Type").First(&payment, id).Error
	return &payment, err
}

func (s *PaymentService) GetMemberPayments(userID uint, page, perPage int) ([]models.Payment, *response.Pagination, error) {
	var member models.Member
	if err := s.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, nil, errors.New("member not found")
	}

	return s.ListPayments(page, perPage, "", "", fmt.Sprintf("%d", member.ID))
}

func (s *PaymentService) VerifyPayment(paymentID, adminUserID uint, status string) (*models.Payment, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var payment models.Payment
	if err := tx.Preload("Member").Preload("Member.User").Preload("Request").First(&payment, paymentID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("payment not found")
	}

	if payment.Status != "pending" {
		tx.Rollback()
		return nil, errors.New("payment already processed")
	}

	now := time.Now()

	// Update payment
	payment.Status = status
	if status == "paid" {
		payment.PaidAt = &now
	}

	if err := tx.Save(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// If verified as paid and has a request, create/update transaction
	if status == "paid" && payment.Request != nil {
		req := payment.Request

		// Calculate total paid for this request
		var totalPaid float64
		tx.Model(&models.Payment{}).
			Where("request_id = ? AND status = ?", req.ID, "paid").
			Select("COALESCE(SUM(amount), 0)").Scan(&totalPaid)

		amountRemaining := *req.TotalAmount - totalPaid
		if amountRemaining < 0 {
			amountRemaining = 0
		}

		// Check if transaction already exists
		var existingTransaction models.Transaction
		err := tx.Where("receipt_no LIKE ? AND member_id = ?", fmt.Sprintf("%%%d%%", payment.ID), payment.MemberID).First(&existingTransaction).Error

		if err == nil {
			// Update existing transaction
			existingTransaction.AmountPaid = totalPaid
			existingTransaction.AmountRemaining = amountRemaining
			if err := tx.Save(&existingTransaction).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			// Create new transaction
			receiptNo := fmt.Sprintf("REC-%d-%d", now.Year(), payment.ID)
			transaction := models.Transaction{
				MemberID:        payment.MemberID,
				FiscalYearID:    req.FiscalYearID,
				ResourceItemID:  &req.ResourceItemID,
				Type:            "resource_sale",
				Quantity:        req.QuantityApproved,
				RatePerUnit:     req.RatePerUnit,
				TotalAmount:     *req.TotalAmount,
				AmountPaid:      totalPaid,
				AmountRemaining: amountRemaining,
				ReceiptNo:       receiptNo,
				Date:            now,
			}
			if err := tx.Create(&transaction).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create transaction: %w", err)
			}
		}

		// Mark request as completed if fully paid
		if amountRemaining <= 0 {
			if err := tx.Model(req).Update("status", "completed").Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}

		// Deduct from stock (only once)
		var transactionCount int64
		tx.Model(&models.Transaction{}).
			Where("member_id = ? AND resource_item_id = ? AND fiscal_year_id = ?",
				payment.MemberID, req.ResourceItemID, req.FiscalYearID).
			Count(&transactionCount)

		if transactionCount == 0 && req.QuantityApproved != nil && *req.QuantityApproved > 0 {
			if err := tx.Model(&models.Stock{}).
				Where("resource_item_id = ? AND fiscal_year_id = ?",
					req.ResourceItemID, req.FiscalYearID).
				UpdateColumn("remaining_quantity", gorm.Expr("remaining_quantity - ?", *req.QuantityApproved)).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to update stock: %w", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload payment with relations
	if err := s.db.Preload("Member").Preload("Member.User").Preload("Request.ResourceItem.Type").First(&payment, paymentID).Error; err != nil {
		return nil, err
	}

	// Notify member
	if status == "paid" && payment.Member != nil && payment.Member.UserID != nil {
		notifService := notifications.NewNotificationService(s.db)
		notifService.NotifyUser(
			*payment.Member.UserID,
			"Payment Verified",
			fmt.Sprintf("Your payment of Rs. %.2f has been verified.", payment.Amount),
			"success",
			stringPtr("payment"),
			&payment.ID,
		)
	}

	return &payment, nil
}

func (s *PaymentService) GetPaymentStats(fiscalYearID string) (map[string]interface{}, error) {
	var stats struct {
		TotalAmount   float64 `json:"total_amount"`
		TotalCount    int64   `json:"total_count"`
		CashAmount    float64 `json:"cash_amount"`
		OnlineAmount  float64 `json:"online_amount"`
		PendingAmount float64 `json:"pending_amount"`
	}

	query := s.db.Model(&models.Payment{}).Where("status = ?", "paid")
	if fiscalYearID != "" {
		query = query.Joins("JOIN requests ON requests.id = payments.request_id").
			Where("requests.fiscal_year_id = ?", fiscalYearID)
	}

	query.Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalAmount)
	s.db.Model(&models.Payment{}).Where("status = ?", "paid").Count(&stats.TotalCount)

	s.db.Model(&models.Payment{}).Where("status = ? AND payment_method = ?", "paid", "cash").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.CashAmount)

	s.db.Model(&models.Payment{}).Where("status = ? AND payment_method IN (?)", "paid", []string{"esewa", "khalti", "bank"}).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.OnlineAmount)

	s.db.Model(&models.Payment{}).Where("status = ?", "pending").
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.PendingAmount)

	return map[string]interface{}{
		"total_amount":   stats.TotalAmount,
		"total_count":    stats.TotalCount,
		"cash_amount":    stats.CashAmount,
		"online_amount":  stats.OnlineAmount,
		"pending_amount": stats.PendingAmount,
	}, nil
}

func stringPtr(s string) *string {
	return &s
}
