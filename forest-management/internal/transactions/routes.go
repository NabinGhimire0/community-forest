package transactions

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterTransactionRoutes(router *gin.RouterGroup, handler *TransactionHandler) {
	txnRoutes := router.Group("/transactions")
	txnRoutes.Use(middleware.AuthMiddleware())
	{
		// Member endpoints
		txnRoutes.GET("/my", handler.MyTransactions)

		// Admin/Staff endpoints
		txnRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		txnRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		txnRoutes.GET("/summary", middleware.RequireRole("admin", "staff"), handler.Summary)
		txnRoutes.GET("/dashboard-summary", middleware.RequireRole("admin", "staff"), handler.GetDashboardSummary)
	}
}
