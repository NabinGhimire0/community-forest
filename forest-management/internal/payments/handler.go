package payments

import (
	"strconv"

	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	service *PaymentService
}

func NewPaymentHandler(service *PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type CreatePaymentInput struct {
	MemberID      *uint   `json:"member_id"` // For admin/staff
	RequestID     *uint   `json:"request_id"`
	Amount        float64 `json:"amount" binding:"required,min=0.01"`
	PaymentMethod string  `json:"payment_method" binding:"required,oneof=cash esewa khalti bank"`
	TransactionID *string `json:"transaction_id"`
}

type VerifyPaymentInput struct {
	Status string `json:"status" binding:"required,oneof=paid failed"`
}

func (h *PaymentHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input CreatePaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid payment data: "+err.Error())
		return
	}

	payment, err := h.service.CreatePayment(userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Payment recorded successfully", payment)
}

func (h *PaymentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	fiscalYearID := c.Query("fiscal_year_id")
	status := c.Query("status")
	memberID := c.Query("member_id")

	payments, meta, err := h.service.ListPayments(page, perPage, fiscalYearID, status, memberID)
	if err != nil {
		response.InternalError(c, "Failed to fetch payments")
		return
	}

	response.Paginated(c, "Payments retrieved", payments, meta)
}

func (h *PaymentHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}

	payment, err := h.service.GetPaymentByID(uint(id))
	if err != nil {
		response.NotFound(c, "Payment not found")
		return
	}

	response.Success(c, "Payment retrieved", payment)
}

func (h *PaymentHandler) MyPayments(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	payments, meta, err := h.service.GetMemberPayments(userID, page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch your payments")
		return
	}

	response.Paginated(c, "Your payments retrieved", payments, meta)
}

func (h *PaymentHandler) Verify(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid payment ID")
		return
	}
	adminUserID := middleware.GetUserID(c)

	var input VerifyPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid status")
		return
	}

	payment, err := h.service.VerifyPayment(uint(id), adminUserID, input.Status)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Payment verified successfully", payment)
}

func (h *PaymentHandler) GetPaymentStats(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	stats, err := h.service.GetPaymentStats(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to fetch payment statistics")
		return
	}

	response.Success(c, "Payment statistics", stats)
}
