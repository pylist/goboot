package router

import (
	"goboot/internal/handler"
	"goboot/internal/middleware"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

func SetupRouter(app *fiber.App) {
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.Cors())
	app.Use(middleware.RateLimiter())

	// 静态文件服务(上传文件访问)
	app.Get("/uploads/*", static.New("./uploads"))

	// 健康检查接口
	app.Get("/ping", handler.Ping)
	app.Get("/health", handler.HealthCheck)

	userHandler := handler.NewUserHandler()
	auditHandler := handler.NewAuditHandler()
	emailHandler := handler.NewEmailHandler()
	uploadHandler := handler.NewUploadHandler()
	configHandler := handler.NewConfigHandler()

	api := app.Group("/api")

	// Public routes
	userAuth := api.Group("/auth")
	userAuth.Post("/register", userHandler.Register)
	userAuth.Post("/login", userHandler.Login)
	userAuth.Post("/refreshToken", userHandler.RefreshToken)
	userAuth.Post("/logout", userHandler.Logout)
	userAuth.Post("/forgotPassword", emailHandler.ForgotPassword)
	userAuth.Post("/resetPassword", emailHandler.ResetPassword)

	// 公开配置(无需登录)
	api.Get("/config/public", configHandler.GetPublicConfigs)

	// User authenticated routes
	auth := api.Group("", middleware.JWTAuth())
	auth.Get("/user/profile", userHandler.GetProfile)
	auth.Post("/user/updateProfile", userHandler.UpdateProfile)
	auth.Post("/user/changePassword", userHandler.ChangePassword)

	// Upload routes (需要登录)
	upload := auth.Group("/upload")
	upload.Post("/file", uploadHandler.UploadFile)
	upload.Post("/image", uploadHandler.UploadImage)
	upload.Post("/files", uploadHandler.UploadFiles)
	upload.Post("/delete", uploadHandler.DeleteFile)
	upload.Get("/info", uploadHandler.GetFileInfo)

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

	// Config management (系统配置管理)
	configAdmin := admin.Group("/config")
	configAdmin.Get("/list", configHandler.GetAllConfigs)
	configAdmin.Get("/group", configHandler.GetConfigsByGroup)
	configAdmin.Post("/add", configHandler.CreateConfig)
	configAdmin.Post("/update", configHandler.UpdateConfig)
	configAdmin.Post("/delete", configHandler.DeleteConfig)
	configAdmin.Post("/batchUpdate", configHandler.BatchUpdateConfigs)
	configAdmin.Post("/refresh", configHandler.RefreshCache)
	configAdmin.Get("/email", configHandler.GetEmailConfig)
	configAdmin.Post("/email", configHandler.UpdateEmailConfig)
}
