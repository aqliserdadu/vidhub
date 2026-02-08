package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"videodownload/config"
	"videodownload/internal/handler"
	"videodownload/internal/service"
	"videodownload/internal/storage"
	"videodownload/pkg/logger"
	"videodownload/pkg/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	if err := logger.Init(&cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Logger.Info("Starting Video Download Server",
		zap.String("host", cfg.Server.Host),
		zap.Int("port", cfg.Server.Port),
	)

	// Initialize storage manager
	storageManager := storage.NewManager(&cfg.Storage)
	if err := storageManager.EnsureDownloadDir(); err != nil {
		logger.Logger.Fatal("Failed to create download directory", zap.Error(err))
	}
	storageManager.Start()
	defer storageManager.Stop()

	// Initialize services
	videoService := service.NewVideoService(
		cfg.Python.Host,
		cfg.Python.Port,
		cfg.Python.Timeout,
	)
	downloadService := service.NewDownloadService(
		cfg.Python.Host,
		cfg.Python.Port,
		cfg.Python.Timeout,
		storageManager,
	)

	// Initialize quota service
	quotaService := service.NewQuotaService(&cfg.Quota)
	defer quotaService.Stop()

	// Initialize rate limit service
	rateLimitService := service.NewRateLimitService(&cfg.RateLimit)
	defer rateLimitService.Stop()

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Add middleware
	router.Use(logger.GinLogger())

	// Add rate limiting middleware
	if cfg.RateLimit.Enabled {
		router.Use(middleware.RateLimitMiddleware(rateLimitService))
		logger.Logger.Info("Rate limiting enabled", zap.Int("requests_per_minute", cfg.RateLimit.RequestsPerMinute))
	}

	// Add quota check middleware
	if cfg.Quota.Enabled {
		router.Use(middleware.QuotaCheckMiddleware(quotaService))
		logger.Logger.Info("Quota limiting enabled", zap.Int64("daily_limit_mb", cfg.Quota.DailyLimitMB), zap.Int("reset_hour", cfg.Quota.ResetHour))
	}

	// Determine frontend path
	frontendPath := "./frontend"
	if _, err := os.Stat("../frontend"); err == nil {
		frontendPath = "../frontend"
	}
	staticPath := filepath.Join(frontendPath, "static")
	indexPath := filepath.Join(frontendPath, "index.html")

	logger.Logger.Info("Frontend paths",
		zap.String("static", staticPath),
		zap.String("index", indexPath))

	// Public frontend
	router.Static("/static", staticPath)
	router.StaticFile("/", indexPath)

	// API handlers
	videoHandler := handler.NewVideoHandler(videoService, cfg)
	downloadHandler := handler.NewDownloadHandler(downloadService, cfg, quotaService, rateLimitService)

	// Routes
	api := router.Group("/api")
	{
		// Video info
		api.GET("/video/info", videoHandler.GetVideoInfo)

		// Downloads
		api.POST("/download", downloadHandler.StartDownload)
		api.GET("/download/:id", downloadHandler.GetFile)

		// Health check
		api.GET("/health", videoHandler.HealthCheck)
	}

	// Start server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.Timeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.Timeout) * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Logger.Info("Server listening", zap.String("address", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Fatal("Server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Logger.Info("Server stopped")
}
