package membershipfee

import (
	"strconv"

	"forest-management/internal/models"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type MembershipFeeHandler struct {
	service *MembershipFeeService
}

func NewMembershipFeeHandler(service *MembershipFeeService) *MembershipFeeHandler {
	return &MembershipFeeHandler{service: service}
}

// Collect — Record a membership fee payment
func (h *MembershipFeeHandler) Collect(c *gin.Context) {
	adminUserID := middleware.GetUserID(c)

	var input CollectFeeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data: "+err.Error())
		return
	}

	result, err := h.service.CollectFee(adminUserID, input)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Membership fee collected", result)
}

// BulkCollect — Collect fees for all unpaid members
func (h *MembershipFeeHandler) BulkCollect(c *gin.Context) {
	adminUserID := middleware.GetUserID(c)

	var input struct {
		FiscalYearID  uint   `json:"fiscal_year_id" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid data")
		return
	}

	results, err := h.service.BulkCollectFee(adminUserID, input.FiscalYearID, input.PaymentMethod)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Created(c, "Bulk fee collection completed", gin.H{
		"collected_count": len(results),
		"results":         results,
	})
}

// MemberFeeStatus — Get fee status for a specific member
func (h *MembershipFeeHandler) MemberFeeStatus(c *gin.Context) {
	memberID, _ := strconv.Atoi(c.Param("id"))

	status, err := h.service.GetMemberFeeStatus(uint(memberID))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Fee status retrieved", status)
}

// MyFeeStatus — Member views their own fee status
func (h *MembershipFeeHandler) MyFeeStatus(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var member models.Member
	if err := h.service.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		response.NotFound(c, "Member profile not found")
		return
	}

	status, err := h.service.GetMemberFeeStatus(member.ID)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Your fee status", status)
}

// ListCollections — Admin views all fee collections
func (h *MembershipFeeHandler) ListCollections(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))
	fiscalYearID := c.Query("fiscal_year_id")
	memberID := c.Query("member_id")

	txn, meta, err := h.service.ListFeeCollections(page, perPage, fiscalYearID, memberID)
	if err != nil {
		response.InternalError(c, "Failed to fetch fee collections")
		return
	}

	var pagination *response.Pagination
	if meta != nil {
		p := response.Pagination(*meta)
		pagination = &p
	}

	response.Paginated(c, "Fee collections retrieved", txn, pagination)
}
