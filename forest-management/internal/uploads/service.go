package uploads

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"forest-management/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadService struct {
	db        *gorm.DB
	uploadDir string
	baseURL   string
}

func NewUploadService(db *gorm.DB) *UploadService {
	uploadDir := "./uploads"
	baseURL := "/api/uploads/files"

	// Create folder structure
	folders := []string{"photos", "documents", "bills", "receipts", "profiles", "misc"}
	for _, folder := range folders {
		os.MkdirAll(filepath.Join(uploadDir, folder), os.ModePerm)
	}

	return &UploadService{db: db, uploadDir: uploadDir, baseURL: baseURL}
}

var allowedMimeTypes = map[string]bool{
	"image/jpeg":         true,
	"image/png":          true,
	"image/gif":          true,
	"image/webp":         true,
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":       true,
	"text/csv": true,
}

var maxFileSize int64 = 10 * 1024 * 1024 // 10MB

// UploadFile handles a single file upload
func (s *UploadService) UploadFile(fileHeader io.Reader, originalName, mimeType string, fileSize int64, folder string, userID uint, entity *string, entityID *uint) (*models.FileUpload, error) {
	// Validate MIME type
	if !allowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("file type %s is not allowed", mimeType)
	}

	// Validate file size
	if fileSize > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit of 10MB")
	}

	// Validate folder
	validFolders := map[string]bool{
		"photos": true, "documents": true, "bills": true,
		"receipts": true, "profiles": true, "misc": true,
	}
	if !validFolders[folder] {
		folder = "misc"
	}

	// Generate unique filename
	ext := filepath.Ext(originalName)
	storedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(s.uploadDir, folder, storedName)
	fileURL := fmt.Sprintf("%s/%s/%s", s.baseURL, folder, storedName)

	// Create the file on disk
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file to the destination
	limitedReader := io.LimitReader(fileHeader, maxFileSize)
	written, err := io.Copy(dst, limitedReader)
	if err != nil {
		os.Remove(filePath) // Cleanup on failure
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Create database record
	upload := models.FileUpload{
		OriginalName: originalName,
		StoredName:   storedName,
		FilePath:     filePath,
		FileURL:      fileURL,
		MimeType:     mimeType,
		FileSize:     written,
		Folder:       folder,
		Entity:       entity,
		EntityID:     entityID,
		UploadedBy:   userID,
	}

	if err := s.db.Create(&upload).Error; err != nil {
		os.Remove(filePath) // Cleanup on failure
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}

	s.db.Preload("Uploader").First(&upload, upload.ID)
	return &upload, nil
}

// DeleteFile removes a file and its database record
func (s *UploadService) DeleteFile(fileID uint) error {
	var upload models.FileUpload
	if err := s.db.First(&upload, fileID).Error; err != nil {
		return fmt.Errorf("file not found")
	}

	// Remove from disk
	os.Remove(upload.FilePath)

	// Remove from database
	return s.db.Delete(&upload).Error
}

// ListFiles returns files with optional filters
func (s *UploadService) ListFiles(folder, entity string, entityID *uint) ([]models.FileUpload, error) {
	var files []models.FileUpload
	query := s.db.Model(&models.FileUpload{})

	if folder != "" {
		query = query.Where("folder = ?", folder)
	}
	if entity != "" {
		query = query.Where("entity = ?", entity)
	}
	if entityID != nil {
		query = query.Where("entity_id = ?", entityID)
	}

	err := query.Preload("Uploader").Order("created_at DESC").Find(&files).Error
	return files, err
}

// GetFilePath returns the actual file path for serving
func (s *UploadService) GetFilePath(folder, filename string) (string, error) {
	filePath := filepath.Join(s.uploadDir, folder, filename)

	// Security: verify file exists and is within upload directory
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("invalid file path")
	}

	uploadDirAbs, _ := filepath.Abs(s.uploadDir)
	if !strings.HasPrefix(absPath, uploadDirAbs) {
		return "", fmt.Errorf("access denied")
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found")
	}

	return absPath, nil
}
