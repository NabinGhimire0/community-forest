package payments

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterPaymentRoutes(router *gin.RouterGroup, handler *PaymentHandler) {
	// Public eSewa browser-return endpoints. Security comes from signed callback
	// validation and server-to-server status verification, not from a user JWT.
	router.GET("/payments/esewa/callback", handler.EsewaCallback)
	router.GET("/payments/esewa/failure", handler.EsewaFailure)

	payRoutes := router.Group("/payments")
	payRoutes.Use(middleware.AuthMiddleware())
	{
		payRoutes.POST("/esewa/initiate", middleware.RequireRole("member"), handler.InitiateEsewa)
		payRoutes.POST("/esewa/:id/check-status", handler.CheckEsewaStatus)
		payRoutes.POST("/cash", middleware.RequireRole("admin"), handler.CreateCash)
		payRoutes.GET("/my", middleware.RequireRole("member"), handler.MyPayments)
		payRoutes.GET("/statistics", middleware.RequireRole("admin", "staff"), handler.GetPaymentStats)
		payRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		payRoutes.GET("/:id", handler.GetByID)
	}
}
