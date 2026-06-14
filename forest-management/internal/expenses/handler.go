package expenses

import (
	"strconv"

	"forest-management/internal/audit"
	"forest-management/pkg/middleware"
	"forest-management/pkg/requestutil"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ExpenseHandler struct {
	service *ExpenseService
}

func NewExpenseHandler(service *ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{service: service}
}

// ==================== Expense DTOs ====================

type CreateExpenseInput struct {
	FiscalYearID  uint    `json:"fiscal_year_id" binding:"required"`
	CategoryID    uint    `json:"category_id" binding:"required"`
	Title         string  `json:"title" binding:"required,max=255"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	ExpenseDate   string  `json:"expense_date" binding:"required"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=cash bank online"`
	PaidTo        string  `json:"paid_to" binding:"required,max=255"`
	ReceiptNo     *string `json:"receipt_no"`
	BillPhoto     *string `json:"bill_photo"`
	Remarks       *string `json:"remarks"`
}

type UpdateExpenseInput struct {
	FiscalYearID  *uint    `json:"fiscal_year_id"`
	CategoryID    *uint    `json:"category_id"`
	Title         *string  `json:"title"`
	Amount        *float64 `json:"amount"`
	ExpenseDate   *string  `json:"expense_date"`
	PaymentMethod *string  `json:"payment_method"`
	PaidTo        *string  `json:"paid_to"`
	ReceiptNo     *string  `json:"receipt_no"`
	BillPhoto     *string  `json:"bill_photo"`
	Remarks       *string  `json:"remarks"`
}

// ==================== Expense Handlers ====================

func (h *ExpenseHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input CreateExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid expense data: "+err.Error())
		return
	}

	expense, err := h.service.CreateExpense(userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := userID
	audit.CreateAuditEntry(h.service.db, &actorID, "create", "expense", &expense.ID, nil, expense, c.ClientIP(), c.Request.UserAgent(), "Expense recorded")
	response.Created(c, "Expense recorded successfully", expense)
}

func (h *ExpenseHandler) List(c *gin.Context) {
	page, perPage := requestutil.Pagination(c)
	fiscalYearID := c.Query("fiscal_year_id")
	categoryID := c.Query("category_id")
	search := c.Query("search")

	expenses, meta, err := h.service.ListExpenses(page, perPage, fiscalYearID, categoryID, search)
	if err != nil {
		response.InternalError(c, "Failed to fetch expenses")
		return
	}

	response.Paginated(c, "Expenses retrieved", expenses, meta)
}

func (h *ExpenseHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	expense, err := h.service.GetExpenseByID(uint(id))
	if err != nil {
		response.NotFound(c, "Expense not found")
		return
	}

	response.Success(c, "Expense retrieved", expense)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	before, _ := h.service.GetExpenseByID(uint(id))

	var input UpdateExpenseInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid expense data")
		return
	}

	expense, err := h.service.UpdateExpense(uint(id), middleware.GetUserID(c), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "update", "expense", &expense.ID, before, expense, c.ClientIP(), c.Request.UserAgent(), "Expense updated")
	response.Success(c, "Expense updated", expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	before, _ := h.service.GetExpenseByID(uint(id))
	if err := h.service.DeleteExpense(uint(id), middleware.GetUserID(c)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	entityID := uint(id)
	audit.CreateAuditEntry(h.service.db, &actorID, "archive", "expense", &entityID, before, nil, c.ClientIP(), c.Request.UserAgent(), "Expense archived; financial row preserved")
	response.Success(c, "Expense archived", nil)
}

// ==================== Expense Category Handlers ====================

type CreateCategoryInput struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

type UpdateCategoryInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (h *ExpenseHandler) CreateCategory(c *gin.Context) {
	var input CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid category data: "+err.Error())
		return
	}

	category, err := h.service.CreateCategory(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Category created successfully", category)
}

func (h *ExpenseHandler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories()
	if err != nil {
		response.InternalError(c, "Failed to fetch categories")
		return
	}

	response.Success(c, "Categories retrieved", categories)
}

func (h *ExpenseHandler) GetCategoryByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	category, err := h.service.GetCategoryByID(uint(id))
	if err != nil {
		response.NotFound(c, "Category not found")
		return
	}

	response.Success(c, "Category retrieved", category)
}

func (h *ExpenseHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	var input UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid category data")
		return
	}

	category, err := h.service.UpdateCategory(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Category updated", category)
}

func (h *ExpenseHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	if err := h.service.DeleteCategory(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Category deleted", nil)
}

// ==================== Bill Photo Upload ====================

func (h *ExpenseHandler) UploadBillPhoto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid expense ID")
		return
	}

	file, header, err := c.Request.FormFile("photo")
	if err != nil {
		response.BadRequest(c, "Photo file is required")
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
	}
	mimeType := header.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		response.BadRequest(c, "Only image files (JPEG, PNG, GIF, WEBP) or PDF are allowed")
		return
	}

	// Max 5MB
	if header.Size > 5*1024*1024 {
		response.BadRequest(c, "File size must be less than 5MB")
		return
	}

	photoURL, err := h.service.UploadBillPhoto(uint(id), file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Bill photo uploaded successfully", gin.H{
		"photo_url": photoURL,
	})
}
