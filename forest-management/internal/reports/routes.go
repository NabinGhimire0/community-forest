package reports

import (
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterReportRoutes(router *gin.RouterGroup, handler *ReportHandler) {
	reportRoutes := router.Group("/reports")
	reportRoutes.Use(middleware.AuthMiddleware())
	reportRoutes.Use(middleware.RequireRole("admin", "staff"))
	{
		reportRoutes.GET("/dashboard", handler.Dashboard)
		reportRoutes.GET("/dashboard/charts", handler.DashboardCharts)
		reportRoutes.GET("/members", handler.MemberReport)
		reportRoutes.GET("/resources", handler.ResourceReport)
		reportRoutes.GET("/financial", handler.FinancialReport)
	}
}
