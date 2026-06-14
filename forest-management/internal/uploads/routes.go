package uploads

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUploadRoutes(router *gin.RouterGroup, handler *UploadHandler) {
	uploadRoutes := router.Group("/uploads")
	uploadRoutes.Use(middleware.AuthMiddleware())
	{
		uploadRoutes.POST("/", middleware.RequireRole("admin", "staff"), handler.Upload)
		uploadRoutes.POST("/multiple", middleware.RequireRole("admin", "staff"), handler.UploadMultiple)
		uploadRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		uploadRoutes.DELETE("/:id", middleware.RequireRole("admin"), handler.Delete)
		uploadRoutes.GET("/files/:folder/:filename", handler.ServeFile)
	}
}
