package handler

import (
	"goboot/internal/service"
	"goboot/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(),
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.Register(req.Username, req.Password, req.Nickname, req.Phone, req.Email)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "注册成功", user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	tokenPair, user, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
		"user":         user,
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	tokenPair, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"accessToken":  tokenPair.AccessToken,
		"refreshToken": tokenPair.RefreshToken,
		"expiresIn":    tokenPair.ExpiresIn,
	})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, user)
}

type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(userID, req.Nickname, req.Phone, req.Email, req.Avatar)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, user)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("userID")
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码修改成功", nil)
}

type LogoutRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (h *UserHandler) Logout(c *gin.Context) {
	userID := c.GetUint("userID")

	// 获取access token
	accessToken := c.GetHeader("Authorization")
	if len(accessToken) > 7 {
		accessToken = accessToken[7:] // 去掉 "Bearer "
	}

	// 获取refresh token
	var req LogoutRequest
	c.ShouldBindJSON(&req)

	if err := h.userService.Logout(userID, accessToken, req.RefreshToken); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "退出成功", nil)
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
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=20"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Role     int8   `json:"role"`
	Status   int8   `json:"status"`
}

type AdminUpdateUserRequest struct {
	ID       uint   `json:"id" binding:"required"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Role     int8   `json:"role"`
	Status   int8   `json:"status"`
}

type AdminUserIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

type AdminResetPasswordRequest struct {
	ID          uint   `json:"id" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6,max=20"`
}

type AdminUpdateStatusRequest struct {
	ID     uint `json:"id" binding:"required"`
	Status int8 `json:"status"`
}

// AdminGetUserList 获取用户列表
func (h *UserHandler) AdminGetUserList(c *gin.Context) {
	var req AdminUserListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
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
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithPage(c, users, total, req.Page, req.PageSize)
}

// AdminCreateUser 创建用户
func (h *UserHandler) AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 默认状态为启用
	if req.Status == 0 {
		req.Status = 1
	}

	user, err := h.userService.AdminCreateUser(req.Username, req.Password, req.Nickname, req.Phone, req.Email, req.Role, req.Status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, user)
}

// AdminUpdateUser 更新用户
func (h *UserHandler) AdminUpdateUser(c *gin.Context) {
	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.userService.AdminUpdateUser(req.ID, req.Nickname, req.Phone, req.Email, req.Avatar, req.Role, req.Status)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, user)
}

// AdminDeleteUser 删除用户
func (h *UserHandler) AdminDeleteUser(c *gin.Context) {
	var req AdminUserIDRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.AdminDeleteUser(req.ID); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// AdminGetUserDetail 获取用户详情
func (h *UserHandler) AdminGetUserDetail(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		response.Fail(c, "参数错误: id必须为有效数字")
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, user)
}

// AdminResetPassword 重置用户密码
func (h *UserHandler) AdminResetPassword(c *gin.Context) {
	var req AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.AdminResetPassword(req.ID, req.NewPassword); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "密码重置成功", nil)
}

// AdminUpdateUserStatus 更新用户状态
func (h *UserHandler) AdminUpdateUserStatus(c *gin.Context) {
	var req AdminUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if err := h.userService.AdminUpdateUserStatus(req.ID, req.Status); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "状态更新成功", nil)
}
