package reports

import (
	"forest-management/internal/models"
	"sort"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type ReportService struct {
	db *gorm.DB
}

func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

// GetDashboard returns summary metrics for the admin dashboard
func (s *ReportService) GetDashboard() (map[string]interface{}, error) {
	var totalMembers, activeMembers, pendingRequests, totalRequests int64
	var membersWithHistoricalDue, unverifiedHistoricalEntries int64
	var totalRevenue, totalExpenses, totalFinesCollected, balance float64
	var historicalOutstanding float64

	// Member counts
	s.db.Model(&models.Member{}).Count(&totalMembers)
	s.db.Model(&models.Member{}).Where("status = ?", "active").Count(&activeMembers)

	// Request counts
	s.db.Model(&models.Request{}).Count(&totalRequests)
	s.db.Model(&models.Request{}).Where("status = ?", "pending").Count(&pendingRequests)

	// Historical register balances across all fiscal years.
	s.db.Model(&models.Transaction{}).
		Where("type LIKE ? AND record_status = ? AND amount_remaining > 0", "legacy_%", "verified").
		Select("COALESCE(SUM(amount_remaining), 0)").Scan(&historicalOutstanding)
	s.db.Model(&models.Transaction{}).
		Where("type LIKE ? AND record_status = ? AND amount_remaining > 0", "legacy_%", "verified").
		Distinct("member_id").Count(&membersWithHistoricalDue)
	s.db.Model(&models.Transaction{}).
		Where("type LIKE ? AND record_status = ?", "legacy_%", "draft").
		Count(&unverifiedHistoricalEntries)

	// Financial totals (current active fiscal year)
	var activeFY models.FiscalYear
	if s.db.Where("is_active = ?", true).First(&activeFY).Error == nil {
		// Collected revenue from verified ledger transactions, including legacy balances.
		s.db.Model(&models.Transaction{}).
			Where("fiscal_year_id = ? AND (record_status = ? OR record_status = '')", activeFY.ID, "verified").
			Select("COALESCE(SUM(amount_paid), 0)").Scan(&totalRevenue)

		// Expenses
		s.db.Model(&models.Expense{}).
			Where("fiscal_year_id = ?", activeFY.ID).
			Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

		// Fine collections are already represented in the transaction ledger.
		s.db.Model(&models.Transaction{}).
			Where("fiscal_year_id = ? AND type = ? AND (record_status = ? OR record_status = '')", activeFY.ID, "fine", "verified").
			Select("COALESCE(SUM(amount_paid), 0)").Scan(&totalFinesCollected)

		balance = totalRevenue - totalExpenses
	}

	// Get recent requests
	var recentRequests []map[string]interface{}
	s.db.Model(&models.Request{}).
		Select("requests.id, members.name as member_name, resource_items.name as resource_name, requests.quantity_requested, requests.status, requests.requested_at").
		Joins("JOIN members ON members.id = requests.member_id").
		Joins("JOIN resource_items ON resource_items.id = requests.resource_item_id").
		Order("requests.created_at DESC").
		Limit(5).
		Scan(&recentRequests)

	// Get recent payments
	var recentPayments []map[string]interface{}
	s.db.Model(&models.Payment{}).
		Select("payments.id, members.name as member_name, payments.amount, payments.payment_method, payments.status, payments.created_at").
		Joins("JOIN members ON members.id = payments.member_id").
		Where("payments.status = ?", "paid").
		Order("payments.created_at DESC").
		Limit(5).
		Scan(&recentPayments)

	return map[string]interface{}{
		"total_members":                 totalMembers,
		"active_members":                activeMembers,
		"total_requests":                totalRequests,
		"pending_requests":              pendingRequests,
		"total_revenue":                 totalRevenue,
		"total_expenses":                totalExpenses,
		"total_fines":                   totalFinesCollected,
		"balance":                       balance,
		"historical_outstanding":        historicalOutstanding,
		"members_with_historical_due":   membersWithHistoricalDue,
		"unverified_historical_entries": unverifiedHistoricalEntries,
		"recent_requests":               recentRequests,
		"recent_payments":               recentPayments,
		"active_fiscal_year":            activeFY.Name,
	}, nil
}

// GetMemberReport returns member statistics
func (s *ReportService) GetMemberReport() (map[string]interface{}, error) {
	var totalMembers, activeMembers, inactiveMembers int64
	var wardWiseCount []struct {
		WardNo int   `json:"ward_no"`
		Count  int64 `json:"count"`
	}
	var monthlyJoining []struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}

	s.db.Model(&models.Member{}).Count(&totalMembers)
	s.db.Model(&models.Member{}).Where("status = ?", "active").Count(&activeMembers)
	s.db.Model(&models.Member{}).Where("status = ?", "inactive").Count(&inactiveMembers)

	// Ward-wise distribution
	s.db.Model(&models.Member{}).
		Select("ward_no, COUNT(*) as count").
		Group("ward_no").
		Order("ward_no ASC").
		Scan(&wardWiseCount)

	// Monthly joining trend (last 12 months)
	s.db.Model(&models.Member{}).
		Select("TO_CHAR(joined_date, 'YYYY-MM') as month, COUNT(*) as count").
		Where("joined_date IS NOT NULL AND joined_date >= NOW() - INTERVAL '12 months'").
		Group("TO_CHAR(joined_date, 'YYYY-MM')").
		Order("month ASC").
		Scan(&monthlyJoining)

	return map[string]interface{}{
		"total":           totalMembers,
		"active":          activeMembers,
		"inactive":        inactiveMembers,
		"ward_wise":       wardWiseCount,
		"monthly_joining": monthlyJoining,
	}, nil
}

// GetResourceReport returns stock and sales data for a fiscal year
func (s *ReportService) GetResourceReport(fiscalYearID string) (map[string]interface{}, error) {
	var stockData []struct {
		ResourceItemID    uint    `json:"resource_item_id"`
		ItemName          string  `json:"item_name"`
		TypeName          string  `json:"type_name"`
		Unit              string  `json:"unit"`
		TotalQuantity     float64 `json:"total_quantity"`
		RemainingQuantity float64 `json:"remaining_quantity"`
		ReservedQuantity  float64 `json:"reserved_quantity"`
		AvailableQuantity float64 `json:"available_quantity"`
		UsedQuantity      float64 `json:"used_quantity"`
		UsedPercent       float64 `json:"used_percent"`
	}

	query := s.db.Model(&models.Stock{}).
		Select("stocks.resource_item_id, resource_items.name as item_name, resource_types.name as type_name, resource_types.unit as unit, stocks.total_quantity, stocks.remaining_quantity, stocks.reserved_quantity, (stocks.remaining_quantity - stocks.reserved_quantity) as available_quantity, (stocks.total_quantity - stocks.remaining_quantity) as used_quantity, CASE WHEN stocks.total_quantity > 0 THEN ROUND(((stocks.total_quantity - stocks.remaining_quantity) / stocks.total_quantity) * 100, 2) ELSE 0 END as used_percent").
		Joins("JOIN resource_items ON resource_items.id = stocks.resource_item_id").
		Joins("JOIN resource_types ON resource_types.id = resource_items.resource_type_id")

	if fiscalYearID != "" {
		query = query.Where("stocks.fiscal_year_id = ?", fiscalYearID)
	}

	query.Scan(&stockData)

	// Total sales by resource type
	var salesByType []struct {
		TypeName    string  `json:"type_name"`
		TotalAmount float64 `json:"total_amount"`
		TotalQty    float64 `json:"total_quantity"`
	}

	salesQuery := s.db.Model(&models.Transaction{}).
		Select("resource_types.name as type_name, COALESCE(SUM(transactions.total_amount), 0) as total_amount, COALESCE(SUM(transactions.quantity), 0) as total_quantity").
		Joins("JOIN resource_items ON resource_items.id = transactions.resource_item_id").
		Joins("JOIN resource_types ON resource_types.id = resource_items.resource_type_id").
		Where("transactions.type = ? AND (transactions.record_status = ? OR transactions.record_status = '')", "resource_sale", "verified")

	if fiscalYearID != "" {
		salesQuery = salesQuery.Where("transactions.fiscal_year_id = ?", fiscalYearID)
	}

	salesQuery.Group("resource_types.name").Order("total_amount DESC").Scan(&salesByType)

	return map[string]interface{}{
		"stock":         stockData,
		"sales_by_type": salesByType,
	}, nil
}

// GetFinancialReport returns comprehensive financial data
func (s *ReportService) GetFinancialReport(fiscalYearID string) (map[string]interface{}, error) {
	var totalSales, totalMembershipFee, totalCollected float64
	var totalExpenses, totalFinesCollected, historicalOutstanding, historicalCollected float64

	// Revenue from resource sales
	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type IN ? AND (record_status = ? OR record_status = '')", fiscalYearID, []string{"resource_sale", "legacy_timber_sale", "legacy_firewood_sale", "legacy_other_sale"}, "verified").
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalSales)

	// Revenue from membership fees
	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type IN ? AND (record_status = ? OR record_status = '')", fiscalYearID, []string{"membership_fee", "legacy_gasti_fee"}, "verified").
		Select("COALESCE(SUM(total_amount), 0)").Scan(&totalMembershipFee)

	// Total collected
	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND (record_status = ? OR record_status = '')", fiscalYearID, "verified").
		Select("COALESCE(SUM(amount_paid), 0)").Scan(&totalCollected)

	// Historical balance recovery and outstanding amount.
	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type LIKE 'legacy_%%' AND (record_status = ? OR record_status = '')", fiscalYearID, "verified").
		Select("COALESCE(SUM(amount_paid), 0)").Scan(&historicalCollected)
	s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type LIKE 'legacy_%%' AND (record_status = ? OR record_status = '')", fiscalYearID, "verified").
		Select("COALESCE(SUM(amount_remaining), 0)").Scan(&historicalOutstanding)

	// Total expenses
	s.db.Model(&models.Expense{}).
		Where("fiscal_year_id = ?", fiscalYearID).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalExpenses)

	// Total fines collected
	s.db.Model(&models.Fine{}).
		Where("fiscal_year_id = ? AND status = ?", fiscalYearID, "paid").
		Select("COALESCE(SUM(fine_amount), 0)").Scan(&totalFinesCollected)

	// Monthly financial data
	var monthlyData []struct {
		Month   string  `json:"month"`
		Income  float64 `json:"income"`
		Expense float64 `json:"expense"`
	}

	// Monthly income
	s.db.Model(&models.Transaction{}).
		Select("TO_CHAR(date, 'YYYY-MM') as month, COALESCE(SUM(amount_paid), 0) as income").
		Where("fiscal_year_id = ? AND (record_status = ? OR record_status = '')", fiscalYearID, "verified").
		Group("TO_CHAR(date, 'YYYY-MM')").
		Order("month ASC").
		Scan(&monthlyData)

	// Monthly expenses
	var monthlyExpenses []struct {
		Month   string  `json:"month"`
		Expense float64 `json:"expense"`
	}
	s.db.Model(&models.Expense{}).
		Select("TO_CHAR(expense_date, 'YYYY-MM') as month, COALESCE(SUM(amount), 0) as expense").
		Where("fiscal_year_id = ?", fiscalYearID).
		Group("TO_CHAR(expense_date, 'YYYY-MM')").
		Order("month ASC").
		Scan(&monthlyExpenses)

	// Merge monthly data
	monthlyMap := make(map[string]map[string]float64)
	for _, m := range monthlyData {
		if monthlyMap[m.Month] == nil {
			monthlyMap[m.Month] = make(map[string]float64)
		}
		monthlyMap[m.Month]["income"] = m.Income
	}
	for _, m := range monthlyExpenses {
		if monthlyMap[m.Month] == nil {
			monthlyMap[m.Month] = make(map[string]float64)
		}
		monthlyMap[m.Month]["expense"] = m.Expense
	}

	var mergedMonthly []map[string]interface{}
	for month, data := range monthlyMap {
		mergedMonthly = append(mergedMonthly, map[string]interface{}{
			"month":   month,
			"income":  data["income"],
			"expense": data["expense"],
		})
	}
	sort.Slice(mergedMonthly, func(i, j int) bool {
		return mergedMonthly[i]["month"].(string) < mergedMonthly[j]["month"].(string)
	})

	// Category-wise expenses
	var categoryExpenses []struct {
		CategoryName string  `json:"category_name"`
		TotalAmount  float64 `json:"total_amount"`
	}
	s.db.Model(&models.Expense{}).
		Select("expense_categories.name as category_name, COALESCE(SUM(expenses.amount), 0) as total_amount").
		Joins("JOIN expense_categories ON expense_categories.id = expenses.category_id").
		Where("expenses.fiscal_year_id = ?", fiscalYearID).
		Group("expense_categories.name").
		Order("total_amount DESC").
		Scan(&categoryExpenses)

	// Payment method distribution
	var paymentMethods []struct {
		Method string  `json:"payment_method"`
		Total  float64 `json:"total"`
		Count  int64   `json:"count"`
	}
	paymentQuery := s.db.Model(&models.Payment{}).
		Select("payments.payment_method, COALESCE(SUM(payments.amount), 0) as total, COUNT(*) as count").
		Joins("LEFT JOIN requests ON requests.id = payments.request_id").
		Joins("LEFT JOIN transactions ledger_targets ON ledger_targets.id = payments.ledger_transaction_id").
		Where("payments.status = ?", "paid")
	if fiscalYearID != "" {
		paymentQuery = paymentQuery.Where("COALESCE(requests.fiscal_year_id, ledger_targets.fiscal_year_id) = ?", fiscalYearID)
	}
	paymentQuery.Group("payments.payment_method").Scan(&paymentMethods)

	return map[string]interface{}{
		"total_revenue":          totalSales + totalMembershipFee,
		"resource_sales":         totalSales,
		"membership_fees":        totalMembershipFee,
		"total_collected":        totalCollected,
		"historical_collected":   historicalCollected,
		"historical_outstanding": historicalOutstanding,
		"total_expenses":         totalExpenses,
		"total_fines":            totalFinesCollected,
		"net_balance":            totalCollected - totalExpenses,
		"monthly_data":           mergedMonthly,
		"category_expenses":      categoryExpenses,
		"payment_methods":        paymentMethods,
	}, nil
}

// GetDashboardCharts returns time-series data for charts
func (s *ReportService) GetDashboardCharts() (map[string]interface{}, error) {
	var activeFY models.FiscalYear
	if err := s.db.Where("is_active = ?", true).First(&activeFY).Error; err != nil {
		// If no active fiscal year, get the most recent one
		s.db.Order("start_date DESC").First(&activeFY)
	}

	charts := map[string]interface{}{}

	// 1. Monthly Revenue Chart (last 12 months)
	var monthlyFinancials []struct {
		Month   string  `json:"month"`
		Income  float64 `json:"income"`
		Expense float64 `json:"expense"`
	}

	var monthlyIncome []struct {
		Month  string  `json:"month"`
		Income float64 `json:"income"`
	}
	s.db.Model(&models.Transaction{}).
		Select("TO_CHAR(date, 'YYYY-MM') as month, COALESCE(SUM(amount_paid), 0) as income").
		Where("fiscal_year_id = ? AND (record_status = ? OR record_status = '')", activeFY.ID, "verified").
		Group("TO_CHAR(date, 'YYYY-MM')").
		Order("month ASC").
		Scan(&monthlyIncome)

	var monthlyExpenses []struct {
		Month   string  `json:"month"`
		Expense float64 `json:"expense"`
	}
	s.db.Model(&models.Expense{}).
		Select("TO_CHAR(expense_date, 'YYYY-MM') as month, COALESCE(SUM(amount), 0) as expense").
		Where("fiscal_year_id = ?", activeFY.ID).
		Group("TO_CHAR(expense_date, 'YYYY-MM')").
		Order("month ASC").
		Scan(&monthlyExpenses)

	monthlyMap := make(map[string]*struct {
		Month   string
		Income  float64
		Expense float64
	})
	for _, item := range monthlyIncome {
		monthlyMap[item.Month] = &struct {
			Month   string
			Income  float64
			Expense float64
		}{Month: item.Month, Income: item.Income}
	}
	for _, item := range monthlyExpenses {
		if monthlyMap[item.Month] == nil {
			monthlyMap[item.Month] = &struct {
				Month   string
				Income  float64
				Expense float64
			}{Month: item.Month}
		}
		monthlyMap[item.Month].Expense = item.Expense
	}
	months := make([]string, 0, len(monthlyMap))
	for month := range monthlyMap {
		months = append(months, month)
	}
	sort.Strings(months)
	for _, month := range months {
		item := monthlyMap[month]
		monthlyFinancials = append(monthlyFinancials, struct {
			Month   string  `json:"month"`
			Income  float64 `json:"income"`
			Expense float64 `json:"expense"`
		}{Month: item.Month, Income: item.Income, Expense: item.Expense})
	}

	charts["monthly_financials"] = monthlyFinancials

	// 2. Request Status Distribution
	var requestStatusData []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	s.db.Model(&models.Request{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&requestStatusData)
	charts["request_status_distribution"] = requestStatusData

	// 3. Resource Sales by Type
	var resourceSalesData []struct {
		TypeName    string  `json:"type_name"`
		TotalAmount float64 `json:"total_amount"`
		TotalQty    float64 `json:"total_quantity"`
	}
	s.db.Model(&models.Transaction{}).
		Select("resource_types.name as type_name, COALESCE(SUM(transactions.total_amount), 0) as total_amount, COALESCE(SUM(transactions.quantity), 0) as total_quantity").
		Joins("JOIN resource_items ON resource_items.id = transactions.resource_item_id").
		Joins("JOIN resource_types ON resource_types.id = resource_items.resource_type_id").
		Where("transactions.fiscal_year_id = ? AND transactions.type = ? AND (transactions.record_status = ? OR transactions.record_status = '')", activeFY.ID, "resource_sale", "verified").
		Group("resource_types.name").
		Order("total_amount DESC").
		Limit(6).
		Scan(&resourceSalesData)
	charts["resource_sales_by_type"] = resourceSalesData

	// 4. Ward-wise Member Distribution
	var wardData []struct {
		WardNo int   `json:"ward_no"`
		Count  int64 `json:"count"`
	}
	s.db.Model(&models.Member{}).
		Select("ward_no, COUNT(*) as count").
		Where("status = ?", "active").
		Group("ward_no").
		Order("ward_no ASC").
		Scan(&wardData)
	charts["ward_wise_members"] = wardData

	// 5. Expense by Category
	var expenseCategoryData []struct {
		CategoryName string  `json:"category_name"`
		TotalAmount  float64 `json:"total_amount"`
	}
	s.db.Model(&models.Expense{}).
		Select("expense_categories.name as category_name, COALESCE(SUM(expenses.amount), 0) as total_amount").
		Joins("JOIN expense_categories ON expense_categories.id = expenses.category_id").
		Where("expenses.fiscal_year_id = ?", activeFY.ID).
		Group("expense_categories.name").
		Order("total_amount DESC").
		Limit(6).
		Scan(&expenseCategoryData)
	charts["expense_by_category"] = expenseCategoryData

	// 6. Payment Method Distribution
	var paymentMethodData []struct {
		Method string  `json:"payment_method"`
		Count  int64   `json:"count"`
		Total  float64 `json:"total"`
	}
	s.db.Model(&models.Payment{}).
		Select("payments.payment_method, COUNT(*) as count, COALESCE(SUM(payments.amount), 0) as total").
		Joins("LEFT JOIN requests ON requests.id = payments.request_id").
		Joins("LEFT JOIN transactions ledger_targets ON ledger_targets.id = payments.ledger_transaction_id").
		Where("payments.status = ? AND COALESCE(requests.fiscal_year_id, ledger_targets.fiscal_year_id) = ?", "paid", activeFY.ID).
		Group("payments.payment_method").
		Scan(&paymentMethodData)
	charts["payment_method_distribution"] = paymentMethodData

	// 7. Member Growth Over Time (last 12 months)
	var memberGrowth []struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}
	s.db.Model(&models.Member{}).
		Select("TO_CHAR(created_at, 'Mon YYYY') as month, COUNT(*) as count").
		Where("created_at >= NOW() - INTERVAL '12 months'").
		Group("TO_CHAR(created_at, 'Mon YYYY'), EXTRACT(MONTH FROM created_at)").
		Order("MIN(created_at) ASC").
		Limit(12).
		Scan(&memberGrowth)
	charts["member_growth"] = memberGrowth

	// 8. Fine Status Summary
	var fineStatusData []struct {
		Status string  `json:"status"`
		Count  int64   `json:"count"`
		Total  float64 `json:"total"`
	}
	s.db.Model(&models.Fine{}).
		Select("status, COUNT(*) as count, COALESCE(SUM(fine_amount), 0) as total").
		Where("fiscal_year_id = ?", activeFY.ID).
		Group("status").
		Scan(&fineStatusData)
	charts["fine_status"] = fineStatusData

	// 9. Stock Overview
	var stockOverview []struct {
		ItemName    string  `json:"item_name"`
		TypeName    string  `json:"type_name"`
		Total       float64 `json:"total_quantity"`
		Remaining   float64 `json:"remaining_quantity"`
		Used        float64 `json:"used_quantity"`
		UsedPercent float64 `json:"used_percent"`
	}
	s.db.Model(&models.Stock{}).
		Select("resource_items.name as item_name, resource_types.name as type_name, stocks.total_quantity, stocks.remaining_quantity, stocks.reserved_quantity, (stocks.remaining_quantity - stocks.reserved_quantity) as available_quantity, (stocks.total_quantity - stocks.remaining_quantity) as used_quantity, CASE WHEN stocks.total_quantity > 0 THEN ROUND(((stocks.total_quantity - stocks.remaining_quantity) / stocks.total_quantity) * 100, 2) ELSE 0 END as used_percent").
		Joins("JOIN resource_items ON resource_items.id = stocks.resource_item_id").
		Joins("JOIN resource_types ON resource_types.id = resource_items.resource_type_id").
		Where("stocks.fiscal_year_id = ?", activeFY.ID).
		Limit(8).
		Scan(&stockOverview)
	charts["stock_overview"] = stockOverview

	// 10. Recent Activities
	var recentActivities []struct {
		Action      string    `json:"action"`
		Description string    `json:"description"`
		Type        string    `json:"type"`
		CreatedAt   time.Time `json:"created_at"`
	}

	// Get recent requests
	var recentReqs []struct {
		MemberName string    `json:"member_name"`
		ItemName   string    `json:"item_name"`
		Status     string    `json:"status"`
		CreatedAt  time.Time `json:"created_at"`
	}
	s.db.Model(&models.Request{}).
		Select("members.name as member_name, resource_items.name as item_name, requests.status, requests.created_at").
		Joins("JOIN members ON members.id = requests.member_id").
		Joins("JOIN resource_items ON resource_items.id = requests.resource_item_id").
		Order("requests.created_at DESC").
		Limit(5).
		Scan(&recentReqs)

	for _, r := range recentReqs {
		recentActivities = append(recentActivities, struct {
			Action      string    `json:"action"`
			Description string    `json:"description"`
			Type        string    `json:"type"`
			CreatedAt   time.Time `json:"created_at"`
		}{
			Action:      "New Request",
			Description: r.MemberName + " requested " + r.ItemName,
			Type:        "request",
			CreatedAt:   r.CreatedAt,
		})
	}

	// Get recent payments
	var recentPays []struct {
		MemberName string    `json:"member_name"`
		Amount     float64   `json:"amount"`
		CreatedAt  time.Time `json:"created_at"`
	}
	s.db.Model(&models.Payment{}).
		Select("members.name as member_name, payments.amount, payments.created_at").
		Joins("JOIN members ON members.id = payments.member_id").
		Where("payments.status = ?", "paid").
		Order("payments.created_at DESC").
		Limit(5).
		Scan(&recentPays)

	for _, p := range recentPays {
		recentActivities = append(recentActivities, struct {
			Action      string    `json:"action"`
			Description string    `json:"description"`
			Type        string    `json:"type"`
			CreatedAt   time.Time `json:"created_at"`
		}{
			Action:      "Payment Received",
			Description: p.MemberName + " paid Rs. " + formatCurrency(p.Amount),
			Type:        "payment",
			CreatedAt:   p.CreatedAt,
		})
	}

	// Sort by created_at desc and limit to 10
	sort.Slice(recentActivities, func(i, j int) bool {
		return recentActivities[i].CreatedAt.After(recentActivities[j].CreatedAt)
	})
	if len(recentActivities) > 10 {
		recentActivities = recentActivities[:10]
	}
	charts["recent_activities"] = recentActivities

	return charts, nil
}

func formatCurrency(amount float64) string {
	return "Rs. " + formatNumber(amount)
}

func formatNumber(num float64) string {
	if num >= 1000000 {
		return formatFloat(num/1000000) + "M"
	}
	if num >= 1000 {
		return formatFloat(num/1000) + "K"
	}
	return formatFloat(num)
}

func formatFloat(num float64) string {
	return strconv.FormatFloat(num, 'f', 1, 64)
}
