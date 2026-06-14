package members

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"forest-management/internal/audit"
	"forest-management/internal/models"
	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	historicalFeeType          = "legacy_gasti_fee"
	historicalTimberSaleType   = "legacy_timber_sale"
	historicalFirewoodSaleType = "legacy_firewood_sale"
	historicalOtherSaleType    = "legacy_other_sale"
)

var historicalTypes = []string{
	historicalFeeType,
	historicalTimberSaleType,
	historicalFirewoodSaleType,
	historicalOtherSaleType,
}

var ErrHistoricalFeeAlreadyExists = errors.New("a past Gasti fee balance already exists for this member and fiscal year")

type CreateHistoricalTransactionRequest struct {
	Category          string  `json:"category" binding:"required"`
	FiscalYearID      uint    `json:"fiscal_year_id" binding:"required"`
	SaleType          string  `json:"sale_type"`
	AmountRemaining   float64 `json:"amount_remaining" binding:"required,gt=0"`
	RecordDate        *string `json:"record_date"`
	PhysicalReference *string `json:"physical_reference"`
	Remarks           *string `json:"remarks"`
	SaveAsDraft       bool    `json:"save_as_draft"`
}

type ReverseHistoricalTransactionRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func (h *MemberHandler) CreateHistoricalTransaction(c *gin.Context) {
	memberID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || memberID == 0 {
		response.BadRequest(c, "Invalid member ID")
		return
	}
	var req CreateHistoricalTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request data: "+err.Error())
		return
	}

	actorID := middleware.GetUserID(c)
	role := middleware.GetUserRole(c)
	// Staff can digitize records, but an admin must verify them.
	verified := role == "admin" && !req.SaveAsDraft
	transaction, err := h.service.CreateHistoricalTransaction(uint(memberID), actorID, verified, req)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			response.NotFound(c, "Member or fiscal year not found")
		case errors.Is(err, ErrHistoricalFeeAlreadyExists):
			response.Error(c, http.StatusConflict, err.Error())
		default:
			response.BadRequest(c, err.Error())
		}
		return
	}

	audit.CreateAuditEntry(h.service.db, &actorID, "create", "historical_transaction", &transaction.ID, nil, transaction, c.ClientIP(), c.Request.UserAgent(), "Historical balance digitized from physical register")
	message := "Historical balance saved as draft"
	if verified {
		message = "Historical balance recorded and verified"
	}
	response.Created(c, message, transaction)
}

func (h *MemberHandler) VerifyHistoricalTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("transactionId"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid transaction ID")
		return
	}
	actorID := middleware.GetUserID(c)
	before, after, err := h.service.VerifyHistoricalTransaction(uint(id), actorID)
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	audit.CreateAuditEntry(h.service.db, &actorID, "verify", "historical_transaction", &after.ID, before, after, c.ClientIP(), c.Request.UserAgent(), "Historical register entry verified")
	response.Success(c, "Historical balance verified", after)
}

func (h *MemberHandler) ReverseHistoricalTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("transactionId"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid transaction ID")
		return
	}
	var req ReverseHistoricalTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Reason) == "" {
		response.BadRequest(c, "A reversal reason is required")
		return
	}
	actorID := middleware.GetUserID(c)
	before, after, err := h.service.ReverseHistoricalTransaction(uint(id), actorID, strings.TrimSpace(req.Reason))
	if err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	audit.CreateAuditEntry(h.service.db, &actorID, "reverse", "historical_transaction", &after.ID, before, after, c.ClientIP(), c.Request.UserAgent(), req.Reason)
	response.Success(c, "Historical balance reversed", after)
}

func (s *MemberService) CreateHistoricalTransaction(memberID, actorID uint, verified bool, req CreateHistoricalTransactionRequest) (*models.Transaction, error) {
	var member models.Member
	if err := s.db.First(&member, memberID).Error; err != nil {
		return nil, err
	}
	var fiscalYear models.FiscalYear
	if err := s.db.First(&fiscalYear, req.FiscalYearID).Error; err != nil {
		return nil, err
	}

	transactionType, err := historicalTransactionType(req.Category, req.SaleType)
	if err != nil {
		return nil, err
	}
	if transactionType == historicalFeeType {
		var count int64
		if err := s.db.Model(&models.Transaction{}).
			Where("member_id = ? AND fiscal_year_id = ? AND type = ? AND record_status <> ?", memberID, req.FiscalYearID, transactionType, "reversed").
			Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, ErrHistoricalFeeAlreadyExists
		}
	}

	recordDate := fiscalYear.EndDate
	if req.RecordDate != nil && strings.TrimSpace(*req.RecordDate) != "" {
		parsed, parseErr := time.Parse("2006-01-02", strings.TrimSpace(*req.RecordDate))
		if parseErr != nil {
			return nil, errors.New("record_date must use YYYY-MM-DD format")
		}
		recordDate = parsed
	}
	status := "draft"
	var verifiedBy *uint
	var verifiedAt *time.Time
	if verified {
		status = "verified"
		verifiedBy = &actorID
		now := time.Now()
		verifiedAt = &now
	}

	transaction := models.Transaction{
		MemberID: memberID, FiscalYearID: req.FiscalYearID, Type: transactionType,
		Source: "legacy_register", RecordStatus: status,
		TotalAmount: req.AmountRemaining, AmountPaid: 0, AmountRemaining: req.AmountRemaining,
		ReceiptNo:         fmt.Sprintf("LEG-%d-%d", memberID, time.Now().UnixNano()),
		PhysicalReference: cleanOptional(req.PhysicalReference), Date: recordDate,
		Remarks: cleanOptional(req.Remarks), EnteredBy: &actorID,
		VerifiedBy: verifiedBy, VerifiedAt: verifiedAt,
	}
	if err := s.db.Create(&transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to save historical balance: %w", err)
	}
	return s.loadHistoricalTransaction(transaction.ID)
}

func (s *MemberService) VerifyHistoricalTransaction(id, actorID uint) (*models.Transaction, *models.Transaction, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	var row models.Transaction
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
		tx.Rollback()
		return nil, nil, errors.New("historical transaction not found")
	}
	if !isHistoricalType(row.Type) {
		tx.Rollback()
		return nil, nil, errors.New("this is not a historical balance")
	}
	if row.RecordStatus != "draft" {
		tx.Rollback()
		return nil, nil, errors.New("only draft historical balances can be verified")
	}
	before := row
	now := time.Now()
	if err := tx.Model(&row).Updates(map[string]interface{}{"record_status": "verified", "verified_by": actorID, "verified_at": &now}).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}
	after, err := s.loadHistoricalTransaction(id)
	return &before, after, err
}

func (s *MemberService) ReverseHistoricalTransaction(id, actorID uint, reason string) (*models.Transaction, *models.Transaction, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, nil, tx.Error
	}
	var row models.Transaction
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, id).Error; err != nil {
		tx.Rollback()
		return nil, nil, errors.New("historical transaction not found")
	}
	if !isHistoricalType(row.Type) {
		tx.Rollback()
		return nil, nil, errors.New("this is not a historical balance")
	}
	if row.RecordStatus == "reversed" {
		tx.Rollback()
		return nil, nil, errors.New("historical balance is already reversed")
	}
	if row.AmountPaid > 0.005 {
		tx.Rollback()
		return nil, nil, errors.New("a paid or partially paid balance cannot be reversed; create an accounting adjustment instead")
	}
	before := row
	now := time.Now()
	if err := tx.Model(&row).Updates(map[string]interface{}{
		"record_status": "reversed", "amount_remaining": 0,
		"reversed_by": actorID, "reversed_at": &now, "reversal_reason": reason,
	}).Error; err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, nil, err
	}
	after, err := s.loadHistoricalTransaction(id)
	return &before, after, err
}

func (s *MemberService) loadHistoricalTransaction(id uint) (*models.Transaction, error) {
	var row models.Transaction
	err := s.db.Preload("Member").Preload("FiscalYear").Preload("EnteredByUser").Preload("VerifiedByUser").First(&row, id).Error
	if err == nil {
		var docs []models.FileUpload
		_ = s.db.Where("entity = ? AND entity_id = ?", "transaction", row.ID).Order("created_at DESC").Find(&docs).Error
		row.Documents = docs
	}
	return &row, err
}

func historicalTransactionType(category, saleType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "fee":
		return historicalFeeType, nil
	case "sales":
		switch strings.ToLower(strings.TrimSpace(saleType)) {
		case "timber":
			return historicalTimberSaleType, nil
		case "firewood":
			return historicalFirewoodSaleType, nil
		case "other":
			return historicalOtherSaleType, nil
		default:
			return "", errors.New("sale_type must be timber, firewood, or other")
		}
	default:
		return "", errors.New("category must be fee or sales")
	}
}

func isHistoricalType(value string) bool {
	for _, item := range historicalTypes {
		if item == value {
			return true
		}
	}
	return false
}

func cleanOptional(value *string) *string {
	if value == nil {
		return nil
	}
	clean := strings.TrimSpace(*value)
	if clean == "" {
		return nil
	}
	return &clean
}
