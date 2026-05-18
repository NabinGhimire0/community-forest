package receipts

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"forest-management/internal/models"

	"github.com/jung-kurt/gofpdf"
	"gorm.io/gorm"
)

type ReceiptService struct {
	db        *gorm.DB
	uploadDir string
}

func NewReceiptService(db *gorm.DB) *ReceiptService {
	uploadDir := "./uploads/receipts"
	os.MkdirAll(uploadDir, os.ModePerm)
	return &ReceiptService{db: db, uploadDir: uploadDir}
}

// GenerateTransactionReceipt creates a PDF receipt for a transaction
func (s *ReceiptService) GenerateTransactionReceipt(transactionID uint) (string, error) {
	var txn models.Transaction
	if err := s.db.Preload("Member").Preload("FiscalYear").Preload("ResourceItem").
		First(&txn, transactionID).Error; err != nil {
		return "", fmt.Errorf("transaction not found")
	}

	// Get samiti settings for header
	var settings models.SamitiSetting
	s.db.First(&settings)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetAutoPageBreak(true, 15)

	// ===== HEADER =====
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(190, 10, settings.Name, "", 1, "C", false, 0, "")

	if settings.RegistrationNo != nil {
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(190, 6, fmt.Sprintf("Reg. No: %s", *settings.RegistrationNo), "", 1, "C", false, 0, "")
	}

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(190, 5, fmt.Sprintf("%s, Ward-%d, %s, %s, %s", settings.Address, settings.WardNo, settings.Municipality, settings.District, settings.Province), "", 1, "C", false, 0, "")

	// Divider
	pdf.Ln(3)
	pdf.SetDrawColor(0, 0, 0)
	pdf.SetLineWidth(0.5)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	// ===== RECEIPT TITLE =====
	pdf.SetFont("Helvetica", "B", 14)

	var title string
	if txn.Type == "membership_fee" {
		title = "MEMBERSHIP FEE RECEIPT"
	} else {
		title = "RESOURCE SALE RECEIPT"
	}
	pdf.CellFormat(190, 10, title, "", 1, "C", false, 0, "")

	pdf.Ln(3)

	// ===== RECEIPT INFO =====
	pdf.SetFont("Helvetica", "", 10)

	receiptInfo := [][]string{
		{"Receipt No:", txn.ReceiptNo},
		{"Date:", txn.Date.Format("2006-01-02")},
		{"Fiscal Year:", txn.FiscalYear.Name},
	}

	for _, row := range receiptInfo {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(40, 7, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(150, 7, row[1], "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)

	// Divider
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(3)

	// ===== MEMBER INFO =====
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(190, 8, "Member Information", "", 1, "L", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	memberInfo := [][]string{
		{"Name:", txn.Member.Name},
		{"Membership No:", txn.Member.MembershipNo},
		{"Father's Name:", txn.Member.FatherName},
		{"Address:", fmt.Sprintf("Ward-%d, %s", txn.Member.WardNo, txn.Member.Tole)},
	}

	if txn.Member.Phone != nil {
		memberInfo = append(memberInfo, []string{"Phone:", *txn.Member.Phone})
	}

	for _, row := range memberInfo {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(40, 7, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(150, 7, row[1], "", 1, "L", false, 0, "")
	}

	pdf.Ln(3)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(3)

	// ===== PAYMENT DETAILS =====
	pdf.SetFont("Helvetica", "B", 11)
	pdf.CellFormat(190, 8, "Payment Details", "", 1, "L", false, 0, "")

	// Table header
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(80, 8, "Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Rate", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Amount", "1", 1, "C", true, 0, "")

	pdf.SetFont("Helvetica", "", 10)

	if txn.Type == "membership_fee" {
		pdf.CellFormat(80, 8, "Membership Fee", "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 8, "1", "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("Rs. %.2f", txn.TotalAmount), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("Rs. %.2f", txn.TotalAmount), "1", 1, "R", false, 0, "")
	} else {
		description := "Resource Sale"
		if txn.ResourceItem != nil {
			description = txn.ResourceItem.Name
		}
		quantity := ""
		if txn.Quantity != nil {
			quantity = fmt.Sprintf("%.2f", *txn.Quantity)
		}
		rate := ""
		if txn.RatePerUnit != nil {
			rate = fmt.Sprintf("Rs. %.2f", *txn.RatePerUnit)
		}
		pdf.CellFormat(80, 8, description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 8, quantity, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, rate, "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("Rs. %.2f", txn.TotalAmount), "1", 1, "R", false, 0, "")
	}

	// Totals
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(150, 8, "Total Amount", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("Rs. %.2f", txn.TotalAmount), "1", 1, "R", false, 0, "")

	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(150, 8, "Amount Paid", "1", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("Rs. %.2f", txn.AmountPaid), "1", 1, "R", false, 0, "")

	if txn.AmountRemaining > 0 {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(150, 8, "Amount Remaining", "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("Rs. %.2f", txn.AmountRemaining), "1", 1, "R", false, 0, "")
	}

	pdf.Ln(8)

	// ===== SIGNATURES =====
	pdf.SetFont("Helvetica", "", 10)

	// Signature row
	pdf.CellFormat(60, 8, "_____________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(70, 8, "", "", 0, "C", false, 0, "")
	pdf.CellFormat(60, 8, "_____________________", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(60, 5, "Member Signature", "", 0, "C", false, 0, "")
	pdf.CellFormat(70, 5, "", "", 0, "C", false, 0, "")
	pdf.CellFormat(60, 5, "Authorized Signature", "", 1, "C", false, 0, "")

	// Footer note
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.CellFormat(190, 5, "This is a computer-generated receipt. No physical signature is required.", "", 1, "C", false, 0, "")
	pdf.CellFormat(190, 5, fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")), "", 1, "C", false, 0, "")

	// ===== SAVE FILE =====
	filename := fmt.Sprintf("receipt_%s_%d.pdf", txn.ReceiptNo, time.Now().Unix())
	filePath := filepath.Join(s.uploadDir, filename)

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return filePath, nil
}

// GenerateExpenseReceipt creates a PDF receipt for an expense
func (s *ReceiptService) GenerateExpenseReceipt(expenseID uint) (string, error) {
	var expense models.Expense
	if err := s.db.Preload("Category").Preload("FiscalYear").Preload("Creator").
		First(&expense, expenseID).Error; err != nil {
		return "", fmt.Errorf("expense not found")
	}

	var settings models.SamitiSetting
	s.db.First(&settings)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Helvetica", "B", 16)
	pdf.CellFormat(190, 10, settings.Name, "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(190, 5, fmt.Sprintf("%s, Ward-%d, %s, %s", settings.Address, settings.WardNo, settings.Municipality, settings.District), "", 1, "C", false, 0, "")

	pdf.Ln(3)
	pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
	pdf.Ln(5)

	// Title
	pdf.SetFont("Helvetica", "B", 14)
	pdf.CellFormat(190, 10, "EXPENSE VOUCHER", "", 1, "C", false, 0, "")
	pdf.Ln(3)

	// Info
	pdf.SetFont("Helvetica", "", 10)
	voucherInfo := [][]string{
		{"Date:", expense.ExpenseDate.Format("2006-01-02")},
		{"Category:", expense.Category.Name},
		{"Title:", expense.Title},
		{"Paid To:", expense.PaidTo},
		{"Payment Method:", expense.PaymentMethod},
		{"Amount:", fmt.Sprintf("Rs. %.2f", expense.Amount)},
	}
	if expense.ReceiptNo != nil {
		voucherInfo = append(voucherInfo, []string{"Receipt No:", *expense.ReceiptNo})
	}
	if expense.Remarks != nil {
		voucherInfo = append(voucherInfo, []string{"Remarks:", *expense.Remarks})
	}

	for _, row := range voucherInfo {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(40, 7, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(150, 7, row[1], "", 1, "L", false, 0, "")
	}

	// Signatures
	pdf.Ln(15)
	pdf.CellFormat(60, 8, "_____________________", "", 0, "C", false, 0, "")
	pdf.CellFormat(70, 8, "", "", 0, "C", false, 0, "")
	pdf.CellFormat(60, 8, "_____________________", "", 1, "C", false, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(60, 5, "Prepared By", "", 0, "C", false, 0, "")
	pdf.CellFormat(70, 5, "", "", 0, "C", false, 0, "")
	pdf.CellFormat(60, 5, "Approved By", "", 1, "C", false, 0, "")

	// Save
	filename := fmt.Sprintf("expense_%d_%d.pdf", expense.ID, time.Now().Unix())
	filePath := filepath.Join(s.uploadDir, filename)

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return filePath, nil
}
