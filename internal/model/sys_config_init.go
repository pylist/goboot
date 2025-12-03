package model

import (
	"fmt"

	"goboot/pkg/database"
	"goboot/pkg/logger"
)

// 默认配置列表
var defaultConfigs = []SysConfig{
	// ============ 基础配置 ============
	{ConfigKey: "site_name", ConfigValue: "Goboot", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupBasic, Name: "网站名称", Remark: "网站显示名称", Sort: 1, IsPublic: true},
	{ConfigKey: "site_logo", ConfigValue: "", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupBasic, Name: "网站Logo", Remark: "网站Logo图片URL", Sort: 2, IsPublic: true},
	{ConfigKey: "site_description", ConfigValue: "基于Go的现代化Web框架", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupBasic, Name: "网站描述", Remark: "网站SEO描述", Sort: 3, IsPublic: true},
	{ConfigKey: "site_keywords", ConfigValue: "go,golang,fiber,web", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupBasic, Name: "网站关键词", Remark: "网站SEO关键词", Sort: 4, IsPublic: true},
	{ConfigKey: "site_icp", ConfigValue: "", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupBasic, Name: "ICP备案号", Remark: "网站ICP备案号", Sort: 5, IsPublic: true},

	// ============ 邮件配置 ============
	{ConfigKey: "email_enabled", ConfigValue: "false", ConfigType: ConfigTypeBool, ConfigGroup: ConfigGroupEmail, Name: "启用邮件服务", Remark: "是否启用邮件发送功能", Sort: 1, IsPublic: false},
	{ConfigKey: "email_host", ConfigValue: "smtp.qq.com", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupEmail, Name: "SMTP服务器", Remark: "SMTP服务器地址", Sort: 2, IsPublic: false},
	{ConfigKey: "email_port", ConfigValue: "465", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupEmail, Name: "SMTP端口", Remark: "SMTP端口号(SSL:465, TLS:587)", Sort: 3, IsPublic: false},
	{ConfigKey: "email_username", ConfigValue: "", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupEmail, Name: "邮箱账号", Remark: "发件邮箱账号", Sort: 4, IsPublic: false},
	{ConfigKey: "email_password", ConfigValue: "", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupEmail, Name: "邮箱密码", Remark: "邮箱密码或授权码", Sort: 5, IsPublic: false},
	{ConfigKey: "email_from_name", ConfigValue: "Goboot", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupEmail, Name: "发件人名称", Remark: "邮件显示的发件人名称", Sort: 6, IsPublic: false},
	{ConfigKey: "email_from_addr", ConfigValue: "", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupEmail, Name: "发件人地址", Remark: "发件人邮箱地址", Sort: 7, IsPublic: false},
	{ConfigKey: "email_ssl", ConfigValue: "true", ConfigType: ConfigTypeBool, ConfigGroup: ConfigGroupEmail, Name: "启用SSL", Remark: "是否使用SSL加密连接", Sort: 8, IsPublic: false},
	{ConfigKey: "email_reset_url", ConfigValue: "http://localhost:3000/reset-password", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupEmail, Name: "密码重置URL", Remark: "密码重置页面地址", Sort: 9, IsPublic: false},
	{ConfigKey: "email_reset_expire", ConfigValue: "30", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupEmail, Name: "重置链接有效期", Remark: "密码重置链接有效期(分钟)", Sort: 10, IsPublic: false},

	// ============ 上传配置 ============
	{ConfigKey: "upload_enabled", ConfigValue: "true", ConfigType: ConfigTypeBool, ConfigGroup: ConfigGroupUpload, Name: "启用上传服务", Remark: "是否启用文件上传功能", Sort: 1, IsPublic: false},
	{ConfigKey: "upload_storage_type", ConfigValue: "local", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupUpload, Name: "存储类型", Remark: "存储类型: local, oss, s3", Sort: 2, IsPublic: false},
	{ConfigKey: "upload_local_path", ConfigValue: "./uploads", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupUpload, Name: "本地存储路径", Remark: "本地文件存储目录", Sort: 3, IsPublic: false},
	{ConfigKey: "upload_base_url", ConfigValue: "http://127.0.0.1:8080/uploads", ConfigType: ConfigTypeString, ConfigGroup: ConfigGroupUpload, Name: "文件访问URL", Remark: "文件访问URL前缀", Sort: 4, IsPublic: false},
	{ConfigKey: "upload_max_size", ConfigValue: "10", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupUpload, Name: "最大文件大小", Remark: "最大上传文件大小(MB)", Sort: 5, IsPublic: false},
	{ConfigKey: "upload_max_image_size", ConfigValue: "5", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupUpload, Name: "最大图片大小", Remark: "最大上传图片大小(MB)", Sort: 6, IsPublic: false},
	{ConfigKey: "upload_allowed_exts", ConfigValue: `[".jpg",".jpeg",".png",".gif",".webp",".pdf",".doc",".docx",".xls",".xlsx",".zip",".rar"]`, ConfigType: ConfigTypeJSON, ConfigGroup: ConfigGroupUpload, Name: "允许的文件类型", Remark: "允许上传的文件扩展名", Sort: 7, IsPublic: false},
	{ConfigKey: "upload_image_exts", ConfigValue: `[".jpg",".jpeg",".png",".gif",".webp"]`, ConfigType: ConfigTypeJSON, ConfigGroup: ConfigGroupUpload, Name: "允许的图片类型", Remark: "允许上传的图片扩展名", Sort: 8, IsPublic: false},

	// ============ 安全配置 ============
	{ConfigKey: "security_max_login_attempts", ConfigValue: "5", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupSecurity, Name: "最大登录尝试", Remark: "登录失败最大尝试次数", Sort: 1, IsPublic: false},
	{ConfigKey: "security_lockout_duration", ConfigValue: "30", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupSecurity, Name: "锁定时长", Remark: "账户锁定时长(分钟)", Sort: 2, IsPublic: false},
	{ConfigKey: "security_password_min_length", ConfigValue: "6", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupSecurity, Name: "密码最小长度", Remark: "用户密码最小长度", Sort: 3, IsPublic: false},
	{ConfigKey: "security_session_timeout", ConfigValue: "120", ConfigType: ConfigTypeInt, ConfigGroup: ConfigGroupSecurity, Name: "会话超时", Remark: "用户会话超时时间(分钟)", Sort: 4, IsPublic: false},
}

// InitDefaultConfigs 初始化默认配置
// 只会插入不存在的配置项，不会覆盖已有配置
func InitDefaultConfigs() error {
	var insertCount int

	for _, cfg := range defaultConfigs {
		// 检查配置是否已存在
		if !ConfigExists(cfg.ConfigKey) {
			if err := database.DB.Create(&cfg).Error; err != nil {
				logger.Error("初始化配置失败: " + cfg.ConfigKey + " - " + err.Error())
				continue
			}
			insertCount++
		}
	}

	if insertCount > 0 {
		logger.Info(fmt.Sprintf("初始化系统配置完成，新增 %d 条配置", insertCount))
	}

	return nil
}

// ResetDefaultConfigs 重置为默认配置
// 警告: 这将覆盖所有已有配置
func ResetDefaultConfigs() error {
	// 删除所有配置
	if err := database.DB.Exec("DELETE FROM sys_configs").Error; err != nil {
		return err
	}

	// 重新插入默认配置
	for _, cfg := range defaultConfigs {
		if err := database.DB.Create(&cfg).Error; err != nil {
			logger.Error("重置配置失败: " + cfg.ConfigKey + " - " + err.Error())
			continue
		}
	}

	logger.Info("系统配置已重置为默认值")
	return nil
}
