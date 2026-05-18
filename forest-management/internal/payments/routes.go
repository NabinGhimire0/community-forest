package payments

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(router *gin.RouterGroup, handler *PaymentHandler) {
	payRoutes := router.Group("/payments")
	payRoutes.Use(middleware.AuthMiddleware())
	{
		// Member endpoints
		payRoutes.POST("/", handler.Create)
		payRoutes.GET("/my", handler.MyPayments)

		// Admin/Staff endpoints
		payRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		payRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		payRoutes.POST("/:id/verify", middleware.RequireRole("admin", "staff"), handler.Verify)
		payRoutes.GET("/statistics", middleware.RequireRole("admin", "staff"), handler.GetPaymentStats)
	}
}
