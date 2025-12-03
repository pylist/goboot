package router

import (
	"goboot/internal/handler"
	"goboot/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

func SetupRouter(app *fiber.App) {
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.Cors())
	app.Use(middleware.RateLimiter())

	// 健康检查接口
	app.Get("/ping", handler.Ping)
	app.Get("/health", handler.HealthCheck)

	userHandler := handler.NewUserHandler()
	auditHandler := handler.NewAuditHandler()
	emailHandler := handler.NewEmailHandler()

	api := app.Group("/api")

	// Public routes
	userAuth := api.Group("/auth")
	userAuth.Post("/register", userHandler.Register)
	userAuth.Post("/login", userHandler.Login)
	userAuth.Post("/refreshToken", userHandler.RefreshToken)
	userAuth.Post("/logout", userHandler.Logout)
	userAuth.Post("/forgotPassword", emailHandler.ForgotPassword)
	userAuth.Post("/resetPassword", emailHandler.ResetPassword)

	// User authenticated routes
	auth := api.Group("", middleware.JWTAuth())
	auth.Get("/user/profile", userHandler.GetProfile)
	auth.Post("/user/updateProfile", userHandler.UpdateProfile)
	auth.Post("/user/changePassword", userHandler.ChangePassword)

	// Admin routes
	admin := api.Group("/admin", middleware.JWTAuth(), middleware.AdminAuth())
	// User management
	admin.Post("/user/list", userHandler.AdminGetUserList)
	admin.Post("/user/add", userHandler.AdminCreateUser)
	admin.Get("/user/detail", userHandler.AdminGetUserDetail)
	admin.Post("/user/update", userHandler.AdminUpdateUser)
	admin.Post("/user/delete", userHandler.AdminDeleteUser)
	admin.Post("/user/resetPassword", userHandler.AdminResetPassword)
	admin.Post("/user/updateStatus", userHandler.AdminUpdateUserStatus)

	// Audit log
	admin.Post("/audit/list", auditHandler.GetAuditLogs)
}
