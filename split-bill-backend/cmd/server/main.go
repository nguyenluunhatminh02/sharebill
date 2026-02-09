// @title          Split Bill API
// @version        4.0.0
// @description    API server for Split Bill application - manage groups, bills, transactions, OCR receipt scanning, and payments
// @termsOfService http://swagger.io/terms/

// @contact.name  Split Bill Support
// @contact.email support@splitbill.app

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Firebase Bearer token. Format: "Bearer {token}"

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/splitbill/backend/internal/config"
	"github.com/splitbill/backend/internal/database"
	"github.com/splitbill/backend/internal/handlers"
	"github.com/splitbill/backend/internal/middleware"
	"github.com/splitbill/backend/internal/repository"
	"github.com/splitbill/backend/internal/services"
	"github.com/splitbill/backend/pkg/visionapi"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connections
	mongoDB := database.NewMongoDB(&cfg.MongoDB)
	defer mongoDB.Disconnect()

	redisClient := database.NewRedisClient(&cfg.Redis)
	defer redisClient.Close()

	// Create MongoDB indexes for performance
	database.EnsureIndexes(mongoDB)

	// Initialize Vision API client
	visionClient, err := visionapi.NewClient(
		logger,
		cfg.Google.VisionCredentials,
		cfg.Google.VisionAPIKey,
	)
	if err != nil {
		logger.Warn("Vision API client initialization failed", zap.Error(err))
	}
	if visionClient != nil {
		defer visionClient.Close()
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(mongoDB)
	groupRepo := repository.NewGroupRepository(mongoDB)
	billRepo := repository.NewBillRepository(mongoDB)
	transactionRepo := repository.NewTransactionRepository(mongoDB)
	ocrRepo := repository.NewOCRRepository(mongoDB)
	activityRepo := repository.NewActivityRepository(mongoDB)

	// Initialize services
	authService := services.NewAuthService(userRepo)
	groupService := services.NewGroupService(groupRepo, userRepo)
	billService := services.NewBillService(billRepo, groupRepo, userRepo)
	debtService := services.NewDebtService(billRepo, transactionRepo, userRepo)
	ocrService := services.NewOCRService(ocrRepo, billRepo, groupRepo, visionClient, logger)
	notifService := services.NewNotificationService(userRepo, logger)
	activityService := services.NewActivityService(activityRepo, userRepo, groupRepo, logger)
	statsService := services.NewStatsService(billRepo, transactionRepo, groupRepo, userRepo)

	// Log notification service status
	_ = notifService // Will be used when FCM is configured

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	groupHandler := handlers.NewGroupHandler(groupService)
	billHandler := handlers.NewBillHandler(billService, debtService)
	transactionHandler := handlers.NewTransactionHandler(transactionRepo, userRepo)
	ocrHandler := handlers.NewOCRHandler(ocrService)
	paymentHandler := handlers.NewPaymentHandler(userRepo)
	activityHandler := handlers.NewActivityHandler(activityService, userRepo)
	statsHandler := handlers.NewStatsHandler(statsService, userRepo)

	// Image upload handler
	uploadDir := filepath.Join(".", "uploads")
	baseURL := fmt.Sprintf("http://localhost%s", cfg.Server.Port)
	imageHandler := handlers.NewImageHandler(uploadDir, baseURL)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.Firebase.CredentialsFile)

	// Setup Gin router
	gin.SetMode(cfg.Server.Mode)
	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryWithLogger(logger))
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimitMiddleware(100)) // 100 requests/second per IP

	// Serve uploaded images statically
	router.Static("/uploads", uploadDir)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "split-bill-api",
			"version": "4.0.0",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Auth routes (require Firebase token)
	auth := v1.Group("/auth")
	auth.Use(authMiddleware.Authenticate())
	{
		auth.POST("/verify-token", authHandler.VerifyToken)
		auth.GET("/me", authHandler.GetMe)
		auth.PUT("/profile", authHandler.UpdateProfile)
	}

	// Group routes
	groups := v1.Group("/groups")
	groups.Use(authMiddleware.Authenticate())
	{
		groups.POST("", groupHandler.CreateGroup)
		groups.GET("", groupHandler.ListGroups)
		groups.POST("/join", groupHandler.JoinGroup)
		groups.GET("/:id", groupHandler.GetGroup)
		groups.PUT("/:id", groupHandler.UpdateGroup)
		groups.DELETE("/:id", groupHandler.DeleteGroup)
		groups.POST("/:id/members", groupHandler.AddMember)
		groups.DELETE("/:id/members/:userId", groupHandler.RemoveMember)

		// Bills within a group
		groups.POST("/:id/bills", billHandler.CreateBill)
		groups.GET("/:id/bills", billHandler.ListBills)

		// Balances and settlements
		groups.GET("/:id/balances", billHandler.GetGroupBalances)
		groups.GET("/:id/settlements", billHandler.GetSettlements)

		// Group activities (Phase 4)
		groups.GET("/:id/activities", activityHandler.GetGroupActivities)

		// Group stats & export (Phase 5)
		groups.GET("/:id/stats", statsHandler.GetGroupStats)
		groups.GET("/:id/stats/categories", statsHandler.GetGroupCategoryStats)
		groups.GET("/:id/export", statsHandler.ExportGroupSummary)
	}

	// Bill routes (direct access)
	bills := v1.Group("/bills")
	bills.Use(authMiddleware.Authenticate())
	{
		bills.GET("/:id", billHandler.GetBill)
		bills.PUT("/:id", billHandler.UpdateBill)
		bills.DELETE("/:id", billHandler.DeleteBill)
	}

	// Transaction routes
	transactions := v1.Group("/transactions")
	transactions.Use(authMiddleware.Authenticate())
	{
		transactions.POST("", transactionHandler.CreateTransaction)
		transactions.PUT("/:id/confirm", transactionHandler.ConfirmTransaction)
	}

	// User routes
	users := v1.Group("/users")
	users.Use(authMiddleware.Authenticate())
	{
		users.GET("/me/debts", transactionHandler.GetUserDebts)
	}

	// OCR routes (Phase 2) - with strict rate limit for expensive operations
	ocr := v1.Group("/ocr")
	ocr.Use(authMiddleware.Authenticate())
	ocr.Use(middleware.RateLimitByUser(30)) // 30 OCR scans per minute per user
	{
		ocr.POST("/scan", ocrHandler.ScanReceipt)
		ocr.POST("/scan-base64", ocrHandler.ScanReceiptBase64)
		ocr.GET("/:id/result", ocrHandler.GetOCRResult)
		ocr.POST("/:id/confirm", ocrHandler.ConfirmOCR)
		ocr.GET("/pending", ocrHandler.GetPendingScans)
	}

	// Image upload routes (Phase 2)
	upload := v1.Group("/upload")
	upload.Use(authMiddleware.Authenticate())
	upload.Use(middleware.RateLimitByUser(60)) // 60 uploads per minute per user
	{
		upload.POST("/image", imageHandler.UploadImage)
		upload.POST("/image-base64", imageHandler.UploadBase64Image)
	}

	// Payment routes (Phase 4)
	payment := v1.Group("/payment")
	payment.Use(authMiddleware.Authenticate())
	{
		payment.POST("/deeplink", paymentHandler.GenerateDeeplink)
		payment.POST("/vietqr", paymentHandler.GenerateVietQR)
		payment.GET("/user/:userId", paymentHandler.GetUserPaymentInfo)
		payment.GET("/banks", paymentHandler.GetSupportedBanks)
	}

	// Activity routes (Phase 4)
	activities := v1.Group("/activities")
	activities.Use(authMiddleware.Authenticate())
	{
		activities.GET("/me", activityHandler.GetUserActivities)
	}

	// Stats routes (Phase 5)
	stats := v1.Group("/stats")
	stats.Use(authMiddleware.Authenticate())
	{
		stats.GET("/me", statsHandler.GetUserStats)
	}

	// Categories route (Phase 5)
	v1.GET("/categories", statsHandler.GetCategoryList)

	// Create HTTP server with proper timeouts
	srv := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Split Bill API server starting on %s", cfg.Server.Port)
		log.Printf("📚 Swagger UI: http://localhost%s/swagger/index.html", cfg.Server.Port)
		log.Printf("📷 OCR endpoint: POST /api/v1/ocr/scan")
		log.Printf("🖼️  Upload endpoint: POST /api/v1/upload/image")
		log.Printf("💰 Payment endpoint: POST /api/v1/payment/deeplink")
		log.Printf("📋 Activity endpoint: GET /api/v1/activities/me")
		log.Printf("📊 Stats endpoint: GET /api/v1/stats/me")
		log.Printf("📁 Categories endpoint: GET /api/v1/categories")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Shutting down server gracefully...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("⚠️  Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}
