package samiti

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterSamitiRoutes(router *gin.RouterGroup, handler *SamitiHandler) {
	samitiRoutes := router.Group("/samiti")
	{
		// Settings - public read, admin write
		samitiRoutes.GET("/settings", handler.GetSettings)
		samitiRoutes.PUT("/settings", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateSettings)
		samitiRoutes.POST("/settings/upload-logo", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UploadLogo)

		// Committee Heads
		samitiRoutes.GET("/heads", handler.ListHeads)
		samitiRoutes.GET("/heads/:id", handler.GetHeadByID)
		samitiRoutes.POST("/heads", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.CreateHead)
		samitiRoutes.PUT("/heads/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateHead)
		samitiRoutes.DELETE("/heads/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.DeleteHead)
		samitiRoutes.POST("/heads/:id/upload-photo", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UploadHeadPhoto)
	}
}
