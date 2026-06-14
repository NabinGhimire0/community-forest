package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"forest-management/config"
	"forest-management/internal/audit"
	"forest-management/internal/models"
	"forest-management/internal/notifications"
	"forest-management/pkg/response"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const moneyTolerance = 0.005

var legacyTypes = []string{
	"legacy_gasti_fee",
	"legacy_timber_sale",
	"legacy_firewood_sale",
	"legacy_other_sale",
}

var feeLedgerTypes = []string{
	"membership_fee",
	"legacy_gasti_fee",
}

var payableLedgerTypes = []string{
	"membership_fee",
	"legacy_gasti_fee",
	"legacy_timber_sale",
	"legacy_firewood_sale",
	"legacy_other_sale",
}

type PaymentService struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{
		db:         db,
		httpClient: &http.Client{Timeout: 12 * time.Second},
	}
}

type EsewaInitiation struct {
	PaymentID uint              `json:"payment_id"`
	ActionURL string            `json:"action_url"`
	Fields    map[string]string `json:"fields"`
}

type esewaCallbackPayload struct {
	TransactionCode  string          `json:"transaction_code"`
	Status           string          `json:"status"`
	TotalAmount      json.RawMessage `json:"total_amount"`
	TransactionUUID  string          `json:"transaction_uuid"`
	ProductCode      string          `json:"product_code"`
	SignedFieldNames string          `json:"signed_field_names"`
	Signature        string          `json:"signature"`
}

type esewaStatusResponse struct {
	ProductCode     string          `json:"product_code"`
	TransactionUUID string          `json:"transaction_uuid"`
	TotalAmount     json.RawMessage `json:"total_amount"`
	Status          string          `json:"status"`
	RefID           *string         `json:"ref_id"`
	Code            *int            `json:"code"`
	ErrorMessage    *string         `json:"error_message"`
}

// InitiateEsewa creates a pending eSewa payment for either the member's own
// approved resource request or one payable annual Gasti/Membership fee ledger.
func (s *PaymentService) InitiateEsewa(userID uint, requestID, ledgerTransactionID *uint) (*EsewaInitiation, error) {
	if (requestID == nil) == (ledgerTransactionID == nil) {
		return nil, errors.New("select either an approved request or a membership fee")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer rollbackOnPanic(tx)

	var member models.Member
	if err := tx.Where("user_id = ? AND status = ?", userID, "active").First(&member).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("active member profile not found")
	}

	var amount float64
	var transactionUUID string
	payment := models.Payment{
		MemberID:      member.ID,
		PaymentMethod: "esewa",
		Status:        "pending",
		CreatedBy:     &userID,
	}

	expiredBefore := time.Now().Add(-5 * time.Minute)
	if requestID != nil {
		var req models.Request
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("ResourceItem.Type").First(&req, *requestID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("request not found")
		}
		if req.MemberID != member.ID {
			tx.Rollback()
			return nil, errors.New("this request does not belong to you")
		}
		if req.Status != "approved" {
			tx.Rollback()
			return nil, errors.New("only approved requests can be paid")
		}
		if req.TotalAmount == nil || *req.TotalAmount <= 0 {
			tx.Rollback()
			return nil, errors.New("request total is not available")
		}

		if err := tx.Model(&models.Payment{}).
			Where("request_id = ? AND payment_method = ? AND status = ? AND created_at < ?", req.ID, "esewa", "pending", expiredBefore).
			Updates(map[string]interface{}{"status": "failed", "gateway_status": "EXPIRED"}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		var activePending int64
		if err := tx.Model(&models.Payment{}).Where("request_id = ? AND payment_method = ? AND status = ?", req.ID, "esewa", "pending").Count(&activePending).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if activePending > 0 {
			tx.Rollback()
			return nil, errors.New("an eSewa payment is already in progress for this request")
		}

		paid, err := paidAmountForRequest(tx, req.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		amount = roundMoney(*req.TotalAmount - paid)
		if amount <= moneyTolerance {
			tx.Rollback()
			return nil, errors.New("this request is already fully paid")
		}
		payment.RequestID = &req.ID
		transactionUUID = fmt.Sprintf("REQ-%d-%s", req.ID, uuid.NewString())
	} else {
		var ledger models.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("FiscalYear").First(&ledger, *ledgerTransactionID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("membership fee record not found")
		}
		if ledger.MemberID != member.ID {
			tx.Rollback()
			return nil, errors.New("this membership fee does not belong to you")
		}
		if !contains(feeLedgerTypes, ledger.Type) || ledger.RecordStatus != "verified" {
			tx.Rollback()
			return nil, errors.New("this ledger entry cannot be paid online")
		}
		if ledger.AmountRemaining <= moneyTolerance {
			tx.Rollback()
			return nil, errors.New("this fiscal-year membership fee is already paid")
		}

		if err := tx.Model(&models.Payment{}).
			Where("ledger_transaction_id = ? AND payment_method = ? AND status = ? AND created_at < ?", ledger.ID, "esewa", "pending", expiredBefore).
			Updates(map[string]interface{}{"status": "failed", "gateway_status": "EXPIRED"}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		var activePending int64
		if err := tx.Model(&models.Payment{}).Where("ledger_transaction_id = ? AND payment_method = ? AND status = ?", ledger.ID, "esewa", "pending").Count(&activePending).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if activePending > 0 {
			tx.Rollback()
			return nil, errors.New("an eSewa payment is already in progress for this fiscal-year fee")
		}

		amount = roundMoney(ledger.AmountRemaining)
		payment.LedgerTransactionID = &ledger.ID
		transactionUUID = fmt.Sprintf("FEE-%d-%s", ledger.ID, uuid.NewString())
	}

	payment.Amount = amount
	payment.GatewayTransactionUUID = &transactionUUID
	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to initialize payment: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	formattedAmount := formatAmount(amount)
	signedFields := "total_amount,transaction_uuid,product_code"
	message := fmt.Sprintf("total_amount=%s,transaction_uuid=%s,product_code=%s", formattedAmount, transactionUUID, config.AppConfig.EsewaProductCode)
	signature := createHMAC(message, config.AppConfig.EsewaSecretKey)
	backend := strings.TrimRight(config.AppConfig.PublicBackendURL, "/")
	fields := map[string]string{
		"amount":                  formattedAmount,
		"tax_amount":              "0",
		"total_amount":            formattedAmount,
		"transaction_uuid":        transactionUUID,
		"product_code":            config.AppConfig.EsewaProductCode,
		"product_service_charge":  "0",
		"product_delivery_charge": "0",
		"success_url":             backend + "/api/payments/esewa/callback",
		"failure_url":             backend + "/api/payments/esewa/failure?transaction_uuid=" + url.QueryEscape(transactionUUID),
		"signed_field_names":      signedFields,
		"signature":               signature,
	}

	return &EsewaInitiation{PaymentID: payment.ID, ActionURL: config.AppConfig.EsewaFormURL, Fields: fields}, nil
}

// ProcessEsewaCallback validates eSewa's signed callback, confirms the status
// through eSewa's status API and then settles the request atomically.
func (s *PaymentService) ProcessEsewaCallback(encodedData string) (*models.Payment, error) {
	encodedData = strings.ReplaceAll(strings.TrimSpace(encodedData), " ", "+")
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, errors.New("invalid eSewa callback data")
	}

	var payload esewaCallbackPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, errors.New("invalid eSewa callback payload")
	}
	if payload.TransactionUUID == "" || payload.Signature == "" || payload.SignedFieldNames == "" {
		return nil, errors.New("incomplete eSewa callback payload")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return nil, errors.New("invalid eSewa callback fields")
	}
	message, err := signedMessage(raw, payload.SignedFieldNames)
	if err != nil {
		return nil, err
	}
	expected := createHMAC(message, config.AppConfig.EsewaSecretKey)
	if !hmac.Equal([]byte(expected), []byte(payload.Signature)) {
		return nil, errors.New("eSewa callback signature verification failed")
	}
	if payload.ProductCode != config.AppConfig.EsewaProductCode {
		return nil, errors.New("unexpected eSewa product code")
	}

	var payment models.Payment
	if err := s.db.Where("gateway_transaction_uuid = ?", payload.TransactionUUID).First(&payment).Error; err != nil {
		return nil, errors.New("payment record not found")
	}
	callbackAmount, err := rawNumber(payload.TotalAmount)
	if err != nil || math.Abs(callbackAmount-payment.Amount) > moneyTolerance {
		return nil, errors.New("eSewa callback amount does not match the payment")
	}
	if payload.Status != "COMPLETE" {
		_ = s.db.Model(&payment).Updates(map[string]interface{}{
			"gateway_status":   payload.Status,
			"gateway_response": string(decoded),
		}).Error
		return nil, fmt.Errorf("eSewa returned status %s", payload.Status)
	}

	statusResponse, rawStatus, err := s.checkEsewaStatus(payload.TransactionUUID, payment.Amount)
	if err != nil {
		return nil, err
	}
	if statusResponse.Status != "COMPLETE" {
		_ = s.db.Model(&payment).Updates(map[string]interface{}{
			"gateway_status":   statusResponse.Status,
			"gateway_response": string(rawStatus),
		}).Error
		return nil, fmt.Errorf("eSewa status verification returned %s", statusResponse.Status)
	}
	statusAmount, err := rawNumber(statusResponse.TotalAmount)
	if err != nil || math.Abs(statusAmount-payment.Amount) > moneyTolerance {
		return nil, errors.New("eSewa verified amount does not match the payment")
	}

	return s.completePayment(payment.ID, payload.TransactionCode, statusResponse.RefID, "COMPLETE", string(rawStatus), nil)
}

// CheckEsewaPaymentStatus reconciles a pending eSewa payment when the browser
// callback was lost or interrupted. It uses the same authoritative status API
// and ownership rules as the normal callback flow.
func (s *PaymentService) CheckEsewaPaymentStatus(paymentID, userID uint, role string) (*models.Payment, error) {
	if !s.UserCanAccessPayment(paymentID, userID, role) {
		return nil, errors.New("you cannot access this payment")
	}
	var payment models.Payment
	if err := s.db.First(&payment, paymentID).Error; err != nil {
		return nil, errors.New("payment not found")
	}
	if payment.PaymentMethod != "esewa" {
		return nil, errors.New("this is not an eSewa payment")
	}
	if payment.Status == "paid" {
		return s.GetPaymentByID(payment.ID)
	}
	if payment.Status != "pending" || payment.GatewayTransactionUUID == nil {
		return nil, errors.New("this eSewa payment cannot be checked")
	}

	statusResponse, rawStatus, err := s.checkEsewaStatus(*payment.GatewayTransactionUUID, payment.Amount)
	if err != nil {
		return nil, err
	}
	verifiedAmount, err := rawNumber(statusResponse.TotalAmount)
	if err != nil || math.Abs(verifiedAmount-payment.Amount) > moneyTolerance {
		return nil, errors.New("eSewa verified amount does not match the payment")
	}

	switch statusResponse.Status {
	case "COMPLETE":
		return s.completePayment(payment.ID, "", statusResponse.RefID, "COMPLETE", string(rawStatus), nil)
	case "CANCELED", "NOT_FOUND", "FULL_REFUND":
		if err := s.db.Model(&payment).Updates(map[string]interface{}{
			"status": "failed", "gateway_status": statusResponse.Status, "gateway_response": string(rawStatus),
		}).Error; err != nil {
			return nil, err
		}
		return s.GetPaymentByID(payment.ID)
	default:
		if err := s.db.Model(&payment).Updates(map[string]interface{}{
			"gateway_status": statusResponse.Status, "gateway_response": string(rawStatus),
		}).Error; err != nil {
			return nil, err
		}
		return s.GetPaymentByID(payment.ID)
	}
}

func (s *PaymentService) ProcessEsewaFailure(transactionUUID string) (*models.Payment, error) {
	if strings.TrimSpace(transactionUUID) == "" {
		return nil, errors.New("transaction UUID is required")
	}
	var payment models.Payment
	if err := s.db.Where("gateway_transaction_uuid = ?", transactionUUID).First(&payment).Error; err != nil {
		return nil, errors.New("payment record not found")
	}
	if payment.Status == "paid" {
		return s.GetPaymentByID(payment.ID)
	}
	if payment.Status != "pending" {
		return s.GetPaymentByID(payment.ID)
	}

	// A browser reaching the failure URL is not sufficient proof that money was
	// not transferred. Reconcile with eSewa before changing the local status.
	statusResponse, rawStatus, err := s.checkEsewaStatus(transactionUUID, payment.Amount)
	if err != nil {
		return nil, fmt.Errorf("payment result could not be confirmed; use Check Status from the payment history: %w", err)
	}
	verifiedAmount, err := rawNumber(statusResponse.TotalAmount)
	if err != nil || math.Abs(verifiedAmount-payment.Amount) > moneyTolerance {
		return nil, errors.New("eSewa verified amount does not match the payment")
	}

	switch statusResponse.Status {
	case "COMPLETE":
		return s.completePayment(payment.ID, "", statusResponse.RefID, "COMPLETE", string(rawStatus), nil)
	case "CANCELED", "NOT_FOUND", "FULL_REFUND":
		if err := s.db.Model(&payment).Updates(map[string]interface{}{
			"status": "failed", "gateway_status": statusResponse.Status, "gateway_response": string(rawStatus),
		}).Error; err != nil {
			return nil, err
		}
	default:
		if err := s.db.Model(&payment).Updates(map[string]interface{}{
			"gateway_status": statusResponse.Status, "gateway_response": string(rawStatus),
		}).Error; err != nil {
			return nil, err
		}
	}
	return s.GetPaymentByID(payment.ID)
}

// CreateCashPayment records an immediately verified cash payment. Only the
// admin route exposes this method. Exactly one target must be supplied.
func (s *PaymentService) CreateCashPayment(adminUserID uint, input CreateCashPaymentInput, clientIP, userAgent string) (*models.Payment, error) {
	if (input.RequestID == nil) == (input.LedgerTransactionID == nil) {
		return nil, errors.New("select either a request or a payable ledger entry")
	}
	if input.Amount <= 0 {
		return nil, errors.New("payment amount must be greater than zero")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer rollbackOnPanic(tx)

	var memberID uint
	if input.RequestID != nil {
		var req models.Request
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&req, *input.RequestID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("request not found")
		}
		if req.Status != "approved" {
			tx.Rollback()
			return nil, errors.New("only approved requests can receive payment")
		}
		memberID = req.MemberID
		if input.MemberID != nil && *input.MemberID != memberID {
			tx.Rollback()
			return nil, errors.New("request does not belong to the selected member")
		}
		if req.TotalAmount == nil {
			tx.Rollback()
			return nil, errors.New("request total is unavailable")
		}
		paid, err := paidAmountForRequest(tx, req.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		outstanding := roundMoney(*req.TotalAmount - paid)
		if input.Amount-outstanding > moneyTolerance {
			tx.Rollback()
			return nil, fmt.Errorf("payment exceeds the outstanding amount of Rs. %.2f", outstanding)
		}
	} else {
		var ledger models.Transaction
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ledger, *input.LedgerTransactionID).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("ledger entry not found")
		}
		if !contains(payableLedgerTypes, ledger.Type) {
			tx.Rollback()
			return nil, errors.New("this ledger entry cannot receive a cash payment")
		}
		if ledger.RecordStatus != "verified" {
			tx.Rollback()
			return nil, errors.New("historical balance must be verified before payment")
		}
		if ledger.AmountRemaining <= moneyTolerance {
			tx.Rollback()
			return nil, errors.New("historical balance is already paid")
		}
		if input.Amount-ledger.AmountRemaining > moneyTolerance {
			tx.Rollback()
			return nil, fmt.Errorf("payment exceeds the remaining balance of Rs. %.2f", ledger.AmountRemaining)
		}
		memberID = ledger.MemberID
		if input.MemberID != nil && *input.MemberID != memberID {
			tx.Rollback()
			return nil, errors.New("ledger entry does not belong to the selected member")
		}
	}

	now := time.Now()
	receipt := newPaymentReceiptNo()
	payment := models.Payment{
		MemberID:            memberID,
		RequestID:           input.RequestID,
		LedgerTransactionID: input.LedgerTransactionID,
		Amount:              roundMoney(input.Amount),
		PaymentMethod:       "cash",
		ReceiptNo:           &receipt,
		Remarks:             input.Remarks,
		Status:              "paid",
		CreatedBy:           &adminUserID,
		PaidAt:              &now,
		VerifiedAt:          &now,
	}
	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to record cash payment: %w", err)
	}
	if err := s.applySettlement(tx, &payment, now); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	audit.CreateAuditEntry(s.db, &adminUserID, "create", "cash_payment", &payment.ID, nil, payment, clientIP, userAgent, "Cash payment recorded and settled")
	s.notifyPayment(&payment, "Payment Received", fmt.Sprintf("Cash payment of Rs. %.2f has been recorded.", payment.Amount), "success")
	return s.GetPaymentByID(payment.ID)
}

func (s *PaymentService) completePayment(paymentID uint, transactionCode string, refID *string, gatewayStatus, gatewayResponse string, verifiedBy *uint) (*models.Payment, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer rollbackOnPanic(tx)

	var payment models.Payment
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&payment, paymentID).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("payment not found")
	}
	if payment.Status == "paid" {
		tx.Rollback()
		return s.GetPaymentByID(payment.ID)
	}
	if payment.Status != "pending" {
		tx.Rollback()
		return nil, errors.New("payment is no longer pending")
	}

	now := time.Now()
	receipt := newPaymentReceiptNo()
	updates := map[string]interface{}{
		"status":           "paid",
		"paid_at":          &now,
		"verified_at":      &now,
		"receipt_no":       receipt,
		"gateway_status":   gatewayStatus,
		"gateway_response": gatewayResponse,
	}
	if transactionCode != "" {
		updates["transaction_id"] = transactionCode
	}
	if refID != nil {
		updates["gateway_reference_id"] = *refID
	}
	if verifiedBy != nil {
		updates["created_by"] = *verifiedBy
	}
	if err := tx.Model(&payment).Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.First(&payment, paymentID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := s.applySettlement(tx, &payment, now); err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	audit.CreateAuditEntry(s.db, verifiedBy, "settle_online_payment", "payment", &payment.ID, nil, payment, "", "", "eSewa payment verified by signed callback and status API")
	s.notifyPayment(&payment, "Payment Successful", fmt.Sprintf("Your payment of Rs. %.2f was completed successfully.", payment.Amount), "success")
	return s.GetPaymentByID(payment.ID)
}

func (s *PaymentService) applySettlement(tx *gorm.DB, payment *models.Payment, now time.Time) error {
	if payment.RequestID != nil {
		return s.settleRequest(tx, payment, now)
	}
	if payment.LedgerTransactionID != nil {
		return s.settleLedger(tx, payment)
	}
	return errors.New("payment has no settlement target")
}

func (s *PaymentService) settleLedger(tx *gorm.DB, payment *models.Payment) error {
	var ledger models.Transaction
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ledger, *payment.LedgerTransactionID).Error; err != nil {
		return errors.New("linked ledger entry not found")
	}
	if ledger.RecordStatus != "verified" || !contains(payableLedgerTypes, ledger.Type) {
		return errors.New("linked ledger balance is not payable")
	}
	if payment.Amount-ledger.AmountRemaining > moneyTolerance {
		return errors.New("payment exceeds ledger balance")
	}
	ledger.AmountPaid = roundMoney(ledger.AmountPaid + payment.Amount)
	ledger.AmountRemaining = roundMoney(ledger.TotalAmount - ledger.AmountPaid)
	if ledger.AmountRemaining < moneyTolerance {
		ledger.AmountRemaining = 0
	}
	return tx.Save(&ledger).Error
}

func (s *PaymentService) settleRequest(tx *gorm.DB, payment *models.Payment, now time.Time) error {
	var req models.Request
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&req, *payment.RequestID).Error; err != nil {
		return errors.New("linked request not found")
	}
	if req.TotalAmount == nil || *req.TotalAmount <= 0 {
		return errors.New("request total is unavailable")
	}
	paid, err := paidAmountForRequest(tx, req.ID)
	if err != nil {
		return err
	}
	if paid-*req.TotalAmount > moneyTolerance {
		return errors.New("payment would exceed request total")
	}
	remaining := roundMoney(*req.TotalAmount - paid)
	if remaining < moneyTolerance {
		remaining = 0
	}

	var ledger models.Transaction
	err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("request_id = ?", req.ID).First(&ledger).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ledger = models.Transaction{
			MemberID:        payment.MemberID,
			FiscalYearID:    req.FiscalYearID,
			ResourceItemID:  &req.ResourceItemID,
			RequestID:       &req.ID,
			Type:            "resource_sale",
			Source:          "system",
			RecordStatus:    "verified",
			Quantity:        req.QuantityApproved,
			RatePerUnit:     req.RatePerUnit,
			TotalAmount:     *req.TotalAmount,
			AmountPaid:      paid,
			AmountRemaining: remaining,
			ReceiptNo:       fmt.Sprintf("LED-%d-%s", req.ID, uuid.NewString()),
			Date:            now,
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return fmt.Errorf("failed to create request ledger: %w", err)
		}
	} else if err != nil {
		return err
	} else {
		ledger.AmountPaid = paid
		ledger.AmountRemaining = remaining
		if err := tx.Save(&ledger).Error; err != nil {
			return err
		}
	}

	if remaining == 0 && req.Status != "completed" {
		if req.QuantityApproved == nil || *req.QuantityApproved <= 0 {
			return errors.New("approved quantity is missing")
		}
		result := tx.Model(&models.Stock{}).
			Where("resource_item_id = ? AND fiscal_year_id = ? AND remaining_quantity >= ? AND reserved_quantity >= ?", req.ResourceItemID, req.FiscalYearID, *req.QuantityApproved, *req.QuantityApproved).
			Updates(map[string]interface{}{
				"remaining_quantity": gorm.Expr("remaining_quantity - ?", *req.QuantityApproved),
				"reserved_quantity":  gorm.Expr("reserved_quantity - ?", *req.QuantityApproved),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("insufficient stock for the approved quantity")
		}
		if err := tx.Model(&models.Request{}).Where("id = ?", req.ID).Update("status", "completed").Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PaymentService) checkEsewaStatus(transactionUUID string, amount float64) (*esewaStatusResponse, []byte, error) {
	endpoint, err := url.Parse(config.AppConfig.EsewaStatusURL)
	if err != nil {
		return nil, nil, errors.New("invalid eSewa status URL")
	}
	q := endpoint.Query()
	q.Set("product_code", config.AppConfig.EsewaProductCode)
	q.Set("total_amount", formatAmount(amount))
	q.Set("transaction_uuid", transactionUUID)
	endpoint.RawQuery = q.Encode()

	resp, err := s.httpClient.Get(endpoint.String())
	if err != nil {
		return nil, nil, fmt.Errorf("could not verify payment with eSewa: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, nil, errors.New("could not read eSewa verification response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, body, fmt.Errorf("eSewa verification returned HTTP %d", resp.StatusCode)
	}
	var result esewaStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, body, errors.New("invalid eSewa verification response")
	}
	if result.ErrorMessage != nil && *result.ErrorMessage != "" {
		return nil, body, errors.New(*result.ErrorMessage)
	}
	if result.ProductCode != config.AppConfig.EsewaProductCode || result.TransactionUUID != transactionUUID {
		return nil, body, errors.New("eSewa verification identifiers do not match")
	}
	return &result, body, nil
}

func (s *PaymentService) ListPayments(page, perPage int, fiscalYearID, status, memberID string) ([]models.Payment, *response.Pagination, error) {
	page, perPage = normalizePagination(page, perPage)
	var payments []models.Payment
	var total int64
	query := s.db.Model(&models.Payment{})
	if fiscalYearID != "" {
		query = query.
			Joins("LEFT JOIN requests ON requests.id = payments.request_id").
			Joins("LEFT JOIN transactions target_transactions ON target_transactions.id = payments.ledger_transaction_id").
			Where("COALESCE(requests.fiscal_year_id, target_transactions.fiscal_year_id) = ?", fiscalYearID)
	}
	if status != "" {
		query = query.Where("payments.status = ?", status)
	}
	if memberID != "" {
		query = query.Where("payments.member_id = ?", memberID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, nil, err
	}
	offset := (page - 1) * perPage
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	err := query.
		Preload("Member").Preload("Member.User").
		Preload("Request.ResourceItem.Type").
		Preload("LedgerTransaction.FiscalYear").
		Order("payments.created_at DESC").Offset(offset).Limit(perPage).Find(&payments).Error
	return payments, &response.Pagination{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}, err
}

func (s *PaymentService) GetPaymentByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := s.db.Preload("Member").Preload("Member.User").Preload("Request.ResourceItem.Type").Preload("Request.FiscalYear").Preload("LedgerTransaction.FiscalYear").First(&payment, id).Error
	return &payment, err
}

func (s *PaymentService) GetMemberPayments(userID uint, page, perPage int) ([]models.Payment, *response.Pagination, error) {
	var member models.Member
	if err := s.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		return nil, nil, errors.New("member not found")
	}
	return s.ListPayments(page, perPage, "", "", strconv.FormatUint(uint64(member.ID), 10))
}

func (s *PaymentService) GetPaymentStats(fiscalYearID string) (map[string]interface{}, error) {
	base := func() *gorm.DB {
		q := s.db.Model(&models.Payment{})
		if fiscalYearID != "" {
			q = q.Joins("LEFT JOIN requests ON requests.id = payments.request_id").
				Joins("LEFT JOIN transactions target_transactions ON target_transactions.id = payments.ledger_transaction_id").
				Where("COALESCE(requests.fiscal_year_id, target_transactions.fiscal_year_id) = ?", fiscalYearID)
		}
		return q
	}
	var totalAmount, cashAmount, onlineAmount, pendingAmount float64
	var totalCount int64
	if err := base().Where("payments.status = ?", "paid").Select("COALESCE(SUM(payments.amount),0)").Scan(&totalAmount).Error; err != nil {
		return nil, err
	}
	if err := base().Where("payments.status = ?", "paid").Count(&totalCount).Error; err != nil {
		return nil, err
	}
	if err := base().Where("payments.status = ? AND payments.payment_method = ?", "paid", "cash").Select("COALESCE(SUM(payments.amount),0)").Scan(&cashAmount).Error; err != nil {
		return nil, err
	}
	if err := base().Where("payments.status = ? AND payments.payment_method = ?", "paid", "esewa").Select("COALESCE(SUM(payments.amount),0)").Scan(&onlineAmount).Error; err != nil {
		return nil, err
	}
	if err := base().Where("payments.status = ?", "pending").Select("COALESCE(SUM(payments.amount),0)").Scan(&pendingAmount).Error; err != nil {
		return nil, err
	}
	return map[string]interface{}{"total_amount": totalAmount, "total_count": totalCount, "cash_amount": cashAmount, "online_amount": onlineAmount, "pending_amount": pendingAmount}, nil
}

func (s *PaymentService) UserCanAccessPayment(paymentID, userID uint, role string) bool {
	if role == "admin" || role == "staff" {
		return true
	}
	var count int64
	s.db.Model(&models.Payment{}).Joins("JOIN members ON members.id = payments.member_id").Where("payments.id = ? AND members.user_id = ?", paymentID, userID).Count(&count)
	return count > 0
}

func (s *PaymentService) notifyPayment(payment *models.Payment, title, message, kind string) {
	var member models.Member
	if err := s.db.Preload("User").First(&member, payment.MemberID).Error; err != nil || member.UserID == nil {
		return
	}
	notifications.NewNotificationService(s.db).NotifyUser(*member.UserID, title, message, kind, stringPtr("payment"), &payment.ID)
}

func paidAmountForRequest(tx *gorm.DB, requestID uint) (float64, error) {
	var paid float64
	err := tx.Model(&models.Payment{}).Where("request_id = ? AND status = ?", requestID, "paid").Select("COALESCE(SUM(amount),0)").Scan(&paid).Error
	return roundMoney(paid), err
}

func signedMessage(raw map[string]json.RawMessage, names string) (string, error) {
	fields := strings.Split(names, ",")
	parts := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		value, ok := raw[field]
		if !ok {
			return "", fmt.Errorf("signed field %s is missing", field)
		}
		parts = append(parts, field+"="+rawString(value))
	}
	return strings.Join(parts, ","), nil
}

func rawString(raw json.RawMessage) string {
	var text string
	if json.Unmarshal(raw, &text) == nil {
		return text
	}
	return strings.TrimSpace(string(raw))
}

func rawNumber(raw json.RawMessage) (float64, error) {
	return strconv.ParseFloat(strings.ReplaceAll(rawString(raw), ",", ""), 64)
}

func createHMAC(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func formatAmount(amount float64) string { return strconv.FormatFloat(roundMoney(amount), 'f', 2, 64) }
func roundMoney(amount float64) float64  { return math.Round(amount*100) / 100 }
func newPaymentReceiptNo() string {
	return fmt.Sprintf("PAY-%d-%s", time.Now().Year(), strings.ToUpper(uuid.NewString()[:8]))
}
func stringPtr(value string) *string { return &value }
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func rollbackOnPanic(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}
func normalizePagination(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	return page, perPage
}
