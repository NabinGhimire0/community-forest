package receipts

import (
	"os"
	"path/filepath"
	"strconv"

	"forest-management/pkg/middleware"
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReceiptHandler struct{ service *ReceiptService }

func NewReceiptHandler(service *ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{service: service}
}

func (h *ReceiptHandler) GenerateTransactionReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid transaction ID")
		return
	}
	filePath, err := h.service.GenerateTransactionReceipt(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	defer os.Remove(filePath)
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.FileAttachment(filePath, filepath.Base(filePath))
}

func (h *ReceiptHandler) GenerateExpenseReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid expense ID")
		return
	}
	filePath, err := h.service.GenerateExpenseReceipt(uint(id))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	defer os.Remove(filePath)
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.FileAttachment(filePath, filepath.Base(filePath))
}

func (h *ReceiptHandler) DownloadPaymentReceipt(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.BadRequest(c, "Invalid payment ID")
		return
	}
	filePath, err := h.service.GeneratePaymentReceipt(uint(id), middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	defer os.Remove(filePath)
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.FileAttachment(filePath, filepath.Base(filePath))
}
