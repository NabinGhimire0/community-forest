package resources

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterResourceRoutes(router *gin.RouterGroup, handler *ResourceHandler) {
	resRoutes := router.Group("/resources")
	{
		// ==================== Resource Types ====================
		resRoutes.GET("/types", handler.ListTypes)
		resRoutes.GET("/types/:id", middleware.AuthMiddleware(), handler.GetTypeByID)
		resRoutes.POST("/types", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.CreateType)
		resRoutes.PUT("/types/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateType)
		resRoutes.DELETE("/types/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.DeleteType)

		// ==================== Resource Items ====================
		resRoutes.GET("/items", handler.ListItems)
		resRoutes.GET("/items/:id", middleware.AuthMiddleware(), handler.GetItemByID)
		resRoutes.POST("/items", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.CreateItem)
		resRoutes.PUT("/items/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateItem)
		resRoutes.DELETE("/items/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.DeleteItem)

		// ==================== Resource Rates ====================
		resRoutes.GET("/rates", middleware.AuthMiddleware(), handler.ListRates)
		resRoutes.GET("/rates/:id", middleware.AuthMiddleware(), handler.GetRateByID)
		resRoutes.POST("/rates", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.SetRate)
		resRoutes.PUT("/rates/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateRate)
		resRoutes.DELETE("/rates/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.DeleteRate)

		// ==================== Stock ====================
		resRoutes.GET("/stock", middleware.AuthMiddleware(), handler.ListStock)
		resRoutes.GET("/stock/:id", middleware.AuthMiddleware(), handler.GetStockByID)
		resRoutes.POST("/stock", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateStock)
		resRoutes.PUT("/stock/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateStockQuantity)
		resRoutes.DELETE("/stock/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.DeleteStock)
	}
}
