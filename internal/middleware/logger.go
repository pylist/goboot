package middleware

import (
	"goboot/pkg/logger"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		attrs := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("user-agent", userAgent),
			slog.String("latency", latency.String()),
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()))
			logger.Error("Request error", attrs...)
		} else if status >= 500 {
			logger.Error("Server error", attrs...)
		} else if status >= 400 {
			logger.Warn("Client error", attrs...)
		} else {
			logger.Info("Request", attrs...)
		}
	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					slog.Any("error", err),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
