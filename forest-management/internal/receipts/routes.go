package receipts

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterReceiptRoutes(router *gin.RouterGroup, handler *ReceiptHandler) {
	receiptRoutes := router.Group("/receipts")
	receiptRoutes.Use(middleware.AuthMiddleware())
	{
		receiptRoutes.GET("/payment/:id", handler.DownloadPaymentReceipt)
		receiptRoutes.POST("/transaction/:id", middleware.RequireRole("admin", "staff"), handler.GenerateTransactionReceipt)
		receiptRoutes.POST("/expense/:id", middleware.RequireRole("admin", "staff"), handler.GenerateExpenseReceipt)
	}
}
