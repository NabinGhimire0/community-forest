package receipts

import (
	"strconv"

	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReceiptHandler struct {
	service *ReceiptService
}

func NewReceiptHandler(service *ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{service: service}
}

// GenerateTransactionReceipt — Generate PDF receipt for a transaction
func (h *ReceiptHandler) GenerateTransactionReceipt(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	filePath, err := h.service.GenerateTransactionReceipt(uint(id))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Receipt generated", gin.H{
		"file_path": filePath,
		"download":  "/api/receipts/download?path=" + filePath,
	})
}

// GenerateExpenseReceipt — Generate PDF receipt for an expense
func (h *ReceiptHandler) GenerateExpenseReceipt(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	filePath, err := h.service.GenerateExpenseReceipt(uint(id))
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}

	response.Success(c, "Expense receipt generated", gin.H{
		"file_path": filePath,
		"download":  "/api/receipts/download?path=" + filePath,
	})
}

// DownloadReceipt — Serve a PDF file for download
func (h *ReceiptHandler) DownloadReceipt(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		response.BadRequest(c, "File path is required")
		return
	}

	// Security: prevent directory traversal
	// Only allow files from the receipts directory
	if len(filePath) < 11 || filePath[:11] != "./uploads/receipts" {
		response.Forbidden(c, "Access denied")
		return
	}

	c.FileAttachment(filePath, "receipt.pdf")
}
