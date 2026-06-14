package adminops

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler *Handler) {
	routes := router.Group("/admin/system")
	routes.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		routes.GET("/exports/datasets", handler.Datasets)
		routes.POST("/exports/csv", handler.ExportCSV)
		routes.POST("/exports/all", handler.ExportAll)
		routes.POST("/backups/database", handler.DatabaseBackup)
		routes.POST("/backups/full", handler.FullBackup)
	}
}
