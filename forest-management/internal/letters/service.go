package letters

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"forest-management/internal/models"
	"forest-management/pkg/response"

	"gorm.io/gorm"
)

type LetterService struct {
	db *gorm.DB
}

func NewLetterService(db *gorm.DB) *LetterService {
	return &LetterService{db: db}
}

func (s *LetterService) CreateLetter(userID uint, input CreateLetterInput) (*models.Letter, error) {
	letterDate, err := time.Parse("2006-01-02", input.LetterDate)
	if err != nil {
		return nil, errors.New("invalid letter date format. Use YYYY-MM-DD")
	}

	var receivedDate, sentDate *time.Time
	if input.ReceivedDate != nil && *input.ReceivedDate != "" {
		t, err := time.Parse("2006-01-02", *input.ReceivedDate)
		if err == nil {
			receivedDate = &t
		}
	}
	if input.SentDate != nil && *input.SentDate != "" {
		t, err := time.Parse("2006-01-02", *input.SentDate)
		if err == nil {
			sentDate = &t
		}
	}

	letter := models.Letter{
		Type:         input.Type,
		ReferenceNo:  input.ReferenceNo,
		Title:        input.Title,
		Subject:      input.Subject,
		FromParty:    input.FromParty,
		ToParty:      input.ToParty,
		LetterDate:   letterDate,
		ReceivedDate: receivedDate,
		SentDate:     sentDate,
		DocumentFile: input.DocumentFile,
		Remarks:      input.Remarks,
		CreatedBy:    userID,
	}

	if err := s.db.Create(&letter).Error; err != nil {
		return nil, fmt.Errorf("failed to create letter: %w", err)
	}

	s.db.Preload("Creator").First(&letter, letter.ID)
	return &letter, nil
}

func (s *LetterService) ListLetters(page, perPage int, letterType, search string) ([]models.Letter, *response.Pagination, error) {
	var letters []models.Letter
	var total int64

	query := s.db.Model(&models.Letter{})

	if letterType != "" {
		query = query.Where("type = ?", letterType)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"title ILIKE ? OR subject ILIKE ? OR reference_no ILIKE ? OR from_party ILIKE ? OR to_party ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	query.Count(&total)
	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("Creator").
		Order("letter_date DESC").
		Offset(offset).
		Limit(perPage).
		Find(&letters).Error

	return letters, &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

func (s *LetterService) GetLetterByID(id uint) (*models.Letter, error) {
	var letter models.Letter
	err := s.db.Preload("Creator").First(&letter, id).Error
	return &letter, err
}

func (s *LetterService) UpdateLetter(id uint, userID uint, input UpdateLetterInput) (*models.Letter, error) {
	var letter models.Letter
	if err := s.db.First(&letter, id).Error; err != nil {
		return nil, errors.New("letter not found")
	}

	updates := make(map[string]interface{})

	if input.Type != nil {
		updates["type"] = *input.Type
	}
	if input.ReferenceNo != nil {
		updates["reference_no"] = *input.ReferenceNo
	}
	if input.Title != nil {
		updates["title"] = *input.Title
	}
	if input.Subject != nil {
		updates["subject"] = *input.Subject
	}
	if input.FromParty != nil {
		updates["from_party"] = *input.FromParty
	}
	if input.ToParty != nil {
		updates["to_party"] = *input.ToParty
	}
	if input.LetterDate != nil {
		letterDate, err := time.Parse("2006-01-02", *input.LetterDate)
		if err == nil {
			updates["letter_date"] = letterDate
		}
	}
	if input.ReceivedDate != nil {
		var receivedDate *time.Time
		if *input.ReceivedDate != "" {
			t, err := time.Parse("2006-01-02", *input.ReceivedDate)
			if err == nil {
				receivedDate = &t
			}
		}
		updates["received_date"] = receivedDate
	}
	if input.SentDate != nil {
		var sentDate *time.Time
		if *input.SentDate != "" {
			t, err := time.Parse("2006-01-02", *input.SentDate)
			if err == nil {
				sentDate = &t
			}
		}
		updates["sent_date"] = sentDate
	}
	if input.DocumentFile != nil {
		updates["document_file"] = *input.DocumentFile
	}
	if input.Remarks != nil {
		updates["remarks"] = *input.Remarks
	}

	if len(updates) > 0 {
		if err := s.db.Model(&letter).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update letter: %w", err)
		}
	}

	s.db.Preload("Creator").First(&letter, id)
	return &letter, nil
}

func (s *LetterService) DeleteLetter(id uint) error {
	var letter models.Letter
	if err := s.db.First(&letter, id).Error; err != nil {
		return errors.New("letter not found")
	}

	// Delete associated document file if exists
	if letter.DocumentFile != nil && *letter.DocumentFile != "" {
		// Extract filename from URL
		filePath := "." + *letter.DocumentFile
		os.Remove(filePath)
	}

	return s.db.Delete(&letter).Error
}

// UploadDocument saves the document file and returns the URL
func (s *LetterService) UploadDocument(letterID uint, file io.Reader, filename string) (string, error) {
	var letter models.Letter
	if err := s.db.First(&letter, letterID).Error; err != nil {
		return "", errors.New("letter not found")
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/letters"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	uniqueName := fmt.Sprintf("letter_%d_%d%s", letterID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	fileURL := fmt.Sprintf("/uploads/letters/%s", uniqueName)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Update letter with document URL
	if err := s.db.Model(&letter).Update("document_file", fileURL).Error; err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to update letter: %w", err)
	}

	return fileURL, nil
}
