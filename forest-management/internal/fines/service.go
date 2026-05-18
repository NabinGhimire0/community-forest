package fines

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"forest-management/internal/models"
	"forest-management/internal/notifications"
	"forest-management/pkg/response"

	"gorm.io/gorm"
)

type FineService struct {
	db *gorm.DB
}

func NewFineService(db *gorm.DB) *FineService {
	return &FineService{db: db}
}

func (s *FineService) CreateFine(adminUserID uint, input CreateFineInput) (*models.Fine, error) {
	incidentDate, err := time.Parse("2006-01-02", input.IncidentDate)
	if err != nil {
		return nil, errors.New("invalid incident date format. Use YYYY-MM-DD")
	}

	// Validate fiscal year exists
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	// If member_id is provided, validate member exists
	var memberName string
	if input.MemberID != nil {
		var member models.Member
		if err := s.db.First(&member, *input.MemberID).Error; err != nil {
			return nil, errors.New("member not found")
		}
		memberName = member.Name
	} else {
		memberName = input.Name
	}

	fine := models.Fine{
		FiscalYearID:  input.FiscalYearID,
		MemberID:      input.MemberID,
		Name:          memberName,
		ViolationType: input.ViolationType,
		Description:   input.Description,
		FineAmount:    input.FineAmount,
		IncidentDate:  incidentDate,
		Status:        "pending",
		Photo:         input.Photo,
		Remarks:       input.Remarks,
		CreatedBy:     adminUserID,
	}

	if err := s.db.Create(&fine).Error; err != nil {
		return nil, fmt.Errorf("failed to create fine: %w", err)
	}

	s.db.Preload("Member").Preload("Member.User").Preload("FiscalYear").Preload("Creator").First(&fine, fine.ID)

	// Notify member if applicable
	if input.MemberID != nil {
		notifService := notifications.NewNotificationService(s.db)
		var member models.Member
		s.db.Preload("User").First(&member, *input.MemberID)
		if member.UserID != nil {
			notifService.NotifyUser(
				*member.UserID,
				"Fine Issued",
				fmt.Sprintf("A fine of Rs. %.2f has been issued for %s. Please pay at the office.", input.FineAmount, input.ViolationType),
				"warning",
				stringPtr("fine"),
				&fine.ID,
			)
		}
	}

	return &fine, nil
}

func (s *FineService) ListFines(page, perPage int, status, fiscalYearID, memberID, search string) ([]models.Fine, *response.Pagination, error) {
	var fines []models.Fine
	var total int64

	query := s.db.Model(&models.Fine{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	if memberID != "" {
		query = query.Where("member_id = ?", memberID)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"name ILIKE ? OR violation_type ILIKE ?",
			searchPattern, searchPattern,
		)
	}

	query.Count(&total)
	offset := (page - 1) * perPage
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	err := query.
		Preload("Member").
		Preload("Member.User").
		Preload("FiscalYear").
		Preload("Creator").
		Order("incident_date DESC").
		Offset(offset).
		Limit(perPage).
		Find(&fines).Error

	return fines, &response.Pagination{
		Page: page, PerPage: perPage, Total: total, TotalPages: totalPages,
	}, err
}

func (s *FineService) GetFineByID(id uint) (*models.Fine, error) {
	var fine models.Fine
	err := s.db.Preload("Member").Preload("Member.User").Preload("FiscalYear").Preload("Creator").First(&fine, id).Error
	return &fine, err
}

func (s *FineService) UpdateFine(id uint, adminUserID uint, input UpdateFineInput) (*models.Fine, error) {
	var fine models.Fine
	if err := s.db.First(&fine, id).Error; err != nil {
		return nil, errors.New("fine not found")
	}

	updates := make(map[string]interface{})

	if input.ViolationType != nil {
		updates["violation_type"] = *input.ViolationType
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.FineAmount != nil {
		updates["fine_amount"] = *input.FineAmount
	}
	if input.IncidentDate != nil {
		incidentDate, err := time.Parse("2006-01-02", *input.IncidentDate)
		if err == nil {
			updates["incident_date"] = incidentDate
		}
	}
	if input.Photo != nil {
		updates["photo"] = *input.Photo
	}
	if input.Remarks != nil {
		updates["remarks"] = *input.Remarks
	}

	if len(updates) > 0 {
		if err := s.db.Model(&fine).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update fine: %w", err)
		}
	}

	s.db.Preload("Member").Preload("Member.User").Preload("FiscalYear").Preload("Creator").First(&fine, id)
	return &fine, nil
}

func (s *FineService) UpdateFineStatus(id uint, adminUserID uint, input UpdateFineStatusInput) (*models.Fine, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var fine models.Fine
	if err := tx.Preload("Member").Preload("Member.User").First(&fine, id).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("fine not found")
	}

	if fine.Status == "paid" {
		tx.Rollback()
		return nil, errors.New("fine already paid")
	}
	if fine.Status == "waived" {
		tx.Rollback()
		return nil, errors.New("fine already waived")
	}

	updates := map[string]interface{}{
		"status": input.Status,
	}

	if input.PaymentReference != nil {
		updates["payment_reference"] = *input.PaymentReference
	}

	if err := tx.Model(&fine).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update fine status: %w", err)
	}

	// If fine is paid, create a transaction record
	if input.Status == "paid" {
		now := time.Now()
		receiptNo := fmt.Sprintf("FINE-REC-%d-%d", now.Year(), fine.ID)

		transaction := models.Transaction{
			MemberID:        *fine.MemberID,
			FiscalYearID:    fine.FiscalYearID,
			ResourceItemID:  nil,
			Type:            "fine",
			Quantity:        nil,
			RatePerUnit:     nil,
			TotalAmount:     fine.FineAmount,
			AmountPaid:      fine.FineAmount,
			AmountRemaining: 0,
			ReceiptNo:       receiptNo,
			Date:            now,
			Remarks:         stringPtr(fmt.Sprintf("Fine payment for: %s", fine.ViolationType)),
		}

		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create transaction: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	s.db.Preload("Member").Preload("Member.User").Preload("FiscalYear").Preload("Creator").First(&fine, id)

	// Notify member
	if fine.MemberID != nil && fine.Member != nil && fine.Member.UserID != nil {
		notifService := notifications.NewNotificationService(s.db)
		var message string
		if input.Status == "paid" {
			message = fmt.Sprintf("Your fine of Rs. %.2f has been paid. Receipt: %s", fine.FineAmount, *input.PaymentReference)
		} else {
			message = fmt.Sprintf("Your fine of Rs. %.2f has been waived.", fine.FineAmount)
		}
		notifService.NotifyUser(
			*fine.Member.UserID,
			fmt.Sprintf("Fine %s", input.Status),
			message,
			"info",
			stringPtr("fine"),
			&fine.ID,
		)
	}

	return &fine, nil
}

func (s *FineService) DeleteFine(id uint) error {
	var fine models.Fine
	if err := s.db.First(&fine, id).Error; err != nil {
		return errors.New("fine not found")
	}

	// Delete associated photo if exists
	if fine.Photo != nil && *fine.Photo != "" {
		filePath := "." + *fine.Photo
		os.Remove(filePath)
	}

	return s.db.Delete(&fine).Error
}

func (s *FineService) GetFineStatistics(fiscalYearID string) (map[string]interface{}, error) {
	var stats struct {
		Total      int64   `json:"total"`
		Pending    int64   `json:"pending"`
		Paid       int64   `json:"paid"`
		Waived     int64   `json:"waived"`
		TotalValue float64 `json:"total_value"`
		PaidValue  float64 `json:"paid_value"`
	}

	query := s.db.Model(&models.Fine{})
	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}

	query.Count(&stats.Total)
	query.Where("status = ?", "pending").Count(&stats.Pending)
	query.Where("status = ?", "paid").Count(&stats.Paid)
	query.Where("status = ?", "waived").Count(&stats.Waived)
	query.Select("COALESCE(SUM(fine_amount), 0)").Scan(&stats.TotalValue)
	query.Where("status = ?", "paid").Select("COALESCE(SUM(fine_amount), 0)").Scan(&stats.PaidValue)

	return map[string]interface{}{
		"total":       stats.Total,
		"pending":     stats.Pending,
		"paid":        stats.Paid,
		"waived":      stats.Waived,
		"total_value": stats.TotalValue,
		"paid_value":  stats.PaidValue,
	}, nil
}

// UploadPhoto saves the fine photo and returns the URL
func (s *FineService) UploadPhoto(fineID uint, file io.Reader, filename string) (string, error) {
	var fine models.Fine
	if err := s.db.First(&fine, fineID).Error; err != nil {
		return "", errors.New("fine not found")
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/fines"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}
	uniqueName := fmt.Sprintf("fine_%d_%d%s", fineID, time.Now().UnixNano(), ext)
	filePath := filepath.Join(uploadDir, uniqueName)
	fileURL := fmt.Sprintf("/uploads/fines/%s", uniqueName)

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

	// Update fine with photo URL
	if err := s.db.Model(&fine).Update("photo", fileURL).Error; err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to update fine: %w", err)
	}

	return fileURL, nil
}

func stringPtr(s string) *string {
	return &s
}
