package handler

import (
	"context"
	"goboot/pkg/database"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// HealthCheck 健康检查接口，检查 MySQL 和 Redis 连接状态
func HealthCheck(c *gin.Context) {
	status := HealthStatus{
		Status: "ok",
		Checks: make(map[string]string),
	}
	httpStatus := http.StatusOK

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 检查 MySQL
	sqlDB, err := database.DB.DB()
	if err != nil {
		status.Checks["mysql"] = "error: " + err.Error()
		status.Status = "error"
		httpStatus = http.StatusServiceUnavailable
	} else if err := sqlDB.PingContext(ctx); err != nil {
		status.Checks["mysql"] = "error: " + err.Error()
		status.Status = "error"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status.Checks["mysql"] = "ok"
	}

	// 检查 Redis
	if err := database.RDB.Ping(ctx).Err(); err != nil {
		status.Checks["redis"] = "error: " + err.Error()
		status.Status = "error"
		httpStatus = http.StatusServiceUnavailable
	} else {
		status.Checks["redis"] = "ok"
	}

	c.JSON(httpStatus, status)
}
