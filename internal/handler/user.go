package handler

import (
	"fmt"
	"goboot/internal/model"
	"goboot/internal/service"
	"goboot/pkg/response"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type UserHandler struct {
	userService  *service.UserService
	auditService *service.AuditService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService:  service.NewUserService(),
		auditService: service.NewAuditService(),
	}
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=20"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.Username == "" || len(req.Username) < 3 || len(req.Username) > 50 {
		return response.Fail(c, "参数错误: 用户名长度必须在3-50位之间")
	}
	if req.Password == "" || len(req.Password) < 6 || len(req.Password) > 20 {
		return response.Fail(c, "参数错误: 密码长度必须在6-20位之间")
	}

	user, err := h.userService.Register(req.Username, req.Password, req.Nickname, req.Phone, req.Email)
	if err != nil {
		h.auditService.LogFail(c, model.ActionRegister, model.ModuleAuth, req.Username, err.Error())
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionRegister, model.ModuleAuth, req.Username, "用户注册成功")
	return response.SuccessWithMessage(c, "注册成功", user)
}

func (h *UserHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.Username == "" || req.Password == "" {
		return response.Fail(c, "参数错误: 用户名和密码不能为空")
	}

	tokenPair, user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		h.auditService.LogFail(c, model.ActionLogin, model.ModuleAuth, req.Username, err.Error())
		return response.Fail(c, err.Error())
	}

	// 登录成功后设置用户信息用于审计日志
	c.Locals("userID", user.ID)
	c.Locals("username", user.Username)
	h.auditService.LogSuccess(c, model.ActionLogin, model.ModuleAuth, req.Username, "用户登录成功")

	return response.Success(c, fiber.Map{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
		"user":         user,
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (h *UserHandler) RefreshToken(c fiber.Ctx) error {
	var req RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.RefreshToken == "" {
		return response.Fail(c, "参数错误: refreshToken不能为空")
	}

	tokenPair, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, err.Error())
	}

	return response.Success(c, fiber.Map{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
	})
}

func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return response.Fail(c, err.Error())
	}

	return response.Success(c, user)
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	user, err := h.userService.UpdateProfile(userID, req.Nickname, req.Phone, req.Email, req.Avatar)
	if err != nil {
		return response.Fail(c, err.Error())
	}

	return response.Success(c, user)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=6,max=20"`
}

func (h *UserHandler) ChangePassword(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	var req ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.OldPassword == "" {
		return response.Fail(c, "参数错误: 原密码不能为空")
	}
	if len(req.NewPassword) < 6 || len(req.NewPassword) > 20 {
		return response.Fail(c, "参数错误: 新密码长度必须在6-20位之间")
	}

	err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		h.auditService.LogFail(c, model.ActionChangePassword, model.ModuleUser, fmt.Sprintf("%d", userID), err.Error())
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionChangePassword, model.ModuleUser, fmt.Sprintf("%d", userID), "用户修改密码")
	return response.SuccessWithMessage(c, "密码修改成功", nil)
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *UserHandler) Logout(c fiber.Ctx) error {
	userID, _ := c.Locals("userID").(uint)

	// 获取access token
	accessToken := c.Get("Authorization")
	if len(accessToken) > 7 {
		accessToken = accessToken[7:] // 去掉 "Bearer "
	}

	// 获取refresh token
	var req LogoutRequest
	_ = c.Bind().Body(&req)

	if err := h.userService.Logout(userID, accessToken, req.RefreshToken); err != nil {
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionLogout, model.ModuleAuth, fmt.Sprintf("%d", userID), "用户退出登录")
	return response.SuccessWithMessage(c, "退出成功", nil)
}

// ==================== 管理员用户管理 ====================

type AdminUserListRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Status   int8   `json:"status"`
}

type AdminCreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=20"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Role     int8   `json:"role"`
	Status   int8   `json:"status"`
}

type AdminUpdateUserRequest struct {
	ID       uint   `json:"id" validate:"required"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Role     int8   `json:"role"`
	Status   int8   `json:"status"`
}

type AdminUserIDRequest struct {
	ID uint `json:"id" validate:"required"`
}

type AdminResetPasswordRequest struct {
	ID          uint   `json:"id" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=6,max=20"`
}

type AdminUpdateStatusRequest struct {
	ID     uint `json:"id" validate:"required"`
	Status int8 `json:"status"`
}

// AdminGetUserList 获取用户列表
func (h *UserHandler) AdminGetUserList(c fiber.Ctx) error {
	var req AdminUserListRequest
	if err := c.Bind().Body(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
		req.Status = -1
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	users, total, err := h.userService.AdminGetUserList(req.Page, req.PageSize, req.Username, req.Phone, req.Email, req.Status)
	if err != nil {
		return response.Fail(c, err.Error())
	}

	return response.SuccessWithPage(c, users, total, req.Page, req.PageSize)
}

// AdminCreateUser 创建用户
func (h *UserHandler) AdminCreateUser(c fiber.Ctx) error {
	var req AdminCreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.Username == "" || len(req.Username) < 3 || len(req.Username) > 50 {
		return response.Fail(c, "参数错误: 用户名长度必须在3-50位之间")
	}
	if req.Password == "" || len(req.Password) < 6 || len(req.Password) > 20 {
		return response.Fail(c, "参数错误: 密码长度必须在6-20位之间")
	}

	// 默认状态为启用
	if req.Status == 0 {
		req.Status = 1
	}

	user, err := h.userService.AdminCreateUser(req.Username, req.Password, req.Nickname, req.Phone, req.Email, req.Role, req.Status)
	if err != nil {
		h.auditService.LogFail(c, model.ActionCreateUser, model.ModuleAdmin, req.Username, err.Error())
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionCreateUser, model.ModuleAdmin, req.Username, fmt.Sprintf("创建用户: %s", req.Username))
	return response.Success(c, user)
}

// AdminUpdateUser 更新用户
func (h *UserHandler) AdminUpdateUser(c fiber.Ctx) error {
	var req AdminUpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ID == 0 {
		return response.Fail(c, "参数错误: id不能为空")
	}

	user, err := h.userService.AdminUpdateUser(req.ID, req.Nickname, req.Phone, req.Email, req.Avatar, req.Role, req.Status)
	if err != nil {
		h.auditService.LogFail(c, model.ActionUpdateUser, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), err.Error())
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionUpdateUser, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), fmt.Sprintf("更新用户ID: %d", req.ID))
	return response.Success(c, user)
}

// AdminDeleteUser 删除用户
func (h *UserHandler) AdminDeleteUser(c fiber.Ctx) error {
	var req AdminUserIDRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ID == 0 {
		return response.Fail(c, "参数错误: id不能为空")
	}

	if err := h.userService.AdminDeleteUser(req.ID); err != nil {
		h.auditService.LogFail(c, model.ActionDeleteUser, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), err.Error())
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionDeleteUser, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), fmt.Sprintf("删除用户ID: %d", req.ID))
	return response.SuccessWithMessage(c, "删除成功", nil)
}

// AdminGetUserDetail 获取用户详情
func (h *UserHandler) AdminGetUserDetail(c fiber.Ctx) error {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return response.Fail(c, "参数错误: id必须为有效数字")
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		return response.Fail(c, err.Error())
	}

	return response.Success(c, user)
}

// AdminResetPassword 重置用户密码
func (h *UserHandler) AdminResetPassword(c fiber.Ctx) error {
	var req AdminResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ID == 0 {
		return response.Fail(c, "参数错误: id不能为空")
	}
	if len(req.NewPassword) < 6 || len(req.NewPassword) > 20 {
		return response.Fail(c, "参数错误: 新密码长度必须在6-20位之间")
	}

	if err := h.userService.AdminResetPassword(req.ID, req.NewPassword); err != nil {
		h.auditService.LogFail(c, model.ActionResetPassword, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), err.Error())
		return response.Fail(c, err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionResetPassword, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), fmt.Sprintf("重置用户密码ID: %d", req.ID))
	return response.SuccessWithMessage(c, "密码重置成功", nil)
}

// AdminUpdateUserStatus 更新用户状态
func (h *UserHandler) AdminUpdateUserStatus(c fiber.Ctx) error {
	var req AdminUpdateStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ID == 0 {
		return response.Fail(c, "参数错误: id不能为空")
	}

	if err := h.userService.AdminUpdateUserStatus(req.ID, req.Status); err != nil {
		h.auditService.LogFail(c, model.ActionUpdateStatus, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), err.Error())
		return response.Fail(c, err.Error())
	}

	statusText := "禁用"
	if req.Status == 1 {
		statusText = "启用"
	}
	h.auditService.LogSuccess(c, model.ActionUpdateStatus, model.ModuleAdmin, fmt.Sprintf("%d", req.ID), fmt.Sprintf("更新用户状态为%s, ID: %d", statusText, req.ID))
	return response.SuccessWithMessage(c, "状态更新成功", nil)
}
