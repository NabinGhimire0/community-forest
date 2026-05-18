package membershipfee

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterMembershipFeeRoutes(router *gin.RouterGroup, handler *MembershipFeeHandler) {
	feeRoutes := router.Group("/membership-fees")
	feeRoutes.Use(middleware.AuthMiddleware())
	{
		// Member views own fee status
		feeRoutes.GET("/my-status", handler.MyFeeStatus)

		// Admin/Staff operations
		feeRoutes.POST("/collect", middleware.RequireRole("admin", "staff"), handler.Collect)
		feeRoutes.POST("/bulk-collect", middleware.RequireRole("admin"), handler.BulkCollect)
		feeRoutes.GET("/collections", middleware.RequireRole("admin", "staff"), handler.ListCollections)
		feeRoutes.GET("/member/:id/status", middleware.RequireRole("admin", "staff"), handler.MemberFeeStatus)
	}
}
