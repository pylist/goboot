package service

import (
	"context"
	"errors"
	"fmt"
	"goboot/config"
	"goboot/internal/model"
	"goboot/pkg/database"
	"goboot/pkg/utils"
	"time"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

func (s *UserService) Register(username, password, nickname, phone, email string) (*model.User, error) {
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Nickname: nickname,
		Phone:    phone,
		Email:    email,
		Status:   1,
		Role:     0,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, errors.New("注册失败")
	}

	return user, nil
}

func (s *UserService) Login(username, password string) (*utils.TokenPair, *model.User, error) {
	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, nil, errors.New("用户不存在")
	}

	if user.Status == 0 {
		return nil, nil, errors.New("账号已被禁用")
	}

	if !utils.CheckPassword(password, user.Password) {
		return nil, nil, errors.New("密码错误")
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, nil, errors.New("生成token失败")
	}

	return tokenPair, &user, nil
}

func (s *UserService) RefreshToken(refreshToken string) (*utils.TokenPair, error) {
	// 检查refresh token是否在黑名单
	if s.IsTokenBlacklisted(refreshToken) {
		return nil, errors.New("token已失效，请重新登录")
	}

	tokenPair, err := utils.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, errors.New("刷新token失败，请重新登录")
	}

	return tokenPair, nil
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func (s *UserService) UpdateProfile(id uint, nickname, phone, email, avatar string) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	updates := map[string]interface{}{}
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if phone != "" {
		updates["phone"] = phone
	}
	if email != "" {
		updates["email"] = email
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	if len(updates) > 0 {
		if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
			return nil, errors.New("更新失败")
		}
	}

	return &user, nil
}

func (s *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	if !utils.CheckPassword(oldPassword, user.Password) {
		return errors.New("原密码错误")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return errors.New("修改密码失败")
	}

	return nil
}

func tokenBlacklistKey(token string) string {
	return fmt.Sprintf("token:blacklist:%s", token)
}

func (s *UserService) Logout(userID uint, accessToken, refreshToken string) error {
	ctx := context.Background()
	cfg := config.AppConfig.JWT

	// 将access token加入黑名单
	accessExpiration := time.Duration(cfg.AccessExpire) * time.Hour
	if err := database.RDB.Set(ctx, tokenBlacklistKey(accessToken), userID, accessExpiration).Err(); err != nil {
		return errors.New("退出登录失败")
	}

	// 将refresh token加入黑名单
	if refreshToken != "" {
		refreshExpiration := time.Duration(cfg.RefreshExpire) * time.Hour
		if err := database.RDB.Set(ctx, tokenBlacklistKey(refreshToken), userID, refreshExpiration).Err(); err != nil {
			return errors.New("退出登录失败")
		}
	}

	return nil
}

func (s *UserService) IsTokenBlacklisted(token string) bool {
	ctx := context.Background()
	exists, _ := database.RDB.Exists(ctx, tokenBlacklistKey(token)).Result()
	return exists > 0
}

// ==================== 管理员用户管理 ====================

// AdminGetUserList 获取用户列表(管理员)
func (s *UserService) AdminGetUserList(page, pageSize int, username, phone, email string, status int8) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := database.DB.Model(&model.User{})

	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("获取用户列表失败")
	}

	offset := (page - 1) * pageSize
	if err := query.Order("id desc").Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, errors.New("获取用户列表失败")
	}

	return users, total, nil
}

// AdminCreateUser 创建用户(管理员)
func (s *UserService) AdminCreateUser(username, password, nickname, phone, email string, role int8, status int8) (*model.User, error) {
	var count int64
	database.DB.Model(&model.User{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Nickname: nickname,
		Phone:    phone,
		Email:    email,
		Status:   status,
		Role:     role,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, errors.New("创建用户失败")
	}

	return user, nil
}

// AdminUpdateUser 更新用户(管理员)
func (s *UserService) AdminUpdateUser(id uint, nickname, phone, email, avatar string, role int8, status int8) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	updates := map[string]interface{}{
		"nickname": nickname,
		"phone":    phone,
		"email":    email,
		"avatar":   avatar,
		"role":     role,
		"status":   status,
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, errors.New("更新用户失败")
	}

	return &user, nil
}

// AdminDeleteUser 删除用户(管理员)
func (s *UserService) AdminDeleteUser(id uint) error {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	// 不允许删除管理员
	if user.Role == 1 {
		return errors.New("不能删除管理员账号")
	}

	// 修改用户名，释放原用户名供重新注册
	deletedUsername := fmt.Sprintf("%s_deleted_%d", user.Username, time.Now().Unix())
	if err := database.DB.Model(&user).Update("username", deletedUsername).Error; err != nil {
		return errors.New("删除用户失败")
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		return errors.New("删除用户失败")
	}

	return nil
}

// AdminResetPassword 重置用户密码(管理员)
func (s *UserService) AdminResetPassword(id uint, newPassword string) error {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return errors.New("重置密码失败")
	}

	return nil
}

// AdminUpdateUserStatus 更新用户状态(管理员)
func (s *UserService) AdminUpdateUserStatus(id uint, status int8) error {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return errors.New("用户不存在")
	}

	if err := database.DB.Model(&user).Update("status", status).Error; err != nil {
		return errors.New("更新状态失败")
	}

	return nil
}
