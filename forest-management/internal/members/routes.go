package members

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterMemberRoutes(router *gin.RouterGroup, handler *MemberHandler) {
	memberRoutes := router.Group("/members")
	{
		// Protected routes require authentication
		memberRoutes.Use(middleware.AuthMiddleware())

		// Member's own profile (any authenticated member)
		memberRoutes.GET("/profile", handler.GetProfile)

		// Admin/Staff only routes
		memberRoutes.POST("/", middleware.RequireRole("admin", "staff"), handler.Create)
		memberRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		memberRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		memberRoutes.PUT("/:id", middleware.RequireRole("admin", "staff"), handler.Update)
		memberRoutes.DELETE("/:id", middleware.RequireRole("admin", "staff"), handler.Delete)

		// Family members
		memberRoutes.POST("/:id/family", middleware.RequireRole("admin", "staff"), handler.AddFamilyMember)
		memberRoutes.GET("/:id/family", middleware.RequireRole("admin", "staff"), handler.ListFamilyMembers)

		// Credential management
		memberRoutes.POST("/:id/reset-credentials", middleware.RequireRole("admin"), handler.ResetCredentials)

		// Bulk import
		memberRoutes.POST("/bulk-import", middleware.RequireRole("admin", "staff"), handler.BulkImport)
		memberRoutes.GET("/import-template", middleware.RequireRole("admin", "staff"), handler.DownloadImportTemplate)

		// Photo uploads
		memberRoutes.POST("/:id/upload-photo", middleware.RequireRole("admin", "staff"), handler.UploadPhoto)
		memberRoutes.POST("/:id/upload-assistant-photo", middleware.RequireRole("admin", "staff"), handler.UploadAssistantPhoto)
	}
}
