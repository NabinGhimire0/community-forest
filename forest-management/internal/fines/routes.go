package fines

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFineRoutes(router *gin.RouterGroup, handler *FineHandler) {
	fineRoutes := router.Group("/fines")
	fineRoutes.Use(middleware.AuthMiddleware())
	{
		// Read endpoints
		fineRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		fineRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		fineRoutes.GET("/statistics", middleware.RequireRole("admin", "staff"), handler.GetStatistics)

		// Write endpoints
		fineRoutes.POST("/", middleware.RequireRole("admin", "staff"), handler.Create)
		fineRoutes.PUT("/:id", middleware.RequireRole("admin", "staff"), handler.Update)
		fineRoutes.PATCH("/:id/status", middleware.RequireRole("admin"), handler.UpdateStatus)
		fineRoutes.DELETE("/:id", middleware.RequireRole("admin"), handler.Delete)
		fineRoutes.POST("/:id/upload-photo", middleware.RequireRole("admin", "staff"), handler.UploadPhoto)
	}
}
