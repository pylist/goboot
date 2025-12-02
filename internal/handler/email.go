package handler

import (
	"goboot/internal/model"
	"goboot/internal/service"
	"goboot/pkg/response"

	"github.com/gin-gonic/gin"
)

type EmailHandler struct {
	emailService *service.EmailService
	userService  *service.UserService
	auditService *service.AuditService
}

func NewEmailHandler() *EmailHandler {
	return &EmailHandler{
		emailService: service.NewEmailService(),
		userService:  service.NewUserService(),
		auditService: service.NewAuditService(),
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
}

// ForgotPassword 忘记密码，发送重置邮件
func (h *EmailHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 根据邮箱查找用户
	user, err := h.userService.GetUserByEmail(req.Email)
	if err != nil {
		// 为了安全，不暴露用户是否存在
		response.SuccessWithMessage(c, "如果该邮箱已注册，您将收到密码重置邮件", nil)
		return
	}

	// 发送重置邮件
	if err := h.emailService.SendPasswordResetEmail(user.Email, user.Username, user.ID); err != nil {
		response.Fail(c, "发送邮件失败，请稍后重试")
		return
	}

	response.SuccessWithMessage(c, "如果该邮箱已注册，您将收到密码重置邮件", nil)
}

// ResetPassword 重置密码
func (h *EmailHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 验证 token
	userID, err := h.emailService.VerifyResetToken(req.Token)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	// 重置密码
	if err := h.userService.AdminResetPassword(userID, req.NewPassword); err != nil {
		response.Fail(c, "重置密码失败: "+err.Error())
		return
	}

	// 删除已使用的 token
	h.emailService.DeleteResetToken(req.Token)

	// 记录审计日志
	h.auditService.LogSuccess(c, model.ActionResetPassword, model.ModuleAuth, "", "用户通过邮件重置密码")

	response.SuccessWithMessage(c, "密码重置成功", nil)
}
