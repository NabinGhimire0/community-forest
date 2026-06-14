package fines

import (
	"strconv"

	"forest-management/internal/audit"
	"forest-management/pkg/middleware"
	"forest-management/pkg/requestutil"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type FineHandler struct {
	service *FineService
}

func NewFineHandler(service *FineService) *FineHandler {
	return &FineHandler{service: service}
}

type CreateFineInput struct {
	FiscalYearID  uint    `json:"fiscal_year_id" binding:"required"`
	MemberID      *uint   `json:"member_id"`
	Name          string  `json:"name" binding:"omitempty,max=255"` // For non-member violators
	ViolationType string  `json:"violation_type" binding:"required,max=255"`
	Description   *string `json:"description"`
	FineAmount    float64 `json:"fine_amount" binding:"required,min=0.01"`
	IncidentDate  string  `json:"incident_date" binding:"required"`
	Photo         *string `json:"photo"`
	Remarks       *string `json:"remarks"`
}

type UpdateFineInput struct {
	ViolationType *string  `json:"violation_type"`
	Description   *string  `json:"description"`
	FineAmount    *float64 `json:"fine_amount"`
	IncidentDate  *string  `json:"incident_date"`
	Photo         *string  `json:"photo"`
	Remarks       *string  `json:"remarks"`
}

type UpdateFineStatusInput struct {
	Status           string  `json:"status" binding:"required,oneof=paid waived"`
	PaymentReference *string `json:"payment_reference"`
}

func (h *FineHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var input CreateFineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid fine data: "+err.Error())
		return
	}

	// Validate: either MemberID or Name must be provided
	if input.MemberID == nil && input.Name == "" {
		response.BadRequest(c, "Either member_id or name must be provided")
		return
	}

	fine, err := h.service.CreateFine(userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := userID
	audit.CreateAuditEntry(h.service.db, &actorID, "create", "fine", &fine.ID, nil, fine, c.ClientIP(), c.Request.UserAgent(), "Fine recorded")
	response.Created(c, "Fine recorded successfully", fine)
}

func (h *FineHandler) List(c *gin.Context) {
	page, perPage := requestutil.Pagination(c)
	status := c.Query("status")
	fiscalYearID := c.Query("fiscal_year_id")
	memberID := c.Query("member_id")
	search := c.Query("search")

	fines, meta, err := h.service.ListFines(page, perPage, status, fiscalYearID, memberID, search)
	if err != nil {
		response.InternalError(c, "Failed to fetch fines")
		return
	}

	response.Paginated(c, "Fines retrieved", fines, meta)
}

func (h *FineHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fine ID")
		return
	}

	fine, err := h.service.GetFineByID(uint(id))
	if err != nil {
		response.NotFound(c, "Fine not found")
		return
	}

	response.Success(c, "Fine retrieved", fine)
}

func (h *FineHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fine ID")
		return
	}
	userID := middleware.GetUserID(c)

	before, _ := h.service.GetFineByID(uint(id))

	var input UpdateFineInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid fine data")
		return
	}

	fine, err := h.service.UpdateFine(uint(id), userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := userID
	audit.CreateAuditEntry(h.service.db, &actorID, "update", "fine", &fine.ID, before, fine, c.ClientIP(), c.Request.UserAgent(), "Pending fine updated")
	response.Success(c, "Fine updated", fine)
}

func (h *FineHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fine ID")
		return
	}
	userID := middleware.GetUserID(c)

	before, _ := h.service.GetFineByID(uint(id))

	var input UpdateFineStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid status data")
		return
	}

	fine, err := h.service.UpdateFineStatus(uint(id), userID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := userID
	audit.CreateAuditEntry(h.service.db, &actorID, "status_change", "fine", &fine.ID, before, fine, c.ClientIP(), c.Request.UserAgent(), "Fine status changed")
	response.Success(c, "Fine status updated", fine)
}

func (h *FineHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fine ID")
		return
	}

	before, _ := h.service.GetFineByID(uint(id))
	if err := h.service.DeleteFine(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	entityID := uint(id)
	audit.CreateAuditEntry(h.service.db, &actorID, "archive", "fine", &entityID, before, nil, c.ClientIP(), c.Request.UserAgent(), "Pending fine archived; row preserved")
	response.Success(c, "Fine archived", nil)
}

func (h *FineHandler) GetStatistics(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	stats, err := h.service.GetFineStatistics(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to fetch statistics")
		return
	}

	response.Success(c, "Fine statistics", stats)
}

// UploadPhoto handles fine photo upload
func (h *FineHandler) UploadPhoto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid fine ID")
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
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	mimeType := header.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		response.BadRequest(c, "Only image files (JPEG, PNG, GIF, WEBP) are allowed")
		return
	}

	// Max 5MB
	if header.Size > 5*1024*1024 {
		response.BadRequest(c, "Image size must be less than 5MB")
		return
	}

	photoURL, err := h.service.UploadPhoto(uint(id), file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Photo uploaded successfully", gin.H{
		"photo_url": photoURL,
	})
}
