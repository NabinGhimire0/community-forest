package requests

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRequestRoutes(router *gin.RouterGroup, handler *RequestHandler) {
	reqRoutes := router.Group("/requests")
	reqRoutes.Use(middleware.AuthMiddleware())
	{
		// Member endpoints
		reqRoutes.POST("/", handler.Create)
		reqRoutes.GET("/my", handler.MyRequests)
		reqRoutes.GET("/:id", handler.GetByID)
		reqRoutes.GET("/statistics", handler.GetStatistics)

		// Admin/Staff endpoints
		reqRoutes.GET("/", middleware.RequireRole("admin", "staff"), handler.List)
		reqRoutes.POST("/:id/approve", middleware.RequireRole("admin", "staff"), handler.Approve)
		reqRoutes.POST("/:id/reject", middleware.RequireRole("admin", "staff"), handler.Reject)
	}
}
