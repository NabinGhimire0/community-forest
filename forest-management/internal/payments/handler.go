package payments

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"forest-management/config"
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

type CreateCashPaymentInput struct {
	MemberID            *uint   `json:"member_id"`
	RequestID           *uint   `json:"request_id"`
	LedgerTransactionID *uint   `json:"ledger_transaction_id"`
	Amount              float64 `json:"amount" binding:"required,gt=0"`
	Remarks             *string `json:"remarks"`
}

type InitiateEsewaInput struct {
	RequestID           *uint `json:"request_id"`
	LedgerTransactionID *uint `json:"ledger_transaction_id"`
}

func (h *PaymentHandler) CreateCash(c *gin.Context) {
	var input CreateCashPaymentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid cash payment data: "+err.Error())
		return
	}
	payment, err := h.service.CreateCashPayment(middleware.GetUserID(c), input, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Created(c, "Cash payment recorded successfully", payment)
}

func (h *PaymentHandler) InitiateEsewa(c *gin.Context) {
	var input InitiateEsewaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid eSewa payment target")
		return
	}
	result, err := h.service.InitiateEsewa(middleware.GetUserID(c), input.RequestID, input.LedgerTransactionID)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Created(c, "eSewa payment initialized", result)
}

// EsewaCallback is public because eSewa redirects the browser to it. The service
// validates the signed response and independently checks the transaction status.
func (h *PaymentHandler) EsewaCallback(c *gin.Context) {
	data := c.Query("data")
	payment, err := h.service.ProcessEsewaCallback(data)
	frontend := strings.TrimRight(config.AppConfig.FrontendURL, "/")
	if err != nil {
		requestID, _ := c.Get("request_id")
		log.Printf("request_id=%v eSewa callback verification failed: %v", requestID, err)
		message := "Payment verification failed. Open Payment History and use Check Status before trying again."
		c.Redirect(http.StatusFound, frontend+"/payments/esewa/result?status=error&message="+url.QueryEscape(message))
		return
	}
	c.Redirect(http.StatusFound, frontend+"/payments/esewa/result?status=success&payment_id="+strconv.FormatUint(uint64(payment.ID), 10))
}

func (h *PaymentHandler) EsewaFailure(c *gin.Context) {
	uuid := c.Query("transaction_uuid")
	payment, err := h.service.ProcessEsewaFailure(uuid)
	frontend := strings.TrimRight(config.AppConfig.FrontendURL, "/")
	if err != nil {
		requestID, _ := c.Get("request_id")
		log.Printf("request_id=%v eSewa failure reconciliation failed: %v", requestID, err)
		message := "Payment status could not be confirmed. Open Payment History and use Check Status."
		c.Redirect(http.StatusFound, frontend+"/payments/esewa/result?status=error&message="+url.QueryEscape(message))
		return
	}
	if payment.Status == "paid" {
		c.Redirect(http.StatusFound, frontend+"/payments/esewa/result?status=success&payment_id="+strconv.FormatUint(uint64(payment.ID), 10))
		return
	}
	message := "Payment was cancelled or failed"
	if payment.Status == "pending" {
		message = "Payment is still pending. Open Payment History and use Check Status before trying again."
	}
	c.Redirect(http.StatusFound, frontend+"/payments/esewa/result?status=failed&message="+url.QueryEscape(message))
}

func (h *PaymentHandler) CheckEsewaStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid payment ID")
		return
	}
	payment, err := h.service.CheckEsewaPaymentStatus(uint(id), middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, "eSewa status checked", payment)
}

func (h *PaymentHandler) List(c *gin.Context) {
	page, perPage := parsePagination(c)
	payments, meta, err := h.service.ListPayments(page, perPage, c.Query("fiscal_year_id"), c.Query("status"), c.Query("member_id"))
	if err != nil {
		response.InternalError(c, "Failed to fetch payments")
		return
	}
	response.Paginated(c, "Payments retrieved", payments, meta)
}

func (h *PaymentHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid payment ID")
		return
	}
	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	if !h.service.UserCanAccessPayment(uint(id), userID, role) {
		response.Forbidden(c, "You cannot access this payment")
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
	page, perPage := parsePagination(c)
	payments, meta, err := h.service.GetMemberPayments(middleware.GetUserID(c), page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch your payments")
		return
	}
	response.Paginated(c, "Your payments retrieved", payments, meta)
}

func (h *PaymentHandler) GetPaymentStats(c *gin.Context) {
	stats, err := h.service.GetPaymentStats(c.Query("fiscal_year_id"))
	if err != nil {
		response.InternalError(c, "Failed to fetch payment statistics")
		return
	}
	response.Success(c, "Payment statistics", stats)
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}
