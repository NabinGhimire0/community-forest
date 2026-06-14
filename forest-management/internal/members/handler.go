package members

import (
	"errors"
	"strconv"

	"forest-management/internal/audit"
	"forest-management/internal/auth"
	"forest-management/pkg/middleware"
	"forest-management/pkg/requestutil"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	service *MemberService
}

func NewMemberHandler(service *MemberService) *MemberHandler {
	return &MemberHandler{service: service}
}

// CreateMemberRequest — DTO for creating a member
type CreateMemberRequest struct {
	MembershipNo   string                `json:"membership_no"`
	Name           string                `json:"name" binding:"required"`
	AssistantName  string                `json:"assistant_name" binding:"required"`
	FatherName     string                `json:"father_name" binding:"required"`
	WardNo         int                   `json:"ward_no" binding:"required"`
	Tole           string                `json:"tole" binding:"required"`
	Phone          *string               `json:"phone" binding:"required"`
	Photo          *string               `json:"photo"`
	AssistantPhoto *string               `json:"assistant_photo"`
	JoinedDate     *string               `json:"joined_date"`
	Remarks        *string               `json:"remarks"`
	FamilyMembers  []FamilyMemberRequest `json:"family_members"`
}

// UpdateMemberRequest — DTO for updating a member
type UpdateMemberRequest struct {
	MembershipNo   string                `json:"membership_no"`
	Name           string                `json:"name" binding:"required"`
	AssistantName  string                `json:"assistant_name" binding:"required"`
	FatherName     string                `json:"father_name" binding:"required"`
	WardNo         int                   `json:"ward_no" binding:"required"`
	Tole           string                `json:"tole" binding:"required"`
	Phone          *string               `json:"phone"`
	Photo          *string               `json:"photo"`
	AssistantPhoto *string               `json:"assistant_photo"`
	Status         *string               `json:"status"`
	Remarks        *string               `json:"remarks"`
	FamilyMembers  []FamilyMemberRequest `json:"family_members"`
}

// AddFamilyMemberRequest — DTO for adding a family member
type FamilyMemberRequest struct {
	Name          string  `json:"name" binding:"required"`
	Relation      string  `json:"relation" binding:"required"`
	Age           *int    `json:"age"`
	Gender        *string `json:"gender"`
	CitizenshipNo *string `json:"citizenship_no"`
	Remarks       *string `json:"remarks"`
}

// Create handles POST /api/members
func (h *MemberHandler) Create(c *gin.Context) {
	var req CreateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	member, credentials, err := h.service.CreateMember(req, &actorID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Member created. The temporary password is displayed once; deliver it through a trusted channel.", gin.H{
		"member":                   member,
		"credentials_sent":         false,
		"phone":                    credentials.Phone,
		"temporary_password":       credentials.PlainPassword,
		"password_must_change":     true,
		"credential_handover_note": "Do not send this password through public or shared channels.",
	})
}

// List handles GET /api/members with pagination and search
func (h *MemberHandler) List(c *gin.Context) {
	page, perPage := requestutil.Pagination(c)
	search := c.Query("search")
	status := c.Query("status")

	members, meta, err := h.service.ListMembers(page, perPage, search, status)
	if err != nil {
		response.InternalError(c, "Failed to fetch members")
		return
	}

	response.Paginated(c, "Members retrieved", members, meta)
}

// GetByID handles GET /api/members/:id
func (h *MemberHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	member, err := h.service.GetMemberByID(uint(id))
	if err != nil {
		response.NotFound(c, "Member not found")
		return
	}

	response.Success(c, "Member retrieved", member)
}

// Update handles PUT /api/members/:id
func (h *MemberHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	var req UpdateMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	uid := middleware.GetUserID(c)
	var updater *uint
	if uid != 0 {
		updater = &uid
	}

	member, err := h.service.UpdateMember(uint(id), req, updater)
	if err != nil {
		if errors.Is(err, ErrMemberUpdateForbidden) {
			response.Forbidden(c, err.Error())
			return
		}
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Member updated", member)
}

// Delete handles DELETE /api/members/:id
func (h *MemberHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	uid := middleware.GetUserID(c)
	var actor *uint
	if uid != 0 {
		actor = &uid
	}
	if err := h.service.DeleteMember(uint(id), actor); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Member deactivated; historical records were preserved", nil)
}

// GetProfile handles GET /api/members/profile — member's own profile
func (h *MemberHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	member, err := h.service.GetMemberByUserID(userID)
	if err != nil {
		response.NotFound(c, "Member profile not found")
		return
	}

	response.Success(c, "Profile retrieved", member)
}

// AddFamilyMember handles POST /api/members/:id/family
func (h *MemberHandler) AddFamilyMember(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	var req FamilyMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data")
		return
	}

	familyMember, err := h.service.AddFamilyMember(uint(memberID), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Family member added", familyMember)
}

// ListFamilyMembers handles GET /api/members/:id/family
func (h *MemberHandler) ListFamilyMembers(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	family, err := h.service.ListFamilyMembers(uint(memberID))
	if err != nil {
		response.InternalError(c, "Failed to fetch family members")
		return
	}

	response.Success(c, "Family members retrieved", family)
}

// UpdateFamilyMember handles PUT /api/members/:id/family/:familyId
func (h *MemberHandler) UpdateFamilyMember(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	familyID, err := strconv.Atoi(c.Param("familyId"))
	if err != nil {
		response.BadRequest(c, "Invalid family member ID")
		return
	}

	var req FamilyMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	familyMember, err := h.service.UpdateFamilyMember(uint(memberID), uint(familyID), req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Family member updated", familyMember)
}

// DeleteFamilyMember handles DELETE /api/members/:id/family/:familyId
func (h *MemberHandler) DeleteFamilyMember(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	familyID, err := strconv.Atoi(c.Param("familyId"))
	if err != nil {
		response.BadRequest(c, "Invalid family member ID")
		return
	}

	if err := h.service.DeleteFamilyMember(uint(memberID), uint(familyID)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, "Family member deleted", nil)
}

type ResetCredentialsRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	MFACode         string `json:"mfa_code"`
}

// ResetCredentials handles POST /api/members/:id/reset-credentials. Password
// resets are step-up protected because they grant control of another account.
func (h *MemberHandler) ResetCredentials(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}
	var req ResetCredentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Current administrator password is required")
		return
	}
	actorID := middleware.GetUserID(c)
	if err := auth.NewAuthService(h.service.db).VerifyPrivilegedStepUp(actorID, req.CurrentPassword, req.MFACode); err != nil {
		response.Forbidden(c, err.Error())
		return
	}

	credentials, err := h.service.ResetCredentials(uint(memberID))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	entityID := uint(memberID)
	audit.CreateAuditEntry(h.service.db, &actorID, "reset_credentials", "member", &entityID, nil, nil, c.ClientIP(), c.Request.UserAgent(), "Administrator generated a one-time temporary password; target sessions revoked")
	response.Success(c, "A new temporary password was generated and all existing sessions were revoked.", gin.H{
		"phone":                credentials.Phone,
		"temporary_password":   credentials.PlainPassword,
		"credentials_sent":     false,
		"password_must_change": true,
	})
}

// BulkImport handles CSV file upload for bulk member creation
func (h *MemberHandler) BulkImport(c *gin.Context) {
	actorID := middleware.GetUserID(c)
	if err := auth.NewAuthService(h.service.db).VerifyPrivilegedStepUp(
		actorID, c.PostForm("current_password"), c.PostForm("mfa_code"),
	); err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "CSV file is required")
		return
	}
	defer file.Close()

	bulkService := NewBulkImportService(h.service.db)
	result, err := bulkService.ImportFromCSV(file)
	if err != nil {
		response.Error(c, 500, "Failed to import CSV: "+err.Error())
		return
	}

	audit.CreateAuditEntry(h.service.db, &actorID, "bulk_import", "member", nil, nil, gin.H{"success_count": result.SuccessCount, "error_count": result.ErrorCount}, c.ClientIP(), c.Request.UserAgent(), "Administrator imported members; temporary passwords shown once")
	response.Success(c, "Bulk import completed. Temporary passwords are shown once; store and deliver them securely.", result)
}

// DownloadImportTemplate provides a CSV template for bulk import
func (h *MemberHandler) DownloadImportTemplate(c *gin.Context) {
	template := GenerateCSVTemplate()
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=member_import_template.csv")
	c.Data(200, "text/csv", template)
}

// UploadPhoto handles member photo upload
func (h *MemberHandler) UploadPhoto(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
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

	photoURL, err := h.service.UploadMemberPhoto(uint(memberID), file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Photo uploaded successfully", gin.H{
		"photo_url": photoURL,
	})
}

// UploadAssistantPhoto handles member assistant photo upload
func (h *MemberHandler) UploadAssistantPhoto(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
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

	photoURL, err := h.service.UploadAssistantPhoto(uint(memberID), file, header.Filename)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Assistant photo uploaded successfully", gin.H{
		"photo_url": photoURL,
	})
}

// GetMemberFeeDetails - Get membership fee details
func (h *MemberHandler) GetMemberFeeDetails(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	details, err := h.service.GetMemberFeeDetails(uint(memberID))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Fee details retrieved", details)
}

// GetMemberSalesDetails - Get sales details
func (h *MemberHandler) GetMemberSalesDetails(c *gin.Context) {
	memberID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid member ID")
		return
	}

	details, err := h.service.GetMemberSalesDetails(uint(memberID))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Sales details retrieved", details)
}
