package uploads

import (
	"fmt"
	"strconv"

	"forest-management/internal/models"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	service *UploadService
}

func NewUploadHandler(service *UploadService) *UploadHandler {
	return &UploadHandler{service: service}
}

// Upload — Upload a single file
func (h *UploadHandler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "No file provided")
		return
	}
	defer file.Close()

	folder := c.DefaultPostForm("folder", "misc")
	entity := c.PostForm("entity")
	var entityID *uint
	if eid := c.PostForm("entity_id"); eid != "" {
		id, _ := strconv.Atoi(eid)
		entityID = uintPtr(uint(id))
	}

	var entityPtr *string
	if entity != "" {
		entityPtr = &entity
	}

	upload, err := h.service.UploadFile(file, header.Filename, header.Header.Get("Content-Type"), header.Size, folder, userID, entityPtr, entityID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "File uploaded", upload)
}

// UploadMultiple — Upload multiple files
func (h *UploadHandler) UploadMultiple(c *gin.Context) {
	userID := middleware.GetUserID(c)

	form, err := c.MultipartForm()
	if err != nil {
		response.BadRequest(c, "Invalid multipart form")
		return
	}

	files := form.File["files"]
	folder := c.DefaultPostForm("folder", "misc")
	entity := c.PostForm("entity")
	var entityID *uint
	if eid := c.PostForm("entity_id"); eid != "" {
		id, _ := strconv.Atoi(eid)
		entityID = uintPtr(uint(id))
	}

	var entityPtr *string
	if entity != "" {
		entityPtr = &entity
	}

	var uploads []models.FileUpload
	var errors []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to open %s: %v", fileHeader.Filename, err))
			continue
		}

		upload, err := h.service.UploadFile(file, fileHeader.Filename, fileHeader.Header.Get("Content-Type"), fileHeader.Size, folder, userID, entityPtr, entityID)
		file.Close()

		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to upload %s: %v", fileHeader.Filename, err))
			continue
		}

		uploads = append(uploads, *upload)
	}

	response.Success(c, "Files uploaded", gin.H{
		"uploaded": uploads,
		"count":    len(uploads),
		"errors":   errors,
	})
}

// List — List uploaded files
func (h *UploadHandler) List(c *gin.Context) {
	folder := c.Query("folder")
	entity := c.Query("entity")
	var entityID *uint
	if eid := c.Query("entity_id"); eid != "" {
		id, _ := strconv.Atoi(eid)
		entityID = uintPtr(uint(id))
	}

	files, err := h.service.ListFiles(folder, entity, entityID)
	if err != nil {
		response.InternalError(c, "Failed to fetch files")
		return
	}

	response.Success(c, "Files retrieved", files)
}

// Delete — Delete an uploaded file
func (h *UploadHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.DeleteFile(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "File deleted", nil)
}

// ServeFile — Serve a file for viewing/download
func (h *UploadHandler) ServeFile(c *gin.Context) {
	folder := c.Param("folder")
	filename := c.Param("filename")

	filePath, err := h.service.GetFilePath(folder, filename)
	if err != nil {
		response.NotFound(c, "File not found")
		return
	}

	c.File(filePath)
}

func uintPtr(u uint) *uint {
	return &u
}
