package resources

import (
	"errors"
	"fmt"

	"forest-management/internal/models"

	"gorm.io/gorm"
)

type ResourceService struct {
	db *gorm.DB
}

func NewResourceService(db *gorm.DB) *ResourceService {
	return &ResourceService{db: db}
}

// ==================== Resource Types ====================

func (s *ResourceService) CreateType(input CreateResourceTypeInput) (*models.ResourceType, error) {
	// Check if type already exists
	var existing models.ResourceType
	if s.db.Where("name = ?", input.Name).First(&existing).Error == nil {
		return nil, errors.New("resource type with this name already exists")
	}

	resourceType := models.ResourceType{
		Name: input.Name,
		Unit: input.Unit,
	}
	if err := s.db.Create(&resourceType).Error; err != nil {
		return nil, fmt.Errorf("failed to create resource type: %w", err)
	}
	return &resourceType, nil
}

func (s *ResourceService) ListTypes() ([]models.ResourceType, error) {
	var types []models.ResourceType
	err := s.db.Preload("Items").Order("name ASC").Find(&types).Error
	return types, err
}

func (s *ResourceService) GetTypeByID(id uint) (*models.ResourceType, error) {
	var resourceType models.ResourceType
	err := s.db.Preload("Items").First(&resourceType, id).Error
	return &resourceType, err
}

func (s *ResourceService) UpdateType(id uint, input UpdateResourceTypeInput) (*models.ResourceType, error) {
	var resourceType models.ResourceType
	if err := s.db.First(&resourceType, id).Error; err != nil {
		return nil, errors.New("resource type not found")
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		// Check unique name
		var existing models.ResourceType
		if s.db.Where("name = ? AND id != ?", input.Name, id).First(&existing).Error == nil {
			return nil, errors.New("resource type with this name already exists")
		}
		updates["name"] = input.Name
	}
	if input.Unit != "" {
		updates["unit"] = input.Unit
	}

	if len(updates) > 0 {
		if err := s.db.Model(&resourceType).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
	}

	s.db.First(&resourceType, id)
	return &resourceType, nil
}

func (s *ResourceService) DeleteType(id uint) error {
	// Check if any items exist under this type
	var count int64
	s.db.Model(&models.ResourceItem{}).Where("resource_type_id = ?", id).Count(&count)
	if count > 0 {
		return errors.New("cannot delete: resource type has items. Delete items first")
	}
	return s.db.Delete(&models.ResourceType{}, id).Error
}

// ==================== Resource Items ====================

func (s *ResourceService) CreateItem(input CreateResourceItemInput) (*models.ResourceItem, error) {
	// Validate type exists
	var resourceType models.ResourceType
	if err := s.db.First(&resourceType, input.ResourceTypeID).Error; err != nil {
		return nil, errors.New("resource type not found")
	}

	// Check if item already exists for this type
	var existing models.ResourceItem
	if s.db.Where("resource_type_id = ? AND name = ?", input.ResourceTypeID, input.Name).First(&existing).Error == nil {
		return nil, errors.New("item already exists in this resource type")
	}

	item := models.ResourceItem{
		ResourceTypeID: input.ResourceTypeID,
		Name:           input.Name,
	}
	if err := s.db.Create(&item).Error; err != nil {
		return nil, err
	}

	s.db.Preload("Type").First(&item, item.ID)
	return &item, nil
}

func (s *ResourceService) ListItems(typeID string) ([]models.ResourceItem, error) {
	var items []models.ResourceItem
	query := s.db.Model(&models.ResourceItem{})
	if typeID != "" {
		query = query.Where("resource_type_id = ?", typeID)
	}
	err := query.Preload("Type").Order("name ASC").Find(&items).Error
	return items, err
}

func (s *ResourceService) GetItemByID(id uint) (*models.ResourceItem, error) {
	var item models.ResourceItem
	err := s.db.Preload("Type").First(&item, id).Error
	return &item, err
}

func (s *ResourceService) UpdateItem(id uint, input UpdateResourceItemInput) (*models.ResourceItem, error) {
	var item models.ResourceItem
	if err := s.db.First(&item, id).Error; err != nil {
		return nil, errors.New("resource item not found")
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.ResourceTypeID != 0 {
		// Validate type exists
		var resourceType models.ResourceType
		if err := s.db.First(&resourceType, input.ResourceTypeID).Error; err != nil {
			return nil, errors.New("resource type not found")
		}
		updates["resource_type_id"] = input.ResourceTypeID
	}

	if len(updates) > 0 {
		if err := s.db.Model(&item).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
	}

	s.db.Preload("Type").First(&item, id)
	return &item, nil
}

func (s *ResourceService) DeleteItem(id uint) error {
	// Check if has rates
	var rateCount int64
	s.db.Model(&models.ResourceRate{}).Where("resource_item_id = ?", id).Count(&rateCount)
	if rateCount > 0 {
		return errors.New("cannot delete: item has rate settings")
	}

	// Check if has stock
	var stockCount int64
	s.db.Model(&models.Stock{}).Where("resource_item_id = ?", id).Count(&stockCount)
	if stockCount > 0 {
		return errors.New("cannot delete: item has stock entries")
	}

	return s.db.Delete(&models.ResourceItem{}, id).Error
}

// ==================== Resource Rates ====================

func (s *ResourceService) SetRate(input SetRateInput) (*models.ResourceRate, error) {
	// Validate item exists
	var item models.ResourceItem
	if err := s.db.First(&item, input.ResourceItemID).Error; err != nil {
		return nil, errors.New("resource item not found")
	}

	// Validate fiscal year exists
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	// Upsert: find existing or create new
	var existingRate models.ResourceRate
	result := s.db.Where(
		"resource_item_id = ? AND fiscal_year_id = ?",
		input.ResourceItemID, input.FiscalYearID,
	).First(&existingRate)

	if result.Error == nil {
		// Update existing
		if err := s.db.Model(&existingRate).Update("rate_per_unit", input.RatePerUnit).Error; err != nil {
			return nil, err
		}
		s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&existingRate, existingRate.ID)
		return &existingRate, nil
	}

	// Create new
	rate := models.ResourceRate{
		ResourceItemID: input.ResourceItemID,
		FiscalYearID:   input.FiscalYearID,
		RatePerUnit:    input.RatePerUnit,
	}
	if err := s.db.Create(&rate).Error; err != nil {
		return nil, err
	}

	s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&rate, rate.ID)
	return &rate, nil
}

func (s *ResourceService) ListRates(fiscalYearID string) ([]models.ResourceRate, error) {
	var rates []models.ResourceRate
	query := s.db.Model(&models.ResourceRate{})
	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	err := query.Preload("Item").Preload("Item.Type").Preload("FiscalYear").Order("id DESC").Find(&rates).Error
	return rates, err
}

func (s *ResourceService) GetRateByID(id uint) (*models.ResourceRate, error) {
	var rate models.ResourceRate
	err := s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&rate, id).Error
	return &rate, err
}

func (s *ResourceService) UpdateRate(id uint, input UpdateRateInput) (*models.ResourceRate, error) {
	var rate models.ResourceRate
	if err := s.db.First(&rate, id).Error; err != nil {
		return nil, errors.New("rate not found")
	}

	if err := s.db.Model(&rate).Update("rate_per_unit", input.RatePerUnit).Error; err != nil {
		return nil, err
	}

	s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&rate, id)
	return &rate, nil
}

func (s *ResourceService) DeleteRate(id uint) error {
	return s.db.Delete(&models.ResourceRate{}, id).Error
}

// ==================== Stock ====================

func (s *ResourceService) UpdateStock(input UpdateStockInput) (*models.Stock, error) {
	// Validate item exists
	var item models.ResourceItem
	if err := s.db.First(&item, input.ResourceItemID).Error; err != nil {
		return nil, errors.New("resource item not found")
	}

	// Validate fiscal year exists
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, input.FiscalYearID).Error; err != nil {
		return nil, errors.New("fiscal year not found")
	}

	var existing models.Stock
	result := s.db.Where(
		"resource_item_id = ? AND fiscal_year_id = ?",
		input.ResourceItemID, input.FiscalYearID,
	).First(&existing)

	if result.Error == nil {
		// Calculate how much has been used/sold
		used := existing.TotalQuantity - existing.RemainingQuantity
		newRemaining := input.TotalQuantity - used
		if newRemaining < 0 {
			newRemaining = 0
		}

		if err := s.db.Model(&existing).Updates(map[string]interface{}{
			"total_quantity":     input.TotalQuantity,
			"remaining_quantity": newRemaining,
		}).Error; err != nil {
			return nil, err
		}
		s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&existing, existing.ID)
		return &existing, nil
	}

	// Create new stock entry
	stock := models.Stock{
		ResourceItemID:    input.ResourceItemID,
		FiscalYearID:      input.FiscalYearID,
		TotalQuantity:     input.TotalQuantity,
		RemainingQuantity: input.TotalQuantity,
	}
	if err := s.db.Create(&stock).Error; err != nil {
		return nil, err
	}

	s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&stock, stock.ID)
	return &stock, nil
}

func (s *ResourceService) ListStock(fiscalYearID string) ([]models.Stock, error) {
	var stocks []models.Stock
	query := s.db.Model(&models.Stock{})
	if fiscalYearID != "" {
		query = query.Where("fiscal_year_id = ?", fiscalYearID)
	}
	err := query.Preload("Item").Preload("Item.Type").Preload("FiscalYear").Order("id DESC").Find(&stocks).Error
	return stocks, err
}

func (s *ResourceService) GetStockByID(id uint) (*models.Stock, error) {
	var stock models.Stock
	err := s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&stock, id).Error
	return &stock, err
}

func (s *ResourceService) UpdateStockQuantity(id uint, input UpdateStockQuantityInput) (*models.Stock, error) {
	var stock models.Stock
	if err := s.db.First(&stock, id).Error; err != nil {
		return nil, errors.New("stock not found")
	}

	// Calculate remaining quantity based on new total
	used := stock.TotalQuantity - stock.RemainingQuantity
	newRemaining := input.TotalQuantity - used
	if newRemaining < 0 {
		newRemaining = 0
	}

	if err := s.db.Model(&stock).Updates(map[string]interface{}{
		"total_quantity":     input.TotalQuantity,
		"remaining_quantity": newRemaining,
	}).Error; err != nil {
		return nil, err
	}

	s.db.Preload("Item").Preload("Item.Type").Preload("FiscalYear").First(&stock, id)
	return &stock, nil
}

func (s *ResourceService) DeleteStock(id uint) error {
	return s.db.Delete(&models.Stock{}, id).Error
}
