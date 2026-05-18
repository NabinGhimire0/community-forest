package expenses

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterExpenseRoutes(router *gin.RouterGroup, handler *ExpenseHandler) {
	expRoutes := router.Group("/expenses")
	expRoutes.Use(middleware.AuthMiddleware())
	{
		// Expense endpoints
		expRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		expRoutes.GET("/:id", middleware.RequireRole("admin", "staff"), handler.GetByID)
		expRoutes.POST("/", middleware.RequireRole("admin", "staff"), handler.Create)
		expRoutes.PUT("/:id", middleware.RequireRole("admin", "staff"), handler.Update)
		expRoutes.DELETE("/:id", middleware.RequireRole("admin"), handler.Delete)
		expRoutes.POST("/:id/upload-photo", middleware.RequireRole("admin", "staff"), handler.UploadBillPhoto)

		// Category endpoints
		expRoutes.GET("/categories", handler.ListCategories)
		expRoutes.GET("/categories/:id", middleware.RequireRole("admin", "staff"), handler.GetCategoryByID)
		expRoutes.POST("/categories", middleware.RequireRole("admin"), handler.CreateCategory)
		expRoutes.PUT("/categories/:id", middleware.RequireRole("admin"), handler.UpdateCategory)
		expRoutes.DELETE("/categories/:id", middleware.RequireRole("admin"), handler.DeleteCategory)
	}
}
