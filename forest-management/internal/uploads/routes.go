package uploads

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUploadRoutes(router *gin.RouterGroup, handler *UploadHandler) {
	uploadRoutes := router.Group("/uploads")
	uploadRoutes.Use(middleware.AuthMiddleware())
	{
		uploadRoutes.POST("/", handler.Upload)
		uploadRoutes.POST("/multiple", handler.UploadMultiple)
		uploadRoutes.GET("/", handler.List)
		uploadRoutes.DELETE("/:id", middleware.RequireRole("admin"), handler.Delete)
	}

	// Public file serving (files are accessed via URL)
	router.GET("/uploads/files/:folder/:filename", handler.ServeFile)
}
