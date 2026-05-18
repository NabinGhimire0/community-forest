package main

import (
	"fmt"
	"log"

	"forest-management/config"
	"forest-management/database"
	"forest-management/internal/audit"
	"forest-management/internal/auth"
	"forest-management/internal/expenses"
	"forest-management/internal/fines"
	"forest-management/internal/fiscalyears"
	"forest-management/internal/letters"
	"forest-management/internal/members"
	"forest-management/internal/membershipfee"
	"forest-management/internal/notifications"
	"forest-management/internal/payments"
	"forest-management/internal/receipts"
	"forest-management/internal/reports"
	"forest-management/internal/requests"
	"forest-management/internal/resources"
	"forest-management/internal/samiti"
	"forest-management/internal/transactions"
	"forest-management/internal/uploads"
	"forest-management/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// ==========================================
	// 1. Load Configuration
	// ==========================================
	config.InitConfig()

	if config.AppConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ==========================================
	// 2. Initialize Database
	// ==========================================
	database.InitDB()

	// ==========================================
	// 3. Create Gin Router
	// ==========================================
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORSMiddleware())

	// Serve static files for uploads
	router.Static("/uploads", "./uploads")

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Community Forestry Management API is running",
		})
	})

	// ==========================================
	// 4. Create API Router Group
	// ==========================================
	api := router.Group("/api")

	// ==========================================
	// 5. Initialize Services & Handlers
	// ==========================================

	// Auth
	authService := auth.NewAuthService(database.DB)
	authHandler := auth.NewAuthHandler(authService)
	auth.RegisterAuthRoutes(api, authHandler)

	// Members
	memberService := members.NewMemberService(database.DB)
	memberHandler := members.NewMemberHandler(memberService)
	members.RegisterMemberRoutes(api, memberHandler)

	// Requests
	requestService := requests.NewRequestService(database.DB)
	requestHandler := requests.NewRequestHandler(requestService)
	requests.RegisterRequestRoutes(api, requestHandler)

	// Payments
	paymentService := payments.NewPaymentService(database.DB)
	paymentHandler := payments.NewPaymentHandler(paymentService)
	payments.RegisterPaymentRoutes(api, paymentHandler)

	// Transactions
	transactionService := transactions.NewTransactionService(database.DB)
	transactionHandler := transactions.NewTransactionHandler(transactionService)
	transactions.RegisterTransactionRoutes(api, transactionHandler)

	// Expenses
	expenseService := expenses.NewExpenseService(database.DB)
	expenseHandler := expenses.NewExpenseHandler(expenseService)
	expenses.RegisterExpenseRoutes(api, expenseHandler)

	// Fines
	fineService := fines.NewFineService(database.DB)
	fineHandler := fines.NewFineHandler(fineService)
	fines.RegisterFineRoutes(api, fineHandler)

	// Letters
	letterService := letters.NewLetterService(database.DB)
	letterHandler := letters.NewLetterHandler(letterService)
	letters.RegisterLetterRoutes(api, letterHandler)

	// Samiti
	samitiService := samiti.NewSamitiService(database.DB)
	samitiHandler := samiti.NewSamitiHandler(samitiService)
	samiti.RegisterSamitiRoutes(api, samitiHandler)

	// Resources
	resourceService := resources.NewResourceService(database.DB)
	resourceHandler := resources.NewResourceHandler(resourceService)
	resources.RegisterResourceRoutes(api, resourceHandler)

	// Fiscal Years
	fiscalYearService := fiscalyears.NewFiscalYearService(database.DB)
	fiscalYearHandler := fiscalyears.NewFiscalYearHandler(fiscalYearService)
	fiscalyears.RegisterFiscalYearRoutes(api, fiscalYearHandler)

	// Reports
	reportService := reports.NewReportService(database.DB)
	reportHandler := reports.NewReportHandler(reportService)
	reports.RegisterReportRoutes(api, reportHandler)

	// === NEW FEATURES ===

	// Membership Fee Collection
	membershipFeeService := membershipfee.NewMembershipFeeService(database.DB)
	membershipFeeHandler := membershipfee.NewMembershipFeeHandler(membershipFeeService)
	membershipfee.RegisterMembershipFeeRoutes(api, membershipFeeHandler)

	// Receipts / Invoices
	receiptService := receipts.NewReceiptService(database.DB)
	receiptHandler := receipts.NewReceiptHandler(receiptService)
	receipts.RegisterReceiptRoutes(api, receiptHandler)

	// Notifications
	notificationService := notifications.NewNotificationService(database.DB)
	notificationHandler := notifications.NewNotificationHandler(notificationService)
	notifications.RegisterNotificationRoutes(api, notificationHandler)

	// Audit Log
	auditService := audit.NewAuditService(database.DB)
	auditHandler := audit.NewAuditHandler(auditService)
	audit.RegisterAuditRoutes(api, auditHandler)

	// File Uploads
	uploadService := uploads.NewUploadService(database.DB)
	uploadHandler := uploads.NewUploadHandler(uploadService)
	uploads.RegisterUploadRoutes(api, uploadHandler)

	// ==========================================
	// 6. Start Server
	// ==========================================
	port := config.AppConfig.AppPort
	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Printf("📚 API Documentation: http://localhost:%s/api/health\n", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
