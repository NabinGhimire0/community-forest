package fiscalyears

import (
	"strconv"

	"forest-management/internal/audit"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type FiscalYearHandler struct {
	service *FiscalYearService
}

func NewFiscalYearHandler(service *FiscalYearService) *FiscalYearHandler {
	return &FiscalYearHandler{service: service}
}

type CreateFiscalYearInput struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
}

type UpdateFiscalYearInput struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	IsActive  *bool  `json:"is_active"`
}

func (h *FiscalYearHandler) Create(c *gin.Context) {
	var input CreateFiscalYearInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}
	fy, err := h.service.Create(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "create", "fiscal_year", &fy.ID, nil, fy, c.ClientIP(), c.Request.UserAgent(), "Fiscal year created")
	response.Created(c, "Fiscal year created", fy)
}

func (h *FiscalYearHandler) List(c *gin.Context) {
	fyList, err := h.service.List()
	if err != nil {
		response.InternalError(c, "Failed to fetch fiscal years")
		return
	}
	response.Success(c, "Fiscal years retrieved", fyList)
}

func (h *FiscalYearHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	fy, err := h.service.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "Fiscal year not found")
		return
	}
	response.Success(c, "Fiscal year retrieved", fy)
}

func (h *FiscalYearHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	before, _ := h.service.GetByID(uint(id))
	var input UpdateFiscalYearInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}
	fy, err := h.service.Update(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "update", "fiscal_year", &fy.ID, before, fy, c.ClientIP(), c.Request.UserAgent(), "Fiscal year metadata updated")
	response.Success(c, "Fiscal year updated", fy)
}

func (h *FiscalYearHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	before, _ := h.service.GetByID(uint(id))
	if err := h.service.Delete(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	entityID := uint(id)
	audit.CreateAuditEntry(h.service.db, &actorID, "delete_unused", "fiscal_year", &entityID, before, nil, c.ClientIP(), c.Request.UserAgent(), "Unused fiscal year deleted")
	response.Success(c, "Fiscal year deleted", nil)
}

func (h *FiscalYearHandler) SetActive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	before, _ := h.service.GetByID(uint(id))
	fy, err := h.service.SetActive(uint(id))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "activate", "fiscal_year", &fy.ID, before, fy, c.ClientIP(), c.Request.UserAgent(), "Fiscal year activated; stock/rates/fees rolled forward and annual fees assigned")
	response.Success(c, "Active fiscal year set", fy)
}

// ==================== Fee Settings Handlers ====================

type SetFeeInput struct {
	FiscalYearID  uint    `json:"fiscal_year_id" binding:"required"`
	MembershipFee float64 `json:"membership_fee" binding:"required"`
}

type UpdateFeeInput struct {
	MembershipFee float64 `json:"membership_fee" binding:"required"`
}

// SetFee handles POST /api/fiscal-years/fee
func (h *FiscalYearHandler) SetFee(c *gin.Context) {
	var input SetFeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}

	// Validate fiscal year exists
	fee, err := h.service.SetFee(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "set_fee", "fee_setting", &fee.ID, nil, fee, c.ClientIP(), c.Request.UserAgent(), "Annual Gasti/Membership fee configured")
	response.Created(c, "Membership fee set successfully", fee)
}

// GetFee handles GET /api/fiscal-years/fee?fiscal_year_id=xxx
func (h *FiscalYearHandler) GetFee(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	if fiscalYearID == "" {
		response.BadRequest(c, "fiscal_year_id query parameter is required")
		return
	}
	fee, err := h.service.GetFee(fiscalYearID)
	if err != nil {
		response.NotFound(c, "Fee not configured for this fiscal year")
		return
	}
	response.Success(c, "Fee retrieved", fee)
}

// GetFeesByFiscalYear handles GET /api/fiscal-years/:id/fees
func (h *FiscalYearHandler) GetFeesByFiscalYear(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fiscal year ID")
		return
	}
	fees, err := h.service.GetFeesByFiscalYear(uint(id))
	if err != nil {
		response.InternalError(c, "Failed to fetch fees")
		return
	}
	response.Success(c, "Fees retrieved", fees)
}

// UpdateFee handles PUT /api/fiscal-years/fee/:id
func (h *FiscalYearHandler) UpdateFee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fee ID")
		return
	}
	var input UpdateFeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}
	fee, err := h.service.UpdateFee(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "update_fee", "fee_setting", &fee.ID, nil, fee, c.ClientIP(), c.Request.UserAgent(), "Annual Gasti/Membership fee updated; only unpaid system-generated charges synchronized")
	response.Success(c, "Fee updated", fee)
}

// DeleteFee handles DELETE /api/fiscal-years/fee/:id
func (h *FiscalYearHandler) DeleteFee(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fee ID")
		return
	}
	if err := h.service.DeleteFee(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	entityID := uint(id)
	audit.CreateAuditEntry(h.service.db, &actorID, "delete_unused_fee", "fee_setting", &entityID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "Unused fee setting deleted")
	response.Success(c, "Fee deleted", nil)
}
