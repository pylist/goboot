package middleware

import (
	"fmt"
	"goboot/config"
	"goboot/pkg/database"
	"goboot/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于 Redis 的滑动窗口限流中间件
func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := config.AppConfig.RateLimit
		if !cfg.Enabled {
			c.Next()
			return
		}

		// 获取限流 key（优先用户ID，否则用IP）
		key := getRateLimitKey(c)

		// 检查是否超过限制
		allowed, err := isAllowed(c, key, cfg.Requests, cfg.Window)
		if err != nil {
			// Redis 出错时放行，避免影响服务
			c.Next()
			return
		}

		if !allowed {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiterWithConfig 支持自定义限流参数
func RateLimiterWithConfig(requests int, window int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := getRateLimitKey(c)

		allowed, err := isAllowed(c, key, requests, window)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		c.Next()
	}
}

// getRateLimitKey 获取限流 key
func getRateLimitKey(c *gin.Context) string {
	// 优先使用用户ID（已登录用户）
	if userID, exists := c.Get("userID"); exists {
		return fmt.Sprintf("ratelimit:user:%v:%s", userID, c.FullPath())
	}
	// 未登录使用 IP
	return fmt.Sprintf("ratelimit:ip:%s:%s", c.ClientIP(), c.FullPath())
}

// isAllowed 使用滑动窗口算法检查是否允许请求
func isAllowed(c *gin.Context, key string, maxRequests int, windowSeconds int) (bool, error) {
	ctx := c.Request.Context()
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
