package audit

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuditRoutes(router *gin.RouterGroup, handler *AuditHandler) {
	auditRoutes := router.Group("/audit")
	auditRoutes.Use(middleware.AuthMiddleware())
	auditRoutes.Use(middleware.RequireRole("admin"))
	{
		auditRoutes.GET("/", handler.List)
		auditRoutes.GET("/entity-history", handler.EntityHistory)
	}
}
