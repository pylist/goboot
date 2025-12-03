package middleware

import (
	"goboot/pkg/logger"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		path := c.Path()
		query := string(c.Request().URI().QueryString())

		err := c.Next()

		latency := time.Since(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Second)
		}
		status := c.Response().StatusCode()
		clientIP := c.IP()
		method := c.Method()
		userAgent := string(c.Request().Header.UserAgent())

		attrs := []any{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("query", query),
			slog.String("ip", clientIP),
			slog.String("user-agent", userAgent),
			slog.String("latency", latency.String()),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			logger.Error("Request error", attrs...)
		} else if status >= 500 {
			logger.Error("Server error", attrs...)
		} else if status >= 400 {
			logger.Warn("Client error", attrs...)
		} else {
			logger.Info("Request", attrs...)
		}

		return err
	}
}

func Recovery() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					slog.Any("error", err),
					slog.String("path", c.Path()),
					slog.String("method", c.Method()),
				)
				_ = c.SendStatus(fiber.StatusInternalServerError)
			}
		}()
		return c.Next()
	}
}
