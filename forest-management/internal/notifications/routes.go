package notifications

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(router *gin.RouterGroup, handler *NotificationHandler) {
	notifRoutes := router.Group("/notifications")
	notifRoutes.Use(middleware.AuthMiddleware())
	{
		// User's own notifications
		notifRoutes.GET("/", handler.List)
		notifRoutes.GET("/unread-count", handler.UnreadCount)
		notifRoutes.POST("/:id/read", handler.MarkRead)
		notifRoutes.POST("/mark-all-read", handler.MarkAllRead)

		// Admin creates notifications
		notifRoutes.POST("/", middleware.RequireRole("admin"), handler.AdminCreate)
	}
}
