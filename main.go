package main

import (
	"context"
	"fmt"
	"goboot/config"
	"goboot/internal/model"
	"goboot/pkg/database"
	"goboot/pkg/logger"
	"goboot/router"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load config
	if err := config.InitConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	logCfg := &logger.Config{
		Level:      config.AppConfig.Log.Level,
		Filename:   config.AppConfig.Log.Filename,
		MaxSize:    config.AppConfig.Log.MaxSize,
		MaxBackups: config.AppConfig.Log.MaxBackups,
		MaxAge:     config.AppConfig.Log.MaxAge,
		Compress:   config.AppConfig.Log.Compress,
		Console:    config.AppConfig.Log.Console,
	}
	if err := logger.InitLogger(logCfg); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	logger.Info("Config loaded successfully")

	// Initialize MySQL
	if err := database.InitMySQL(); err != nil {
		logger.Error("Failed to connect to MySQL", slog.Any("error", err))
		return
	}
	logger.Info("MySQL connected successfully")

	// Initialize Redis
	if err := database.InitRedis(); err != nil {
		logger.Error("Failed to connect to Redis", slog.Any("error", err))
		return
	}
	logger.Info("Redis connected successfully")

	// Auto migrate database tables
	if err := model.AutoMigrate(); err != nil {
		logger.Error("Failed to migrate database", slog.Any("error", err))
		return
	}
	logger.Info("Database migrated successfully")

	// Setup router
	r := router.SetupRouter()

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server", slog.Any("error", err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 10 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("Server exited")
}
