package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"visekai/backend/internal/config"
	"visekai/backend/internal/database"
	"visekai/backend/internal/handlers"
	"visekai/backend/internal/middleware"
	"visekai/backend/internal/ocr"
	"visekai/backend/internal/repository"
	"visekai/backend/internal/services"
	"visekai/backend/pkg/logger"
	"visekai/backend/pkg/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logger.Init(cfg.LogLevel)

	// Initialize database
	db, err := database.New(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.Pool)
	documentRepo := repository.NewDocumentRepository(db.Pool)
	jobRepo := repository.NewJobRepository(db.Pool)
	resultRepo := repository.NewResultRepository(db.Pool)

	// Initialize storage
	fileStorage, err := storage.NewStorage(cfg.StoragePath)
	if err != nil {
		logger.Fatal("Failed to initialize storage", "error", err)
	}

	// Initialize OCR client
	ocrClient := ocr.NewClient(cfg.OCRServiceURL)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	jobService := services.NewJobService(jobRepo, resultRepo, documentRepo, ocrClient)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, userRepo)
	documentHandler := handlers.NewDocumentHandler(documentRepo, fileStorage, cfg.MaxFileSize, []string{".jpg", ".jpeg", ".png", ".pdf", ".tiff", ".tif", ".gif", ".bmp", ".webp"})
	jobHandler := handlers.NewJobHandler(jobService)
	healthCheckHandler := handlers.NewHealthCheckHandler(db.Pool)

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Create router
	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health check endpoint with database verification
	router.GET("/api/v1/health", healthCheckHandler.Handle)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes with rate limiting
		authRateLimiter := middleware.NewRateLimiter(10, 1*time.Minute) // 10 requests per minute
		auth := v1.Group("/auth")
		auth.Use(authRateLimiter.RateLimit())
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.GET("/me", middleware.AuthRequired(authService), authHandler.GetCurrentUser)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.AuthRequired(authService))
		{
			// Document routes
			documents := protected.Group("/documents")
			{
				documents.POST("/upload", documentHandler.Upload)
				documents.GET("", documentHandler.List)
				documents.GET("/:id", documentHandler.Get)
				documents.DELETE("/:id", documentHandler.Delete)
			}

			// OCR routes
			ocr := protected.Group("/ocr")
			{
				ocr.POST("/submit", jobHandler.SubmitJob)
				ocr.POST("/batch", jobHandler.SubmitBatchJob)
				ocr.GET("/jobs", jobHandler.ListJobs)
				ocr.GET("/jobs/:id", jobHandler.GetJob)
				ocr.GET("/jobs/:id/result", jobHandler.GetJobResult)
				ocr.PUT("/jobs/:id/cancel", jobHandler.CancelJob)
				ocr.DELETE("/jobs/:id", jobHandler.DeleteJob)
			}

			// Results routes
			results := protected.Group("/results")
			{
				results.GET("/:id", handlers.GetResult)
				results.GET("/:id/download", handlers.DownloadResult)
				results.GET("/:id/preview", handlers.PreviewResult)
			}

			// Settings routes
			settings := protected.Group("/settings")
			{
				settings.GET("", handlers.GetSettings)
				settings.PUT("", handlers.UpdateSettings)
			}
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
