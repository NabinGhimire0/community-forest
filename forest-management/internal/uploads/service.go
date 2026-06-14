package uploads

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"forest-management/config"
	"forest-management/internal/models"
	"forest-management/pkg/fileutil"

	"gorm.io/gorm"
)

type UploadService struct {
	db        *gorm.DB
	uploadDir string
	baseURL   string
}

func NewUploadService(db *gorm.DB) *UploadService {
	uploadDir := config.AppConfig.UploadDir
	baseURL := "/uploads"
	for _, folder := range []string{"photos", "documents", "bills", "receipts", "profiles", "misc"} {
		_ = os.MkdirAll(filepath.Join(uploadDir, folder), 0o750)
	}
	return &UploadService{db: db, uploadDir: uploadDir, baseURL: baseURL}
}

const maxFileSize int64 = 10 * 1024 * 1024

func (s *UploadService) UploadFile(reader io.Reader, originalName, declaredMime string, fileSize int64, folder string, userID uint, entity *string, entityID *uint) (*models.FileUpload, error) {
	if fileSize <= 0 || fileSize > maxFileSize {
		return nil, fmt.Errorf("file size must be between 1 byte and 10MB")
	}
	if entity != nil && *entity == "transaction" && (entityID == nil || *entityID == 0) {
		return nil, fmt.Errorf("transaction document requires a valid transaction ID")
	}
	validFolders := map[string]bool{"photos": true, "documents": true, "bills": true, "receipts": true, "profiles": true, "misc": true}
	if !validFolders[folder] {
		folder = "misc"
	}

	policy := fileutil.EvidencePolicy
	if folder == "photos" || folder == "profiles" {
		policy = fileutil.ImagePolicy
	}
	saved, err := fileutil.Save(reader, folder, "upload", policy)
	if err != nil {
		return nil, err
	}

	visibility := "private"
	digest := saved.SHA256
	upload := models.FileUpload{
		OriginalName: filepath.Base(originalName),
		StoredName:   saved.StoredName,
		FilePath:     saved.Path,
		FileURL:      saved.URL,
		MimeType:     saved.MimeType,
		FileSize:     saved.Size,
		Folder:       folder,
		Entity:       entity,
		EntityID:     entityID,
		UploadedBy:   userID,
		Visibility:   visibility,
		SHA256:       &digest,
	}
	if err := s.db.Create(&upload).Error; err != nil {
		_ = os.Remove(saved.Path)
		return nil, fmt.Errorf("failed to save file record: %w", err)
	}
	s.db.Preload("Uploader").First(&upload, upload.ID)
	return &upload, nil
}

func (s *UploadService) DeleteFile(fileID uint) error {
	var upload models.FileUpload
	if err := s.db.First(&upload, fileID).Error; err != nil {
		return fmt.Errorf("file not found")
	}
	if err := s.db.Delete(&upload).Error; err != nil {
		return err
	}
	_ = os.Remove(upload.FilePath)
	return nil
}

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
	err := query.Preload("Uploader").Order("created_at DESC").Limit(1000).Find(&files).Error
	return files, err
}

func (s *UploadService) GetFilePath(folder, filename string) (string, error) {
	if filepath.Base(folder) != folder || filepath.Base(filename) != filename {
		return "", fmt.Errorf("invalid file path")
	}
	root, err := filepath.Abs(s.uploadDir)
	if err != nil {
		return "", fmt.Errorf("invalid upload directory")
	}
	candidate, err := filepath.Abs(filepath.Join(root, folder, filename))
	if err != nil {
		return "", fmt.Errorf("invalid file path")
	}
	rel, err := filepath.Rel(root, candidate)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("access denied")
	}
	if _, err := os.Stat(candidate); err != nil {
		return "", fmt.Errorf("file not found")
	}
	return candidate, nil
}
