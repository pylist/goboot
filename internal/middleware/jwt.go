package middleware

import (
	"goboot/internal/service"
	"goboot/pkg/response"
	"goboot/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

var userService = service.NewUserService()

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "无效的认证格式")
			c.Abort()
			return
		}

		token := parts[1]

		// 检查token是否在黑名单中
		if userService.IsTokenBlacklisted(token) {
			response.Unauthorized(c, "token已失效，请重新登录")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(token)
		if err != nil {
			response.Unauthorized(c, "无效的token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		if role.(int8) != 1 {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}
