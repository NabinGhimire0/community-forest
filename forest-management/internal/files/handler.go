package files

import (
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"forest-management/config"
	"forest-management/internal/models"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db        *gorm.DB
	uploadDir string
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db, uploadDir: config.AppConfig.UploadDir}
}

func RegisterRoutes(router *gin.Engine, handler *Handler) {
	// The logo and elected committee photographs are intentionally public.
	router.GET("/uploads/samiti/:filename", handler.ServePublic)
	router.GET("/uploads/heads/:filename", handler.ServePublic)

	private := router.Group("/uploads")
	private.Use(middleware.AuthMiddleware())
	private.GET("/:folder/:filename", handler.ServePrivate)
}

func (h *Handler) ServePublic(c *gin.Context) {
	folder := c.Param("folder")
	if folder == "" {
		// Gin does not include a named folder in the specific routes, derive it.
		parts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			folder = parts[1]
		}
	}
	if folder != "samiti" && folder != "heads" {
		response.NotFound(c, "File not found")
		return
	}
	path, err := h.safePath(folder, c.Param("filename"))
	if err != nil || !isPublicImage(path) {
		response.NotFound(c, "File not found")
		return
	}
	c.Header("Cache-Control", "public, max-age=86400")
	h.serve(c, path)
}

func (h *Handler) ServePrivate(c *gin.Context) {
	folder, filename := c.Param("folder"), c.Param("filename")
	if folder == "samiti" || folder == "heads" {
		h.ServePublic(c)
		return
	}
	path, err := h.safePath(folder, filename)
	if err != nil {
		response.NotFound(c, "File not found")
		return
	}
	role := middleware.GetUserRole(c)
	if role != "admin" && role != "staff" {
		allowed, err := h.memberMayAccess(middleware.GetUserID(c), folder, filename)
		if err != nil || !allowed {
			response.Forbidden(c, "You do not have access to this document")
			return
		}
	}
	c.Header("Cache-Control", "private, no-store")
	h.auditPrivateAccess(c, folder, filename)
	h.serve(c, path)
}

func (h *Handler) memberMayAccess(userID uint, folder, filename string) (bool, error) {
	var member models.Member
	if err := h.db.Select("id", "photo", "assistant_photo").Where("user_id = ?", userID).First(&member).Error; err != nil {
		return false, err
	}
	legacyURL := fmt.Sprintf("/uploads/%s/%s", folder, filename)
	if folder == "members" {
		return (member.Photo != nil && *member.Photo == legacyURL) || (member.AssistantPhoto != nil && *member.AssistantPhoto == legacyURL), nil
	}
	if folder == "fines" {
		var count int64
		err := h.db.Model(&models.Fine{}).Where("member_id = ? AND photo = ?", member.ID, legacyURL).Count(&count).Error
		return count > 0, err
	}

	var upload models.FileUpload
	if err := h.db.Where("folder = ? AND stored_name = ?", folder, filename).First(&upload).Error; err != nil {
		return false, nil
	}
	if upload.Visibility == "public" {
		return true, nil
	}
	if upload.Entity == nil || upload.EntityID == nil {
		return false, nil
	}
	switch *upload.Entity {
	case "member":
		return *upload.EntityID == member.ID, nil
	case "transaction":
		var count int64
		err := h.db.Model(&models.Transaction{}).Where("id = ? AND member_id = ?", *upload.EntityID, member.ID).Count(&count).Error
		return count > 0, err
	case "request":
		var count int64
		err := h.db.Model(&models.Request{}).Where("id = ? AND member_id = ?", *upload.EntityID, member.ID).Count(&count).Error
		return count > 0, err
	case "payment":
		var count int64
		err := h.db.Model(&models.Payment{}).Where("id = ? AND member_id = ?", *upload.EntityID, member.ID).Count(&count).Error
		return count > 0, err
	case "fine":
		var count int64
		err := h.db.Model(&models.Fine{}).Where("id = ? AND member_id = ?", *upload.EntityID, member.ID).Count(&count).Error
		return count > 0, err
	default:
		return false, nil
	}
}

func (h *Handler) safePath(folder, filename string) (string, error) {
	if filepath.Base(folder) != folder || filepath.Base(filename) != filename || folder == "." || filename == "." {
		return "", fmt.Errorf("invalid path")
	}
	root, err := filepath.Abs(h.uploadDir)
	if err != nil {
		return "", err
	}
	candidate, err := filepath.Abs(filepath.Join(root, folder, filename))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path")
	}
	info, err := os.Stat(candidate)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("file not found")
	}
	return candidate, nil
}

func (h *Handler) auditPrivateAccess(c *gin.Context, folder, filename string) {
	userID := middleware.GetUserID(c)
	var upload models.FileUpload
	var entityID *uint
	if err := h.db.Select("id").Where("folder = ? AND stored_name = ?", folder, filename).First(&upload).Error; err == nil {
		id := upload.ID
		entityID = &id
	}
	ip, userAgent := c.ClientIP(), c.Request.UserAgent()
	remarks := fmt.Sprintf("Authorized private file access: %s/%s", folder, filename)
	_ = h.db.Create(&models.AuditLog{
		UserID: &userID, Action: "private_file_access", Entity: "file_upload", EntityID: entityID,
		IPAddress: &ip, UserAgent: &userAgent, Remarks: &remarks,
	}).Error
}

func (h *Handler) serve(c *gin.Context, path string) {
	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(path)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("X-Content-Type-Options", "nosniff")
	disposition := "attachment"
	if strings.HasPrefix(contentType, "image/") {
		disposition = "inline"
	}
	c.Header("Content-Disposition", fmt.Sprintf("%s; filename=%q", disposition, filepath.Base(path)))
	http.ServeFile(c.Writer, c.Request, path)
}

func isPublicImage(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return extension == ".jpg" || extension == ".jpeg" || extension == ".png" || extension == ".webp"
}
