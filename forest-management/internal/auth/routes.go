package auth

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authRoutes := router.Group("/auth")
	{
		// Public routes
		authRoutes.POST("/login", handler.Login)

		// Protected routes
		authRoutes.Use(middleware.AuthMiddleware())
		authRoutes.GET("/profile", handler.GetProfile)
		authRoutes.PUT("/change-password", handler.ChangePassword)

		// Admin only
		authRoutes.POST("/admin/reset-password", middleware.RequireRole("admin"), handler.AdminResetPassword)
	}
}
