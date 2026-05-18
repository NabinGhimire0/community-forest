package reports

import (
	"forest-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	service *ReportService
}

func NewReportHandler(service *ReportService) *ReportHandler {
	return &ReportHandler{service: service}
}

// Dashboard - Admin dashboard with key metrics
func (h *ReportHandler) Dashboard(c *gin.Context) {
	dashboard, err := h.service.GetDashboard()
	if err != nil {
		response.InternalError(c, "Failed to load dashboard")
		return
	}
	response.Success(c, "Dashboard data", dashboard)
}

// DashboardCharts - Chart data for the dashboard
func (h *ReportHandler) DashboardCharts(c *gin.Context) {
	charts, err := h.service.GetDashboardCharts()
	if err != nil {
		response.InternalError(c, "Failed to load chart data")
		return
	}
	response.Success(c, "Dashboard chart data", charts)
}

// MemberReport - Member statistics
func (h *ReportHandler) MemberReport(c *gin.Context) {
	report, err := h.service.GetMemberReport()
	if err != nil {
		response.InternalError(c, "Failed to generate member report")
		return
	}
	response.Success(c, "Member report", report)
}

// ResourceReport - Resource stock and sales
func (h *ReportHandler) ResourceReport(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	report, err := h.service.GetResourceReport(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to generate resource report")
		return
	}
	response.Success(c, "Resource report", report)
}

// FinancialReport - Complete financial overview
func (h *ReportHandler) FinancialReport(c *gin.Context) {
	fiscalYearID := c.Query("fiscal_year_id")
	if fiscalYearID == "" {
		response.BadRequest(c, "fiscal_year_id is required")
		return
	}
	report, err := h.service.GetFinancialReport(fiscalYearID)
	if err != nil {
		response.InternalError(c, "Failed to generate financial report")
		return
	}
	response.Success(c, "Financial report", report)
}
