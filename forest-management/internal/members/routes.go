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
		memberRoutes.POST("/", middleware.RequireRole("admin"), handler.Create)
		memberRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		memberRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		memberRoutes.PUT("/:id", middleware.RequireRole("admin", "staff"), handler.Update)
		memberRoutes.DELETE("/:id", middleware.RequireRole("admin"), handler.Delete)

		// Family members
		memberRoutes.POST("/:id/family", middleware.RequireRole("admin", "staff"), handler.AddFamilyMember)
		memberRoutes.GET("/:id/family", middleware.RequireRole("admin", "staff"), handler.ListFamilyMembers)
		memberRoutes.PUT("/:id/family/:familyId", middleware.RequireRole("admin", "staff"), handler.UpdateFamilyMember)
		memberRoutes.DELETE("/:id/family/:familyId", middleware.RequireRole("admin", "staff"), handler.DeleteFamilyMember)

		// Credential management
		memberRoutes.POST("/:id/reset-credentials", middleware.RequireRole("admin"), handler.ResetCredentials)

		// Bulk import
		memberRoutes.POST("/bulk-import", middleware.RequireRole("admin"), handler.BulkImport)
		memberRoutes.GET("/import-template", middleware.RequireRole("admin", "staff"), handler.DownloadImportTemplate)

		// Photo uploads
		memberRoutes.POST("/:id/upload-photo", middleware.RequireRole("admin", "staff"), handler.UploadPhoto)
		memberRoutes.POST("/:id/upload-assistant-photo", middleware.RequireRole("admin", "staff"), handler.UploadAssistantPhoto)

		memberRoutes.GET("/:id/financial-summary", middleware.RequireRole("admin", "staff"), handler.GetMemberFinancialSummary)
		memberRoutes.GET("/:id/fee-details", middleware.RequireRole("admin", "staff"), handler.GetMemberFeeDetails)
		memberRoutes.GET("/:id/sales-details", middleware.RequireRole("admin", "staff"), handler.GetMemberSalesDetails)
		memberRoutes.POST("/:id/historical-transaction", middleware.RequireRole("admin", "staff"), handler.CreateHistoricalTransaction)
		memberRoutes.POST("/historical-transactions/:transactionId/verify", middleware.RequireRole("admin"), handler.VerifyHistoricalTransaction)
		memberRoutes.POST("/historical-transactions/:transactionId/reverse", middleware.RequireRole("admin"), handler.ReverseHistoricalTransaction)
	}
}
