package membershipfees

import (
	"errors"
	"fmt"
	"time"

	"forest-management/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const TransactionType = "membership_fee"

// AssignForFiscalYear creates one annual membership/Gasti fee ledger entry for
// every active member who belongs to the fiscal year. Existing entries are not
// repeated, so this function is safe to call again.
func AssignForFiscalYear(tx *gorm.DB, fiscalYear models.FiscalYear, amount float64) (int, error) {
	if amount <= 0 {
		return 0, errors.New("membership fee must be greater than zero")
	}

	var members []models.Member
	query := tx.Where("status = ?", "active")
	query = query.Where("joined_date IS NULL OR joined_date <= ?", fiscalYear.EndDate)
	if err := query.Find(&members).Error; err != nil {
		return 0, err
	}

	assigned := 0
	for _, member := range members {
		created, err := AssignForMember(tx, member, fiscalYear, amount)
		if err != nil {
			return assigned, err
		}
		if created {
			assigned++
		}
	}
	return assigned, nil
}

// AssignForMember creates the current fiscal-year membership fee for one member.
// The deterministic ledger number and duplicate check make this idempotent.
func AssignForMember(tx *gorm.DB, member models.Member, fiscalYear models.FiscalYear, amount float64) (bool, error) {
	if member.Status != "active" || amount <= 0 {
		return false, nil
	}
	if member.JoinedDate != nil && member.JoinedDate.After(fiscalYear.EndDate) {
		return false, nil
	}

	var count int64
	if err := tx.Model(&models.Transaction{}).
		Where("member_id = ? AND fiscal_year_id = ? AND type = ? AND record_status <> ?", member.ID, fiscalYear.ID, TransactionType, "reversed").
		Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}

	date := fiscalYear.StartDate
	if date.IsZero() {
		date = time.Now()
	}
	remarks := fmt.Sprintf("Automatically assigned annual Gasti/Membership fee for fiscal year %s", fiscalYear.Name)
	ledger := models.Transaction{
		MemberID:        member.ID,
		FiscalYearID:    fiscalYear.ID,
		Type:            TransactionType,
		Source:          "system",
		RecordStatus:    "verified",
		TotalAmount:     amount,
		AmountPaid:      0,
		AmountRemaining: amount,
		ReceiptNo:       fmt.Sprintf("FEE-%d-%d", fiscalYear.ID, member.ID),
		Date:            date,
		Remarks:         &remarks,
	}

	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&ledger)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// AssignActiveYearForMember assigns the active year's configured fee to a new
// member. It does nothing when there is no active year or no fee setting yet.
func AssignActiveYearForMember(tx *gorm.DB, member models.Member) (bool, error) {
	var fiscalYear models.FiscalYear
	if err := tx.Where("is_active = ?", true).First(&fiscalYear).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	var fee models.FeeSetting
	if err := tx.Where("fiscal_year_id = ?", fiscalYear.ID).First(&fee).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return AssignForMember(tx, member, fiscalYear, fee.MembershipFee)
}

// SyncUnpaidAmount updates automatically generated, completely unpaid fee rows
// when an administrator changes the fee setting. Paid or partially paid rows are
// intentionally preserved as historical financial records.
func SyncUnpaidAmount(tx *gorm.DB, fiscalYearID uint, amount float64) error {
	return tx.Model(&models.Transaction{}).
		Where("fiscal_year_id = ? AND type = ? AND source = ? AND amount_paid = 0 AND record_status = ?", fiscalYearID, TransactionType, "system", "verified").
		Updates(map[string]interface{}{
			"total_amount":     amount,
			"amount_remaining": amount,
		}).Error
}
