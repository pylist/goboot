package handler

import (
	"context"
	"goboot/pkg/database"
	"time"

	"github.com/gofiber/fiber/v3"
)

type HealthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// HealthCheck 健康检查接口，检查 MySQL 和 Redis 连接状态
func HealthCheck(c fiber.Ctx) error {
	status := HealthStatus{
		Status: "ok",
		Checks: make(map[string]string),
	}
	httpStatus := fiber.StatusOK

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 检查 MySQL
	sqlDB, err := database.DB.DB()
	if err != nil {
		status.Checks["mysql"] = "error: " + err.Error()
		status.Status = "error"
		httpStatus = fiber.StatusServiceUnavailable
	} else if err := sqlDB.PingContext(ctx); err != nil {
		status.Checks["mysql"] = "error: " + err.Error()
		status.Status = "error"
		httpStatus = fiber.StatusServiceUnavailable
	} else {
		status.Checks["mysql"] = "ok"
	}

	// 检查 Redis
	if err := database.RDB.Ping(ctx).Err(); err != nil {
		status.Checks["redis"] = "error: " + err.Error()
		status.Status = "error"
		httpStatus = fiber.StatusServiceUnavailable
	} else {
		status.Checks["redis"] = "ok"
	}

	return c.Status(httpStatus).JSON(status)
}

func Ping(c fiber.Ctx) error {
	return c.SendString("pong")
}
