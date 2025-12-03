package middleware

import (
	"goboot/internal/service"
	"goboot/pkg/response"
	"goboot/pkg/utils"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var userService = service.NewUserService()

func JWTAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Unauthorized(c, "请先登录")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			return response.Unauthorized(c, "无效的认证格式")
		}

		token := parts[1]

		// 检查token是否在黑名单中
		if userService.IsTokenBlacklisted(token) {
			return response.Unauthorized(c, "token已失效，请重新登录")
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			return response.Unauthorized(c, "无效的token")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

func AdminAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		role := c.Locals("role")
		if role == nil {
			return response.Unauthorized(c, "请先登录")
		}

		if role.(int8) != 1 {
			return response.Forbidden(c, "无权限访问")
		}

		return c.Next()
	}
}
