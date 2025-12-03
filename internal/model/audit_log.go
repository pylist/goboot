package model

import (
	"goboot/pkg/database"
	"time"
)

// AuditLog 操作审计日志
type AuditLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index"`                    // 操作用户ID，0表示未登录
	Username  string    `json:"username" gorm:"size:64"`                 // 操作用户名
	Action    string    `json:"action" gorm:"size:32;index"`             // 操作类型
	Module    string    `json:"module" gorm:"size:32;index"`             // 模块名称
	Target    string    `json:"target" gorm:"size:128"`                  // 操作目标（如被操作的用户ID）
	Detail    string    `json:"detail" gorm:"type:text"`                 // 操作详情
	IP        string    `json:"ip" gorm:"size:64"`                       // 客户端IP
	UserAgent string    `json:"user_agent" gorm:"size:256"`              // 客户端UA
	Status    int       `json:"status" gorm:"default:1"`                 // 状态：1成功 0失败
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// 操作类型常量
const (
	ActionLogin          = "login"          // 登录
	ActionLogout         = "logout"         // 登出
	ActionRegister       = "register"       // 注册
	ActionChangePassword = "change_pwd"     // 修改密码
	ActionResetPassword  = "reset_pwd"      // 重置密码
	ActionCreateUser     = "create_user"    // 创建用户
	ActionUpdateUser     = "update_user"    // 更新用户
	ActionDeleteUser     = "delete_user"    // 删除用户
	ActionUpdateStatus   = "update_status"  // 更新状态
	ActionUpload         = "upload"         // 上传文件
	ActionDelete         = "delete"         // 删除
	ActionCreate         = "create"         // 创建
	ActionUpdate         = "update"         // 更新
)

// 模块常量
const (
	ModuleAuth   = "auth"   // 认证模块
	ModuleUser   = "user"   // 用户模块
	ModuleAdmin  = "admin"  // 管理模块
	ModuleFile   = "file"   // 文件模块
	ModuleConfig = "config" // 配置模块
)

// CreateAuditLog 创建审计日志
func CreateAuditLog(log *AuditLog) error {
	return database.DB.Create(log).Error
}

// GetAuditLogs 获取审计日志列表
func GetAuditLogs(page, pageSize int, userID uint, action, module string, startTime, endTime *time.Time) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	db := database.DB.Model(&AuditLog{})

	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if action != "" {
		db = db.Where("action = ?", action)
	}
	if module != "" {
		db = db.Where("module = ?", module)
	}
	if startTime != nil {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != nil {
		db = db.Where("created_at <= ?", endTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
