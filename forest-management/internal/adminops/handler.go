package adminops

import (
	"fmt"
	"os"
	"strings"
	"time"

	"forest-management/internal/models"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	service *Service
	db      *gorm.DB
}

type exportRequest struct {
	Dataset         string `json:"dataset"`
	FiscalYearID    uint   `json:"fiscal_year_id"`
	FromDate        string `json:"from_date"`
	ToDate          string `json:"to_date"`
	CurrentPassword string `json:"current_password" binding:"required"`
	MFACode         string `json:"mfa_code"`
	Passphrase      string `json:"passphrase"`
}

type backupRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	MFACode         string `json:"mfa_code"`
	Passphrase      string `json:"passphrase" binding:"required"`
}

func NewHandler(service *Service, db *gorm.DB) *Handler { return &Handler{service: service, db: db} }

func (h *Handler) Datasets(c *gin.Context) {
	response.Success(c, "Available exports", gin.H{"datasets": DatasetNames()})
}

func (h *Handler) ExportCSV(c *gin.Context) {
	var req exportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dataset and current administrator password are required")
		return
	}
	if err := h.service.VerifyAdminCredentials(middleware.GetUserID(c), req.CurrentPassword, req.MFACode); err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	definition, ok := datasets[req.Dataset]
	if !ok {
		response.BadRequest(c, "Unsupported export dataset")
		return
	}
	filename := strings.TrimSuffix(definition.Filename, ".csv") + "-" + time.Now().UTC().Format("20060102-150405") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if err := h.service.WriteDatasetCSV(c.Writer, req.Dataset, ExportFilter{FiscalYearID: req.FiscalYearID, FromDate: req.FromDate, ToDate: req.ToDate}); err != nil {
		if !c.Writer.Written() {
			response.BadRequest(c, err.Error())
		}
		return
	}
	h.audit(c, "export_csv", req.Dataset, "Administrative CSV export")
}

func (h *Handler) ExportAll(c *gin.Context) {
	var req exportRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Passphrase) < 16 {
		response.BadRequest(c, "Current password and a backup passphrase of at least 16 characters are required")
		return
	}
	if err := h.service.VerifyAdminCredentials(middleware.GetUserID(c), req.CurrentPassword, req.MFACode); err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	path, filename, err := h.service.CreateAllDataExport(ExportFilter{FiscalYearID: req.FiscalYearID, FromDate: req.FromDate, ToDate: req.ToDate}, req.Passphrase)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	defer os.Remove(path)
	h.audit(c, "export_all", "data", "Encrypted full data export")
	downloadFile(c, path, filename)
}

func (h *Handler) DatabaseBackup(c *gin.Context) {
	h.handleBackup(c, false)
}

func (h *Handler) FullBackup(c *gin.Context) {
	h.handleBackup(c, true)
}

func (h *Handler) handleBackup(c *gin.Context, includeUploads bool) {
	var req backupRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Passphrase) < 16 {
		response.BadRequest(c, "Current password and a backup passphrase of at least 16 characters are required")
		return
	}
	if err := h.service.VerifyAdminCredentials(middleware.GetUserID(c), req.CurrentPassword, req.MFACode); err != nil {
		response.Forbidden(c, err.Error())
		return
	}
	var path, filename string
	var err error
	if includeUploads {
		path, filename, err = h.service.CreateFullBackup(req.Passphrase)
	} else {
		path, filename, err = h.service.CreateDatabaseBackup(req.Passphrase)
	}
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	defer os.Remove(path)
	action := "database_backup"
	remarks := "Encrypted PostgreSQL custom-format backup"
	if includeUploads {
		action = "full_backup"
		remarks = "Encrypted database and upload-file backup"
	}
	h.audit(c, action, "system", remarks)
	downloadFile(c, path, filename)
}

func (h *Handler) audit(c *gin.Context, action, entity, remarks string) {
	userID := middleware.GetUserID(c)
	ip, userAgent := c.ClientIP(), c.Request.UserAgent()
	_ = h.db.Create(&models.AuditLog{UserID: &userID, Action: action, Entity: entity, IPAddress: &ip, UserAgent: &userAgent, Remarks: &remarks}).Error
}

func downloadFile(c *gin.Context, path, filename string) {
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(path)
}
