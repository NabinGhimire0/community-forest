package auth

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, handler *AuthHandler) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/login", middleware.LoginRateLimit(), handler.Login)

		protected := authRoutes.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/profile", handler.GetProfile)
			protected.POST("/logout", handler.Logout)
			protected.PUT("/change-password", handler.ChangePassword)
			protected.GET("/sessions", handler.ListSessions)
			protected.POST("/sessions/revoke-all", handler.RevokeAllSessions)
			protected.POST("/mfa/setup", handler.BeginMFA)
			protected.POST("/mfa/enable", handler.EnableMFA)
			protected.POST("/mfa/disable", handler.DisableMFA)
			protected.POST("/admin/reset-password", middleware.RequireRole("admin"), handler.AdminResetPassword)
		}
	}
}
