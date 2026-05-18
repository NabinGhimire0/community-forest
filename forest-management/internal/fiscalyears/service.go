package fiscalyears

import (
	"errors"
	"fmt"
	"time"

	"forest-management/internal/models"

	"gorm.io/gorm"
)

type FiscalYearService struct {
	db *gorm.DB
}

func NewFiscalYearService(db *gorm.DB) *FiscalYearService {
	return &FiscalYearService{db: db}
}

// Create creates a new fiscal year
func (s *FiscalYearService) Create(input CreateFiscalYearInput) (*models.FiscalYear, error) {
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format. Use YYYY-MM-DD")
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return nil, errors.New("invalid end date format. Use YYYY-MM-DD")
	}
	if endDate.Before(startDate) {
		return nil, errors.New("end date must be after start date")
	}

	// Check unique name
	var existing models.FiscalYear
	if s.db.Where("name = ?", input.Name).First(&existing).Error == nil {
		return nil, errors.New("fiscal year with this name already exists")
	}

	fy := models.FiscalYear{
		Name:      input.Name,
		StartDate: startDate,
		EndDate:   endDate,
		IsActive:  false,
	}

	if err := s.db.Create(&fy).Error; err != nil {
		return nil, fmt.Errorf("failed to create fiscal year: %w", err)
	}

	return &fy, nil
}

// List returns all fiscal years ordered by date
func (s *FiscalYearService) List() ([]models.FiscalYear, error) {
	var fyList []models.FiscalYear
	err := s.db.Order("start_date DESC").Find(&fyList).Error
	return fyList, err
}

// GetByID returns a single fiscal year
func (s *FiscalYearService) GetByID(id uint) (*models.FiscalYear, error) {
	var fy models.FiscalYear
	err := s.db.First(&fy, id).Error
	return &fy, err
}

// Update updates a fiscal year
func (s *FiscalYearService) Update(id uint, input UpdateFiscalYearInput) (*models.FiscalYear, error) {
	var fy models.FiscalYear
	if err := s.db.First(&fy, id).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	updates := make(map[string]interface{})

	if input.Name != "" {
		var existing models.FiscalYear
		if s.db.Where("name = ? AND id != ?", input.Name, id).First(&existing).Error == nil {
			return nil, errors.New("fiscal year with this name already exists")
		}
		updates["name"] = input.Name
	}
	if input.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return nil, errors.New("invalid start date format")
		}
		updates["start_date"] = startDate
	}
	if input.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return nil, errors.New("invalid end date format")
		}
		updates["end_date"] = endDate
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	if len(updates) > 0 {
		if err := s.db.Model(&fy).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
	}

	s.db.First(&fy, id)
	return &fy, nil
}

// Delete deletes a fiscal year
func (s *FiscalYearService) Delete(id uint) error {
	// Check if has any associated data
	var feeCount int64
	s.db.Model(&models.FeeSetting{}).Where("fiscal_year_id = ?", id).Count(&feeCount)
	if feeCount > 0 {
		return errors.New("cannot delete: fiscal year has fee settings")
	}

	var stockCount int64
	s.db.Model(&models.Stock{}).Where("fiscal_year_id = ?", id).Count(&stockCount)
	if stockCount > 0 {
		return errors.New("cannot delete: fiscal year has stock entries")
	}

	var rateCount int64
	s.db.Model(&models.ResourceRate{}).Where("fiscal_year_id = ?", id).Count(&rateCount)
	if rateCount > 0 {
		return errors.New("cannot delete: fiscal year has resource rates")
	}

	return s.db.Delete(&models.FiscalYear{}, id).Error
}

// SetActive deactivates all fiscal years and activates the specified one
func (s *FiscalYearService) SetActive(id uint) (*models.FiscalYear, error) {
	var fy models.FiscalYear
	if err := s.db.First(&fy, id).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	tx := s.db.Begin()

	// Deactivate ALL fiscal years
	if err := tx.Model(&models.FiscalYear{}).Where("is_active = ?", true).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Activate the selected one
	if err := tx.Model(&fy).Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	s.db.First(&fy, id)
	return &fy, nil
}

// SetFee sets (or updates) the membership fee for a fiscal year
func (s *FiscalYearService) SetFee(input SetFeeInput) (*models.FeeSetting, error) {
	// Validate fiscal year exists
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	var existing models.FeeSetting
	result := s.db.Where("fiscal_year_id = ?", input.FiscalYearID).First(&existing)

	if result.Error == nil {
		// Update existing
		if err := s.db.Model(&existing).Update("membership_fee", input.MembershipFee).Error; err != nil {
			return nil, err
		}
		s.db.Preload("FiscalYear").First(&existing, existing.ID)
		return &existing, nil
	}

	// Create new
	fee := models.FeeSetting{
		FiscalYearID:  input.FiscalYearID,
		MembershipFee: input.MembershipFee,
	}
	if err := s.db.Create(&fee).Error; err != nil {
		return nil, err
	}
	s.db.Preload("FiscalYear").First(&fee, fee.ID)
	return &fee, nil
}

// GetFee returns the membership fee for a fiscal year
func (s *FiscalYearService) GetFee(fiscalYearID string) (*models.FeeSetting, error) {
	var fee models.FeeSetting
	err := s.db.Preload("FiscalYear").
		Where("fiscal_year_id = ?", fiscalYearID).First(&fee).Error
	return &fee, err
}

// GetFeesByFiscalYear returns all fee settings for a fiscal year
func (s *FiscalYearService) GetFeesByFiscalYear(fiscalYearID uint) ([]models.FeeSetting, error) {
	var fees []models.FeeSetting
	err := s.db.Where("fiscal_year_id = ?", fiscalYearID).Find(&fees).Error
	return fees, err
}

// UpdateFee updates a fee setting
func (s *FiscalYearService) UpdateFee(id uint, input UpdateFeeInput) (*models.FeeSetting, error) {
	var fee models.FeeSetting
	if err := s.db.First(&fee, id).Error; err != nil {
		return nil, errors.New("fee setting not found")
	}

	if err := s.db.Model(&fee).Update("membership_fee", input.MembershipFee).Error; err != nil {
		return nil, err
	}

	s.db.First(&fee, id)
	return &fee, nil
}

// DeleteFee deletes a fee setting
func (s *FiscalYearService) DeleteFee(id uint) error {
	return s.db.Delete(&models.FeeSetting{}, id).Error
}
