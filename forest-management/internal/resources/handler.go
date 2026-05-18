package resources

import (
	"strconv"

	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	service *ResourceService
}

func NewResourceHandler(service *ResourceService) *ResourceHandler {
	return &ResourceHandler{service: service}
}

// ==================== Resource Types ====================

type CreateResourceTypeInput struct {
	Name string `json:"name" binding:"required"`
	Unit string `json:"unit" binding:"required"`
}

type UpdateResourceTypeInput struct {
	Name string `json:"name"`
	Unit string `json:"unit"`
}

func (h *ResourceHandler) CreateType(c *gin.Context) {
	var input CreateResourceTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}
	resourceType, err := h.service.CreateType(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Created(c, "Resource type created", resourceType)
}

func (h *ResourceHandler) ListTypes(c *gin.Context) {
	types, err := h.service.ListTypes()
	if err != nil {
		response.InternalError(c, "Failed to fetch resource types")
		return
	}
	response.Success(c, "Resource types retrieved", types)
}

func (h *ResourceHandler) GetTypeByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	resourceType, err := h.service.GetTypeByID(uint(id))
	if err != nil {
		response.NotFound(c, "Resource type not found")
		return
	}
	response.Success(c, "Resource type retrieved", resourceType)
}

func (h *ResourceHandler) UpdateType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	var input UpdateResourceTypeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}
	resourceType, err := h.service.UpdateType(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Resource type updated", resourceType)
}

func (h *ResourceHandler) DeleteType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	if err := h.service.DeleteType(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Resource type deleted", nil)
}

// ==================== Resource Items ====================

type CreateResourceItemInput struct {
	ResourceTypeID uint   `json:"resource_type_id" binding:"required"`
	Name           string `json:"name" binding:"required"`
}

type UpdateResourceItemInput struct {
	ResourceTypeID uint   `json:"resource_type_id"`
	Name           string `json:"name"`
}

func (h *ResourceHandler) CreateItem(c *gin.Context) {
	var input CreateResourceItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}
	item, err := h.service.CreateItem(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Created(c, "Resource item created", item)
}

func (h *ResourceHandler) ListItems(c *gin.Context) {
	typeID := c.Query("type_id")
	items, err := h.service.ListItems(typeID)
	if err != nil {
		response.InternalError(c, "Failed to fetch resource items")
		return
	}
	response.Success(c, "Resource items retrieved", items)
}

func (h *ResourceHandler) GetItemByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	item, err := h.service.GetItemByID(uint(id))
	if err != nil {
		response.NotFound(c, "Resource item not found")
		return
	}
	response.Success(c, "Resource item retrieved", item)
}

func (h *ResourceHandler) UpdateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	var input UpdateResourceItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}
	item, err := h.service.UpdateItem(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Resource item updated", item)
}

func (h *ResourceHandler) DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	if err := h.service.DeleteItem(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Resource item deleted", nil)
}

// ==================== Resource Rates ====================

type SetRateInput struct {
	ResourceItemID uint    `json:"resource_item_id" binding:"required"`
	FiscalYearID   uint    `json:"fiscal_year_id" binding:"required"`
	RatePerUnit    float64 `json:"rate_per_unit" binding:"required"`
}

type UpdateRateInput struct {
	RatePerUnit float64 `json:"rate_per_unit" binding:"required"`
}

func (h *ResourceHandler) SetRate(c *gin.Context) {
	var input SetRateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}
	rate, err := h.service.SetRate(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Created(c, "Rate set successfully", rate)
}

func (h *ResourceHandler) ListRates(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	rates, err := h.service.ListRates(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to fetch rates")
		return
	}
	response.Success(c, "Rates retrieved", rates)
}

func (h *ResourceHandler) GetRateByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	rate, err := h.service.GetRateByID(uint(id))
	if err != nil {
		response.NotFound(c, "Rate not found")
		return
	}
	response.Success(c, "Rate retrieved", rate)
}

func (h *ResourceHandler) UpdateRate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	var input UpdateRateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}
	rate, err := h.service.UpdateRate(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Rate updated", rate)
}

func (h *ResourceHandler) DeleteRate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	if err := h.service.DeleteRate(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Rate deleted", nil)
}

// ==================== Stock ====================

type UpdateStockInput struct {
	ResourceItemID uint    `json:"resource_item_id" binding:"required"`
	FiscalYearID   uint    `json:"fiscal_year_id" binding:"required"`
	TotalQuantity  float64 `json:"total_quantity" binding:"required"`
}

type UpdateStockQuantityInput struct {
	TotalQuantity float64 `json:"total_quantity" binding:"required"`
}

func (h *ResourceHandler) UpdateStock(c *gin.Context) {
	var input UpdateStockInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}
	stock, err := h.service.UpdateStock(input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Stock updated", stock)
}

func (h *ResourceHandler) ListStock(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	stocks, err := h.service.ListStock(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to fetch stock")
		return
	}
	response.Success(c, "Stock retrieved", stocks)
}

func (h *ResourceHandler) GetStockByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	stock, err := h.service.GetStockByID(uint(id))
	if err != nil {
		response.NotFound(c, "Stock not found")
		return
	}
	response.Success(c, "Stock retrieved", stock)
}

func (h *ResourceHandler) UpdateStockQuantity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	var input UpdateStockQuantityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}
	stock, err := h.service.UpdateStockQuantity(uint(id), input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Stock quantity updated", stock)
}

func (h *ResourceHandler) DeleteStock(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid ID")
		return
	}
	if err := h.service.DeleteStock(uint(id)); err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.Success(c, "Stock deleted", nil)
}
