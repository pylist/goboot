package router

import (
	"goboot/config"
	"goboot/internal/handler"
	"goboot/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	gin.SetMode(config.AppConfig.Server.Mode)

	r := gin.New()

	// 设置可信代理，避免安全警告
	// 生产环境应设置为实际的代理 IP，如: r.SetTrustedProxies([]string{"192.168.1.1"})
	r.SetTrustedProxies(nil) // nil 表示不信任任何代理，直接使用远程地址

	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())

	userHandler := handler.NewUserHandler()

	api := r.Group("/api")
	{
		// Public routes
		userAuth := api.Group("/auth")
		userAuth.POST("/register", userHandler.Register)
		userAuth.POST("/login", userHandler.Login)
		userAuth.POST("/refreshToken", userHandler.RefreshToken)
		userAuth.POST("/logout", userHandler.Logout)


		// User authenticated routes
		auth := api.Group("")
		auth.Use(middleware.JWTAuth())
		{
			// User
			auth.GET("/user/profile", userHandler.GetProfile)
			auth.POST("/user/updateProfile", userHandler.UpdateProfile)
			auth.POST("/user/changePassword", userHandler.ChangePassword)
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(), middleware.AdminAuth())
		{
			// User management
			admin.POST("/user/list", userHandler.AdminGetUserList)
			admin.POST("/user/add", userHandler.AdminCreateUser)
			admin.GET("/user/detail", userHandler.AdminGetUserDetail)
			admin.POST("/user/update", userHandler.AdminUpdateUser)
			admin.POST("/user/delete", userHandler.AdminDeleteUser)
			admin.POST("/user/resetPassword", userHandler.AdminResetPassword)
			admin.POST("/user/updateStatus", userHandler.AdminUpdateUserStatus)
		}
	}

	return r
}
