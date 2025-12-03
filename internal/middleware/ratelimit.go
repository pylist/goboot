package middleware

import (
	"context"
	"fmt"
	"goboot/config"
	"goboot/pkg/database"
	"goboot/pkg/response"
	"time"

	"github.com/gofiber/fiber/v3"
)

// RateLimiter 基于 Redis 的滑动窗口限流中间件
func RateLimiter() fiber.Handler {
	return func(c fiber.Ctx) error {
		cfg := config.AppConfig.RateLimit
		if !cfg.Enabled {
			return c.Next()
		}

		// 获取限流 key（优先用户ID，否则用IP）
		key := getRateLimitKey(c)

		// 检查是否超过限制
		allowed, err := isAllowed(c, key, cfg.Requests, cfg.Window)
		if err != nil {
			// Redis 出错时放行，避免影响服务
			return c.Next()
		}

		if !allowed {
			return response.TooManyRequests(c, "请求过于频繁，请稍后再试")
		}

		return c.Next()
	}
}

// RateLimiterWithConfig 支持自定义限流参数
func RateLimiterWithConfig(requests int, window int) fiber.Handler {
	return func(c fiber.Ctx) error {
		key := getRateLimitKey(c)

		allowed, err := isAllowed(c, key, requests, window)
		if err != nil {
			return c.Next()
		}

		if !allowed {
			return response.TooManyRequests(c, "请求过于频繁，请稍后再试")
		}

		return c.Next()
	}
}

// getRateLimitKey 获取限流 key
func getRateLimitKey(c fiber.Ctx) string {
	// 优先使用用户ID（已登录用户）
	if userID := c.Locals("userID"); userID != nil {
		return fmt.Sprintf("ratelimit:user:%v:%s", userID, c.Path())
	}
	// 未登录使用 IP
	return fmt.Sprintf("ratelimit:ip:%s:%s", c.IP(), c.Path())
}

// isAllowed 使用滑动窗口算法检查是否允许请求
func isAllowed(c fiber.Ctx, key string, maxRequests int, windowSeconds int) (bool, error) {
	ctx := context.Background()
	now := time.Now().UnixMilli()
	window := int64(windowSeconds) * 1000

	pipe := database.RDB.Pipeline()

	// 移除窗口外的旧记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", now-window))

	// 统计当前窗口内的请求数
	countCmd := pipe.ZCard(ctx, key)

	// 添加当前请求
	pipe.ZAdd(ctx, key, database.Z{
		Score:  float64(now),
		Member: now,
	})

	// 设置 key 过期时间
	pipe.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := countCmd.Val()
	return count < int64(maxRequests), nil
}
