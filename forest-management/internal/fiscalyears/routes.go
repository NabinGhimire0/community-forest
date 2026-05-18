package fiscalyears

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterFiscalYearRoutes(router *gin.RouterGroup, handler *FiscalYearHandler) {
	fyRoutes := router.Group("/fiscal-years")
	{
		// Fiscal Year CRUD
		fyRoutes.GET("/", middleware.AuthMiddleware(), handler.List)
		fyRoutes.GET("/:id", middleware.AuthMiddleware(), handler.GetByID)
		fyRoutes.POST("/", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.Create)
		fyRoutes.PUT("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.Update)
		fyRoutes.DELETE("/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.Delete)
		fyRoutes.POST("/:id/set-active", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.SetActive)

		// Fee Settings - Note the route order matters! More specific routes first
		fyRoutes.GET("/fee/by-fiscal-year", middleware.AuthMiddleware(), handler.GetFee) // Query param route
		fyRoutes.POST("/fee", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.SetFee)
		fyRoutes.PUT("/fee/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.UpdateFee)
		fyRoutes.DELETE("/fee/:id", middleware.AuthMiddleware(), middleware.RequireRole("admin"), handler.DeleteFee)

		// Get fees by fiscal year ID - This should come AFTER the /fee routes
		fyRoutes.GET("/:id/fees", middleware.AuthMiddleware(), handler.GetFeesByFiscalYear)
	}
}
