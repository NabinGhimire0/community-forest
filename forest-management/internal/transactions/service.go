package transactions

import (
	"errors"

	"forest-management/internal/models"
	"forest-management/pkg/response"

	"gorm.io/gorm"
)

type TransactionService struct {
	db *gorm.DB
}

func NewTransactionService(db *gorm.DB) *TransactionService {
	return &TransactionService{db: db}
}

func (s *TransactionService) ListTransactions(page, perPage int, fiscalYearID, txnType, memberID, search string) ([]models.Transaction, *response.Pagination, error) {
	var txns []models.Transaction
	var total int64

	query := s.db.Model(&models.Transaction{})

	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	if txnType != "" {
		query = query.Where("type = ?", txnType)
	}
	if memberID != "" {
		query = query.Where("member_id = ?", memberID)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Joins("JOIN members ON members.id = transactions.member_id").
			Where("members.name ILIKE ? OR members.membership_no ILIKE ? OR transactions.receipt_no ILIKE ?",
				searchPattern, searchPattern, searchPattern)
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
		Preload("ResourceItem.Type").
		Order("date DESC").
		Offset(offset).
		Limit(perPage).
		Find(&txns).Error

	return txns, &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

func (s *TransactionService) GetTransactionByID(id uint) (*models.Transaction, error) {
	var txn models.Transaction
	err := s.db.
		Preload("Member").
		Preload("FiscalYear").
		Preload("ResourceItem.Type").
		First(&txn, id).Error
	return &txn, err
}

func (s *TransactionService) GetMemberTransactions(userID uint, page, perPage int) ([]models.Transaction, *response.Pagination, error) {
	var member models.Member
	if err := s.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, nil, errors.New("member not found")
	}

	return s.ListTransactions(page, perPage, "", "", string(rune(member.ID)), "")
}

func (s *TransactionService) GetFiscalYearSummary(fiscalYearID string) (map[string]interface{}, error) {
	var totalSales, totalMembershipFee, totalCollected, totalRemaining float64

	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type = ?", fiscalYearID, "resource_sale").
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalSales)

	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type = ?", fiscalYearID, "membership_fee").
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalMembershipFee)

	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ?", fiscalYearID).
		Select("COALESCE(SUM(amount_paid), 0)").Scan(&totalCollected)

	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ?", fiscalYearID).
		Select("COALESCE(SUM(amount_remaining), 0)").Scan(&totalRemaining)

	var totalExpenses float64
	s.db.Model(&models.Expense{}).
		Where("fiscal_year_id = ?", fiscalYearID).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

	var totalFines float64
	s.db.Model(&models.Fine{}).
		Where("fiscal_year_id = ? AND status = ?", fiscalYearID, "paid").
		Select("COALESCE(SUM(fine_amount), 0)").Scan(&totalFines)

	return map[string]interface{}{
		"total_revenue":         totalSales + totalMembershipFee,
		"resource_sales":        totalSales,
		"membership_fees":       totalMembershipFee,
		"total_collected":       totalCollected,
		"total_remaining":       totalRemaining,
		"total_expenses":        totalExpenses,
		"total_fines_collected": totalFines,
		"net_balance":           (totalSales + totalMembershipFee + totalFines) - totalExpenses,
	}, nil
}

func (s *TransactionService) GetDashboardSummary() (map[string]interface{}, error) {
	var totalMembers int64
	var pendingRequests int64
	var totalRevenue float64
	var totalExpenses float64

	s.db.Model(&models.Member{}).Count(&totalMembers)
	s.db.Model(&models.Request{}).Where("status = ?", "pending").Count(&pendingRequests)
	s.db.Model(&models.Transaction{}).Select("COALESCE(SUM(amount_paid), 0)").Scan(&totalRevenue)
	s.db.Model(&models.Expense{}).Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

	return map[string]interface{}{
		"total_members":    totalMembers,
		"pending_requests": pendingRequests,
		"total_revenue":    totalRevenue,
		"total_expenses":   totalExpenses,
		"net_balance":      totalRevenue - totalExpenses,
	}, nil
}
