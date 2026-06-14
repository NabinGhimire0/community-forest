package fiscalyears

import (
	"errors"
	"fmt"
	"time"

	"forest-management/internal/membershipfees"
	"forest-management/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// Update updates a fiscal year. Activation must use SetActive so stock carry-forward cannot be bypassed.
func (s *FiscalYearService) Update(id uint, input UpdateFiscalYearInput) (*models.FiscalYear, error) {
	var fy models.FiscalYear
	if err := s.db.First(&fy, id).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	name := fy.Name
	startDate := fy.StartDate
	endDate := fy.EndDate
	if input.Name != "" {
		var count int64
		s.db.Model(&models.FiscalYear{}).Where("name = ? AND id != ?", input.Name, id).Count(&count)
		if count > 0 {
			return nil, errors.New("fiscal year with this name already exists")
		}
		name = input.Name
	}
	if input.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return nil, errors.New("invalid start date format")
		}
		startDate = parsed
	}
	if input.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return nil, errors.New("invalid end date format")
		}
		endDate = parsed
	}
	if endDate.Before(startDate) {
		return nil, errors.New("end date must be after start date")
	}
	if input.IsActive != nil && *input.IsActive != fy.IsActive {
		return nil, errors.New("use the Set Active action to change the active fiscal year")
	}

	if err := s.db.Model(&fy).Updates(map[string]interface{}{
		"name": name, "start_date": startDate, "end_date": endDate,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to update: %w", err)
	}
	s.db.First(&fy, id)
	return &fy, nil
}

// Delete removes only a completely unused, inactive fiscal year. Official
// financial and stock history is never cascaded or silently destroyed.
func (s *FiscalYearService) Delete(id uint) error {
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, id).Error; err != nil {
		return errors.New("fiscal year not found")
	}
	if fiscalYear.IsActive {
		return errors.New("the active fiscal year cannot be deleted")
	}

	checks := []struct {
		model   interface{}
		message string
	}{
		{&models.FeeSetting{}, "cannot delete: fiscal year has fee settings"},
		{&models.Stock{}, "cannot delete: fiscal year has stock entries"},
		{&models.ResourceRate{}, "cannot delete: fiscal year has resource rates"},
		{&models.Transaction{}, "cannot delete: fiscal year has ledger transactions"},
		{&models.Request{}, "cannot delete: fiscal year has resource requests"},
		{&models.Expense{}, "cannot delete: fiscal year has expenses"},
		{&models.Fine{}, "cannot delete: fiscal year has fines"},
	}
	for _, check := range checks {
		var count int64
		if err := s.db.Unscoped().Model(check.model).Where("fiscal_year_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New(check.message)
		}
	}
	return s.db.Delete(&fiscalYear).Error
}

// SetActive activates a fiscal year and carries the previous year's remaining stock,
// resource rates and membership fee into missing rows in the target year. Existing
// target-year rows are never overwritten, so the operation is safe to retry.
func (s *FiscalYearService) SetActive(id uint) (*models.FiscalYear, error) {
	var target models.FiscalYear
	if err := s.db.First(&target, id).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	// Serialize fiscal-year activation so concurrent administrators cannot race
	// stock rollover or leave more than one active year.
	if err := tx.Exec("LOCK TABLE fiscal_years IN SHARE ROW EXCLUSIVE MODE").Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	defer func() {
		if recoverValue := recover(); recoverValue != nil {
			tx.Rollback()
			panic(recoverValue)
		}
	}()

	var source models.FiscalYear
	hasSource := false
	if err := tx.Where("is_active = ? AND id != ?", true, id).First(&source).Error; err == nil {
		hasSource = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, err
	}
	if !hasSource {
		if err := tx.Where("start_date < ? AND id != ?", target.StartDate, id).
			Order("start_date DESC").First(&source).Error; err == nil {
			hasSource = true
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, err
		}
	}

	carriedStock, carriedRates, carriedFees := 0, 0, 0
	if hasSource {
		var stocks []models.Stock
		if err := tx.Where("fiscal_year_id = ?", source.ID).Find(&stocks).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for _, old := range stocks {
			available := old.RemainingQuantity - old.ReservedQuantity
			if available < 0 {
				available = 0
			}
			row := models.Stock{ResourceItemID: old.ResourceItemID, FiscalYearID: target.ID, TotalQuantity: available, RemainingQuantity: available, ReservedQuantity: 0}
			result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "resource_item_id"}, {Name: "fiscal_year_id"}}, DoNothing: true}).Create(&row)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			if result.RowsAffected > 0 {
				carriedStock++
			}
		}

		var rates []models.ResourceRate
		if err := tx.Where("fiscal_year_id = ?", source.ID).Find(&rates).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for _, old := range rates {
			row := models.ResourceRate{ResourceItemID: old.ResourceItemID, FiscalYearID: target.ID, RatePerUnit: old.RatePerUnit}
			result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "resource_item_id"}, {Name: "fiscal_year_id"}}, DoNothing: true}).Create(&row)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			if result.RowsAffected > 0 {
				carriedRates++
			}
		}

		var fee models.FeeSetting
		if err := tx.Where("fiscal_year_id = ?", source.ID).First(&fee).Error; err == nil {
			row := models.FeeSetting{FiscalYearID: target.ID, MembershipFee: fee.MembershipFee}
			result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "fiscal_year_id"}}, DoNothing: true}).Create(&row)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			if result.RowsAffected > 0 {
				carriedFees++
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			return nil, err
		}
	}

	// A fiscal year cannot become active without a configured membership fee.
	// The setting may already exist or may have been carried from the previous year.
	var targetFee models.FeeSetting
	if err := tx.Where("fiscal_year_id = ?", target.ID).First(&targetFee).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("set the Gasti/Membership fee before activating this fiscal year")
		}
		return nil, err
	}

	assignedMemberFees, err := membershipfees.AssignForFiscalYear(tx, target, targetFee.MembershipFee)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to assign annual membership fees: %w", err)
	}

	if err := tx.Model(&models.FiscalYear{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Model(&models.FiscalYear{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	if err := s.db.First(&target, id).Error; err != nil {
		return nil, err
	}
	target.CarriedStockItems = carriedStock
	target.CarriedRateItems = carriedRates
	target.CarriedFeeItems = carriedFees
	target.AssignedMemberFees = assignedMemberFees
	return &target, nil
}

// SetFee sets (or updates) the membership fee for a fiscal year. When the
// fiscal year is active, missing member charges are created immediately and
// completely unpaid auto-generated charges are synchronized to the new amount.
func (s *FiscalYearService) SetFee(input SetFeeInput) (*models.FeeSetting, error) {
	if input.MembershipFee <= 0 {
		return nil, errors.New("membership fee must be greater than zero")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var fiscalYear models.FiscalYear
	if err := tx.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("fiscal year not found")
	}

	var fee models.FeeSetting
	result := tx.Where("fiscal_year_id = ?", input.FiscalYearID).First(&fee)
	if result.Error == nil {
		if err := tx.Model(&fee).Update("membership_fee", input.MembershipFee).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fee = models.FeeSetting{FiscalYearID: input.FiscalYearID, MembershipFee: input.MembershipFee}
		if err := tx.Create(&fee).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		tx.Rollback()
		return nil, result.Error
	}

	if fiscalYear.IsActive {
		if err := membershipfees.SyncUnpaidAmount(tx, fiscalYear.ID, input.MembershipFee); err != nil {
			tx.Rollback()
			return nil, err
		}
		if _, err := membershipfees.AssignForFiscalYear(tx, fiscalYear, input.MembershipFee); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	s.db.First(&fee, fee.ID)
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
	return s.SetFee(SetFeeInput{FiscalYearID: fee.FiscalYearID, MembershipFee: input.MembershipFee})
}

// DeleteFee removes only an unused fee setting. Once annual charges exist,
// changing the amount must use UpdateFee so paid history remains traceable.
func (s *FiscalYearService) DeleteFee(id uint) error {
	var fee models.FeeSetting
	if err := s.db.Preload("FiscalYear").First(&fee, id).Error; err != nil {
		return errors.New("fee setting not found")
	}
	if fee.FiscalYear != nil && fee.FiscalYear.IsActive {
		return errors.New("the active fiscal year's fee cannot be deleted")
	}
	var ledgerCount int64
	if err := s.db.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type IN ?", fee.FiscalYearID, []string{"membership_fee", "legacy_gasti_fee"}).
		Count(&ledgerCount).Error; err != nil {
		return err
	}
	if ledgerCount > 0 {
		return errors.New("cannot delete: membership-fee ledger entries already exist")
	}
	return s.db.Delete(&fee).Error
}
