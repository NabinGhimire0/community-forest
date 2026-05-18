package transactions

import (
	"strconv"

	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	service *TransactionService
}

func NewTransactionHandler(service *TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	fiscalYearID := c.Query("fiscal_year_id")
	txnType := c.Query("type")
	memberID := c.Query("member_id")
	search := c.Query("search")

	txns, meta, err := h.service.ListTransactions(page, perPage, fiscalYearID, txnType, memberID, search)
	if err != nil {
		response.InternalError(c, "Failed to fetch transactions")
		return
	}

	response.Paginated(c, "Transactions retrieved", txns, meta)
}

func (h *TransactionHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID")
		return
	}

	txn, err := h.service.GetTransactionByID(uint(id))
	if err != nil {
		response.NotFound(c, "Transaction not found")
		return
	}

	response.Success(c, "Transaction retrieved", txn)
}

func (h *TransactionHandler) Summary(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	if fiscalYearID == "" {
		response.BadRequest(c, "fiscal_year_id is required")
		return
	}

	summary, err := h.service.GetFiscalYearSummary(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to generate summary")
		return
	}

	response.Success(c, "Financial summary", summary)
}

func (h *TransactionHandler) MyTransactions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	txns, meta, err := h.service.GetMemberTransactions(userID, page, perPage)
	if err != nil {
		response.InternalError(c, "Failed to fetch your transactions")
		return
	}

	response.Paginated(c, "Your transactions", txns, meta)
}

func (h *TransactionHandler) GetDashboardSummary(c *gin.Context) {
	summary, err := h.service.GetDashboardSummary()
	if err != nil {
		response.InternalError(c, "Failed to fetch summary")
		return
	}

	response.Success(c, "Dashboard summary", summary)
}
