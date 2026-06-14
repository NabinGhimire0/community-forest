package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"forest-management/config"
	"forest-management/database"
	"forest-management/internal/adminops"
	"forest-management/internal/audit"
	"forest-management/internal/auth"
	"forest-management/internal/expenses"
	"forest-management/internal/files"
	"forest-management/internal/fines"
	"forest-management/internal/fiscalyears"
	"forest-management/internal/letters"
	"forest-management/internal/members"
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
	config.InitConfig()
	if config.AppConfig.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	database.InitDB()

	router := gin.New()
	// Prevent multipart requests from buffering an unbounded amount of data in memory.
	// Individual upload policies enforce stricter per-file limits.
	router.MaxMultipartMemory = 12 << 20
	router.Use(
		middleware.RequestID(),
		gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{
			"/api/health/live",
			"/api/payments/esewa/callback",
			"/api/payments/esewa/failure",
		}}),
		middleware.SecureRecovery(),
		middleware.SecurityHeaders(),
		middleware.CORSMiddleware(config.AppConfig.CORSOrigins),
		middleware.GlobalRateLimit(),
		middleware.LimitJSONBody(config.AppConfig.MaxJSONBodyBytes),
	)

	configureTrustedProxies(router)
	files.RegisterRoutes(router, files.NewHandler(database.DB))

	router.GET("/api/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/health", func(c *gin.Context) {
		sqlDB, err := database.DB.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
			return
		}
		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api")

	authService := auth.NewAuthService(database.DB)
	auth.RegisterAuthRoutes(api, auth.NewAuthHandler(authService))

	memberService := members.NewMemberService(database.DB)
	members.RegisterMemberRoutes(api, members.NewMemberHandler(memberService))

	requestService := requests.NewRequestService(database.DB)
	requests.RegisterRequestRoutes(api, requests.NewRequestHandler(requestService))

	paymentService := payments.NewPaymentService(database.DB)
	payments.RegisterPaymentRoutes(api, payments.NewPaymentHandler(paymentService))

	transactionService := transactions.NewTransactionService(database.DB)
	transactions.RegisterTransactionRoutes(api, transactions.NewTransactionHandler(transactionService))

	expenseService := expenses.NewExpenseService(database.DB)
	expenses.RegisterExpenseRoutes(api, expenses.NewExpenseHandler(expenseService))

	fineService := fines.NewFineService(database.DB)
	fines.RegisterFineRoutes(api, fines.NewFineHandler(fineService))

	letterService := letters.NewLetterService(database.DB)
	letters.RegisterLetterRoutes(api, letters.NewLetterHandler(letterService))

	samitiService := samiti.NewSamitiService(database.DB)
	samiti.RegisterSamitiRoutes(api, samiti.NewSamitiHandler(samitiService))

	resourceService := resources.NewResourceService(database.DB)
	resources.RegisterResourceRoutes(api, resources.NewResourceHandler(resourceService))

	fiscalYearService := fiscalyears.NewFiscalYearService(database.DB)
	fiscalyears.RegisterFiscalYearRoutes(api, fiscalyears.NewFiscalYearHandler(fiscalYearService))

	reportService := reports.NewReportService(database.DB)
	reports.RegisterReportRoutes(api, reports.NewReportHandler(reportService))

	receiptService := receipts.NewReceiptService(database.DB)
	receipts.RegisterReceiptRoutes(api, receipts.NewReceiptHandler(receiptService))

	notificationService := notifications.NewNotificationService(database.DB)
	notifications.RegisterNotificationRoutes(api, notifications.NewNotificationHandler(notificationService))

	auditService := audit.NewAuditService(database.DB)
	audit.RegisterAuditRoutes(api, audit.NewAuditHandler(auditService))

	uploadService := uploads.NewUploadService(database.DB)
	uploads.RegisterUploadRoutes(api, uploads.NewUploadHandler(uploadService))

	adminService := adminops.NewService(database.DB)
	adminops.RegisterRoutes(api, adminops.NewHandler(adminService, database.DB))

	server := &http.Server{
		Addr:              ":" + config.AppConfig.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      20 * time.Minute, // encrypted backup downloads can be large
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	fmt.Printf("🚀 Server starting on port %s\n", config.AppConfig.AppPort)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case signalValue := <-stop:
		log.Printf("shutdown signal received: %s", signalValue)
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("forced server shutdown: %v", err)
	}
	if sqlDB, err := database.DB.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func configureTrustedProxies(router *gin.Engine) {
	configured := strings.TrimSpace(config.AppConfig.TrustedProxies)
	if configured == "" {
		if err := router.SetTrustedProxies(nil); err != nil {
			log.Fatalf("could not disable trusted proxies: %v", err)
		}
		return
	}
	proxies := make([]string, 0)
	for _, value := range strings.Split(configured, ",") {
		if proxy := strings.TrimSpace(value); proxy != "" {
			proxies = append(proxies, proxy)
		}
	}
	if err := router.SetTrustedProxies(proxies); err != nil {
		log.Fatalf("invalid TRUSTED_PROXIES: %v", err)
	}
}
