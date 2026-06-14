package requests

import (
	"errors"
	"strconv"

	"forest-management/internal/audit"
	"forest-management/pkg/middleware"
	"forest-management/pkg/requestutil"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	service *RequestService
}

func NewRequestHandler(service *RequestService) *RequestHandler {
	return &RequestHandler{service: service}
}

type CreateRequestInput struct {
	MemberID          *uint   `json:"member_id"` // Required for admin/staff
	ResourceItemID    uint    `json:"resource_item_id" binding:"required"`
	FiscalYearID      uint    `json:"fiscal_year_id" binding:"required"`
	QuantityRequested float64 `json:"quantity_requested" binding:"required,min=0.01"`
	Remarks           *string `json:"remarks"`
}

type ApproveRequestInput struct {
	QuantityApproved *float64 `json:"quantity_approved"`
	Remarks          *string  `json:"remarks"`
}

type RejectRequestInput struct {
	Remarks *string `json:"remarks"`
}

func (h *RequestHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req CreateRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	request, err := h.service.CreateRequest(userID, req)
	if err != nil {
		if errors.Is(err, ErrNoActiveFiscalYear) || errors.Is(err, ErrMemberActiveYearOnly) {
			response.BadRequest(c, err.Error())
			return
		}
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "create", "request", &request.ID, nil, request, c.ClientIP(), c.Request.UserAgent(), "Resource request submitted")
	response.Created(c, "Request submitted successfully", request)
}

func (h *RequestHandler) List(c *gin.Context) {
	page, perPage := requestutil.Pagination(c)
	status := c.Query("status")
	fiscalYearID := c.Query("fiscal_year_id")
	memberID := c.Query("member_id")
	search := c.Query("search")

	requests, meta, err := h.service.ListRequests(page, perPage, status, fiscalYearID, memberID, search)
	if err != nil {
		response.InternalError(c, "Failed to fetch requests")
		return
	}

	response.Paginated(c, "Requests retrieved", requests, meta)
}

func (h *RequestHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}

	userID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	req, err := h.service.GetRequestByID(uint(id), userID, role)
	if err != nil {
		response.NotFound(c, "Request not found")
		return
	}

	response.Success(c, "Request retrieved", req)
}

func (h *RequestHandler) MyRequests(c *gin.Context) {
	userID := middleware.GetUserID(c)
	page, perPage := requestutil.Pagination(c)
	status := c.Query("status")

	requests, meta, err := h.service.GetMemberRequests(userID, page, perPage, status)
	if err != nil {
		response.InternalError(c, "Failed to fetch your requests")
		return
	}

	response.Paginated(c, "Your requests retrieved", requests, meta)
}

func (h *RequestHandler) Approve(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}
	userID := middleware.GetUserID(c)

	var req ApproveRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}

	request, err := h.service.ApproveRequest(uint(id), userID, req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "approve", "request", &request.ID, nil, request, c.ClientIP(), c.Request.UserAgent(), "Administrator approved resource request and reserved stock")
	response.Success(c, "Request approved successfully", request)
}

func (h *RequestHandler) Reject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid request ID")
		return
	}
	userID := middleware.GetUserID(c)

	var req RejectRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}

	request, err := h.service.RejectRequest(uint(id), userID, req.Remarks)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "reject", "request", &request.ID, nil, request, c.ClientIP(), c.Request.UserAgent(), "Administrator rejected resource request")
	response.Success(c, "Request rejected", request)
}

func (h *RequestHandler) GetStatistics(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	stats, err := h.service.GetRequestStatistics(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to fetch statistics")
		return
	}

	response.Success(c, "Request statistics", stats)
}
