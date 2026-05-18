package expenses

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"forest-management/internal/models"
	"forest-management/pkg/response"

	"gorm.io/gorm"
)

type ExpenseService struct {
	db *gorm.DB
}

func NewExpenseService(db *gorm.DB) *ExpenseService {
	return &ExpenseService{db: db}
}

// ==================== Expense Methods ====================

func (s *ExpenseService) CreateExpense(adminUserID uint, input CreateExpenseInput) (*models.Expense, error) {
	expenseDate, err := time.Parse("2006-01-02", input.ExpenseDate)
	if err != nil {
		return nil, errors.New("invalid expense date format. Use YYYY-MM-DD")
	}

	// Validate fiscal year exists
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	// Validate category exists
	var category models.ExpenseCategory
	if err := s.db.First(&category, input.CategoryID).Error; err != nil {
		return nil, errors.New("expense category not found")
	}

	expense := models.Expense{
		FiscalYearID:  input.FiscalYearID,
		CategoryID:    input.CategoryID,
		Title:         input.Title,
		Amount:        input.Amount,
		ExpenseDate:   expenseDate,
		PaymentMethod: input.PaymentMethod,
		PaidTo:        input.PaidTo,
		ReceiptNo:     input.ReceiptNo,
		BillPhoto:     input.BillPhoto,
		Remarks:       input.Remarks,
		CreatedBy:     adminUserID,
	}

	if err := s.db.Create(&expense).Error; err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	s.db.Preload("Category").Preload("FiscalYear").Preload("Creator").First(&expense, expense.ID)
	return &expense, nil
}

func (s *ExpenseService) ListExpenses(page, perPage int, fiscalYearID, categoryID, search string) ([]models.Expense, *response.Pagination, error) {
	var expenses []models.Expense
	var total int64

	query := s.db.Model(&models.Expense{})

	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"title ILIKE ? OR paid_to ILIKE ? OR receipt_no ILIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	query.Count(&total)
	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("Category").
		Preload("FiscalYear").
		Preload("Creator").
		Order("expense_date DESC").
		Offset(offset).
		Limit(perPage).
		Find(&expenses).Error

	return expenses, &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

func (s *ExpenseService) GetExpenseByID(id uint) (*models.Expense, error) {
	var expense models.Expense
	err := s.db.Preload("Category").Preload("FiscalYear").Preload("Creator").First(&expense, id).Error
	return &expense, err
}

func (s *ExpenseService) UpdateExpense(id uint, input UpdateExpenseInput) (*models.Expense, error) {
	var expense models.Expense
	if err := s.db.First(&expense, id).Error; err != nil {
		return nil, errors.New("expense not found")
	}

	updates := make(map[string]interface{})

	if input.FiscalYearID != nil {
		var fiscalYear models.FiscalYear
		if err := s.db.First(&fiscalYear, *input.FiscalYearID).Error; err != nil {
			return nil, errors.New("fiscal year not found")
		}
		updates["fiscal_year_id"] = *input.FiscalYearID
	}
	if input.CategoryID != nil {
		var category models.ExpenseCategory
		if err := s.db.First(&category, *input.CategoryID).Error; err != nil {
			return nil, errors.New("expense category not found")
		}
		updates["category_id"] = *input.CategoryID
	}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Amount != nil {
		updates["amount"] = *input.Amount
	}
	if input.ExpenseDate != nil {
		expenseDate, err := time.Parse("2006-01-02", *input.ExpenseDate)
		if err != nil {
			return nil, errors.New("invalid expense date format")
		}
		updates["expense_date"] = expenseDate
	}
	if input.PaymentMethod != nil {
		updates["payment_method"] = *input.PaymentMethod
	}
	if input.PaidTo != nil {
		updates["paid_to"] = *input.PaidTo
	}
	if input.ReceiptNo != nil {
		updates["receipt_no"] = *input.ReceiptNo
	}
	if input.BillPhoto != nil {
		updates["bill_photo"] = *input.BillPhoto
	}
	if input.Remarks != nil {
		updates["remarks"] = *input.Remarks
	}

	if len(updates) > 0 {
		if err := s.db.Model(&expense).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update expense: %w", err)
		}
	}

	s.db.Preload("Category").Preload("FiscalYear").Preload("Creator").First(&expense, id)
	return &expense, nil
}

func (s *ExpenseService) DeleteExpense(id uint) error {
	var expense models.Expense
	if err := s.db.First(&expense, id).Error; err != nil {
		return errors.New("expense not found")
	}

	// Delete associated bill photo if exists
	if expense.BillPhoto != nil && *expense.BillPhoto != "" {
		filePath := "." + *expense.BillPhoto
		os.Remove(filePath)
	}

	return s.db.Delete(&expense).Error
}

// ==================== Expense Category Methods ====================

func (s *ExpenseService) CreateCategory(input CreateCategoryInput) (*models.ExpenseCategory, error) {
	// Check if category already exists
	var existing models.ExpenseCategory
	if s.db.Where("name = ?", input.Name).First(&existing).Error == nil {
		return nil, errors.New("category with this name already exists")
	}

	category := models.ExpenseCategory{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.db.Create(&category).Error; err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

func (s *ExpenseService) ListCategories() ([]models.ExpenseCategory, error) {
	var categories []models.ExpenseCategory
	err := s.db.Order("name ASC").Find(&categories).Error
	return categories, err
}

func (s *ExpenseService) GetCategoryByID(id uint) (*models.ExpenseCategory, error) {
	var category models.ExpenseCategory
	err := s.db.First(&category, id).Error
	return &category, err
}

func (s *ExpenseService) UpdateCategory(id uint, input UpdateCategoryInput) (*models.ExpenseCategory, error) {
	var category models.ExpenseCategory
	if err := s.db.First(&category, id).Error; err != nil {
		return nil, errors.New("category not found")
	}

	updates := make(map[string]interface{})

	if input.Name != nil {
		// Check unique name
		var existing models.ExpenseCategory
		if s.db.Where("name = ? AND id != ?", *input.Name, id).First(&existing).Error == nil {
			return nil, errors.New("category with this name already exists")
		}
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	if len(updates) > 0 {
		if err := s.db.Model(&category).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update category: %w", err)
		}
	}

	s.db.First(&category, id)
	return &category, nil
}

func (s *ExpenseService) DeleteCategory(id uint) error {
	// Check if category has expenses
	var count int64
	s.db.Model(&models.Expense{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete category: it has associated expenses")
	}
	return s.db.Delete(&models.ExpenseCategory{}, id).Error
}

// ==================== Bill Photo Upload ====================

func (s *ExpenseService) UploadBillPhoto(expenseID uint, file io.Reader, filename string) (string, error) {
	var expense models.Expense
	if err := s.db.First(&expense, expenseID).Error; err != nil {
		return "", errors.New("expense not found")
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/expenses"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	uniqueName := fmt.Sprintf("expense_%d_%d%s", expenseID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	fileURL := fmt.Sprintf("/uploads/expenses/%s", uniqueName)

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

	// Update expense with bill photo URL
	if err := s.db.Model(&expense).Update("bill_photo", fileURL).Error; err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to update expense: %w", err)
	}

	return fileURL, nil
}
