package samiti

import (
	"strconv"

	"forest-management/internal/audit"
	"forest-management/internal/auth"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type SamitiHandler struct {
	service *SamitiService
}

func NewSamitiHandler(service *SamitiService) *SamitiHandler {
	return &SamitiHandler{service: service}
}

// ==================== Samiti Settings ====================

func (h *SamitiHandler) GetSettings(c *gin.Context) {
	settings, err := h.service.GetSettings()
	if err != nil {
		response.NotFound(c, "Settings not found")
		return
	}
	response.Success(c, "Settings retrieved", settings)
}

type UpdateSettingsInput struct {
	Name            *string  `json:"name"`
	RegistrationNo  *string  `json:"registration_no"`
	Address         *string  `json:"address"`
	WardNo          *int     `json:"ward_no"`
	Municipality    *string  `json:"municipality"`
	District        *string  `json:"district"`
	Province        *string  `json:"province"`
	ContactPhone    *string  `json:"contact_phone"`
	ContactEmail    *string  `json:"contact_email"`
	Description     *string  `json:"description"`
	Logo            *string  `json:"logo"`
	MapImage        *string  `json:"map_image"`
	Latitude        *float64 `json:"latitude"`
	Longitude       *float64 `json:"longitude"`
	EstablishedDate *string  `json:"established_date"`
}

func (h *SamitiHandler) UpdateSettings(c *gin.Context) {
	var input UpdateSettingsInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}

	before, _ := h.service.GetSettings()
	settings, err := h.service.UpdateSettings(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "update", "samiti_settings", &settings.ID, before, settings, c.ClientIP(), c.Request.UserAgent(), "Organization settings updated")
	response.Success(c, "Settings updated successfully", settings)
}

// ==================== Samiti Heads ====================

type CreateHeadInput struct {
	Name                 string  `json:"name" binding:"required"`
	Post                 string  `json:"post" binding:"required,oneof=chairperson secretary treasurer member"`
	Phone                *string `json:"phone"`
	Email                *string `json:"email"`
	Address              *string `json:"address"`
	Photo                *string `json:"photo"`
	TenureStart          *string `json:"tenure_start"`
	TenureEnd            *string `json:"tenure_end"`
	IsActive             *bool   `json:"is_active"`
	Remarks              *string `json:"remarks"`
	CreateLogin          bool    `json:"create_login"`
	AccountRole          string  `json:"account_role"`
	AccountPassword      string  `json:"account_password"`
	CurrentAdminPassword string  `json:"current_admin_password"`
	AdminMFACode         string  `json:"admin_mfa_code"`
}

type UpdateHeadInput struct {
	Name        *string `json:"name"`
	Post        *string `json:"post"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Address     *string `json:"address"`
	Photo       *string `json:"photo"`
	TenureStart *string `json:"tenure_start"`
	TenureEnd   *string `json:"tenure_end"`
	IsActive    *bool   `json:"is_active"`
	Remarks     *string `json:"remarks"`
}

func (h *SamitiHandler) CreateHead(c *gin.Context) {
	var input CreateHeadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}

	if input.CreateLogin {
		if err := auth.NewAuthService(h.service.db).VerifyPrivilegedStepUp(
			middleware.GetUserID(c), input.CurrentAdminPassword, input.AdminMFACode,
		); err != nil {
			response.Forbidden(c, err.Error())
			return
		}
	}

	head, err := h.service.CreateHead(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "create", "committee_head", &head.ID, nil, head, c.ClientIP(), c.Request.UserAgent(), "Committee member created; linked login may have been created or promoted")
	response.Created(c, "Committee member added successfully", head)
}

func (h *SamitiHandler) ListHeads(c *gin.Context) {
	heads, err := h.service.ListHeads()
	if err != nil {
		response.InternalError(c, "Failed to fetch committee members")
		return
	}
	response.Success(c, "Committee members retrieved", heads)
}

func (h *SamitiHandler) ListHeadsForAdmin(c *gin.Context) {
	heads, err := h.service.ListHeadsForAdmin()
	if err != nil {
		response.InternalError(c, "Failed to fetch committee members")
		return
	}
	response.Success(c, "Committee members retrieved", heads)
}

func (h *SamitiHandler) GetHeadForAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	head, err := h.service.GetHeadForAdmin(uint(id))
	if err != nil {
		response.NotFound(c, "Committee member not found")
		return
	}
	response.Success(c, "Committee member retrieved", head)
}

func (h *SamitiHandler) GetHeadByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	head, err := h.service.GetHeadByID(uint(id))
	if err != nil {
		response.NotFound(c, "Committee member not found")
		return
	}
	response.Success(c, "Committee member retrieved", head)
}

func (h *SamitiHandler) UpdateHead(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}

	before, _ := h.service.GetHeadForAdmin(uint(id))

	var input UpdateHeadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}

	head, err := h.service.UpdateHead(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "update", "committee_head", &head.ID, before, head, c.ClientIP(), c.Request.UserAgent(), "Committee member updated")
	response.Success(c, "Committee member updated", head)
}

func (h *SamitiHandler) DeleteHead(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	before, _ := h.service.GetHeadForAdmin(uint(id))
	if err := h.service.DeleteHead(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	actorID := middleware.GetUserID(c)
	entityID := uint(id)
	audit.CreateAuditEntry(h.service.db, &actorID, "deactivate", "committee_head", &entityID, before, nil, c.ClientIP(), c.Request.UserAgent(), "Committee member and linked login deactivated; history preserved")
	response.Success(c, "Committee member deactivated", nil)
}

// ==================== Logo Upload ====================

func (h *SamitiHandler) UploadLogo(c *gin.Context) {
	file, header, err := c.Request.FormFile("logo")
	if err != nil {
		response.BadRequest(c, "Logo file is required")
		return
	}
	defer file.Close()

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/gif":     true,
		"image/webp":    true,
		"image/svg+xml": true,
	}
	mimeType := header.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		response.BadRequest(c, "Only image files (JPEG, PNG, GIF, WEBP, SVG) are allowed")
		return
	}

	// Max 2MB for logo
	if header.Size > 2*1024*1024 {
		response.BadRequest(c, "Logo size must be less than 2MB")
		return
	}

	logoURL, err := h.service.UploadLogo(file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	audit.CreateAuditEntry(h.service.db, &actorID, "upload_logo", "samiti_settings", nil, nil, gin.H{"logo_url": logoURL}, c.ClientIP(), c.Request.UserAgent(), "Organization logo replaced")
	response.Success(c, "Logo uploaded successfully", gin.H{
		"logo_url": logoURL,
	})
}

// ==================== Head Photo Upload ====================

func (h *SamitiHandler) UploadHeadPhoto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid head ID")
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

	photoURL, err := h.service.UploadHeadPhoto(uint(id), file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	entityID := uint(id)
	audit.CreateAuditEntry(h.service.db, &actorID, "upload_photo", "committee_head", &entityID, nil, gin.H{"photo_url": photoURL}, c.ClientIP(), c.Request.UserAgent(), "Committee photograph replaced")
	response.Success(c, "Photo uploaded successfully", gin.H{
		"photo_url": photoURL,
	})
}
