package main

import (
	"fmt"
	"goboot/config"
	"goboot/internal/model"
	"goboot/pkg/database"
	"goboot/pkg/logger"
	"goboot/router"
	"log"
	"log/slog"
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

	// Start server
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	logger.Info("Server starting", slog.String("addr", addr))
	if err := r.Run(addr); err != nil {
		logger.Error("Failed to start server", slog.Any("error", err))
	}
}
