package main

import (
	"fmt"
	"goboot/config"
	"goboot/internal/model"
	"goboot/internal/service"
	"goboot/pkg/database"
	"goboot/pkg/logger"
	"goboot/router"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v3"
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

	// Create Fiber app
	app := fiber.New()

	// Setup router
	router.SetupRouter(app)

	// Initialize and start cron scheduler
	cronSvc := service.GetCronService()
	registerCronJobs(cronSvc)
	cronSvc.Start()

	// Start server in goroutine
	addr := fmt.Sprintf("%s:%d", config.AppConfig.Server.Host, config.AppConfig.Server.Port)
	go func() {
		logger.Info("Server starting", slog.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			logger.Error("Failed to start server", slog.Any("error", err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop cron scheduler and wait for running jobs
	cronSvc.Stop()

	// Graceful shutdown
	if err := app.Shutdown(); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	logger.Info("Server exited")
}

// registerCronJobs 注册所有定时任务
func registerCronJobs(cronSvc *service.CronService) {
	// 示例：每分钟执行一次的健康检查任务
	_ = cronSvc.AddJob("health-check", "0 * * * * *", func() {
		logger.Info("Health check cron job executed")
		// TODO: 在此添加实际的健康检查逻辑
	})

	// 示例：每天凌晨 2 点清理过期数据
	_ = cronSvc.AddJob("cleanup-expired-data", "0 0 2 * * *", func() {
		logger.Info("Cleanup expired data job executed")
		// TODO: 在此添加清理过期令牌、日志等逻辑
	})

	// 示例：每小时执行一次的统计任务
	_ = cronSvc.AddJob("hourly-stats", "0 0 * * * *", func() {
		logger.Info("Hourly stats job executed")
		// TODO: 在此添加统计逻辑
	})
}
