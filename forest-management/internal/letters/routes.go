package letters

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterLetterRoutes(router *gin.RouterGroup, handler *LetterHandler) {
	letterRoutes := router.Group("/letters")
	letterRoutes.Use(middleware.AuthMiddleware())
	{
		letterRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		letterRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		letterRoutes.POST("/", middleware.RequireRole("admin", "staff"), handler.Create)
		letterRoutes.PUT("/:id", middleware.RequireRole("admin", "staff"), handler.Update)
		letterRoutes.DELETE("/:id", middleware.RequireRole("admin"), handler.Delete)
		letterRoutes.POST("/:id/upload-document", middleware.RequireRole("admin", "staff"), handler.UploadDocument)
	}
}
