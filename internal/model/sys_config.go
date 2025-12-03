package model

import (
	"time"

	"goboot/pkg/database"
)

// SysConfig 系统配置模型
type SysConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ConfigKey   string    `json:"configKey" gorm:"size:100;uniqueIndex;not null"` // 配置键
	ConfigValue string    `json:"configValue" gorm:"type:text"`                   // 配置值
	ConfigType  string    `json:"configType" gorm:"size:20;default:string"`       // 值类型: string, int, bool, json
	ConfigGroup string    `json:"configGroup" gorm:"size:50;index;default:basic"` // 配置分组
	Name        string    `json:"name" gorm:"size:100"`                           // 配置名称(中文)
	Remark      string    `json:"remark" gorm:"size:255"`                         // 备注说明
	Sort        int       `json:"sort" gorm:"default:0"`                          // 排序
	IsPublic    bool      `json:"isPublic" gorm:"default:false"`                  // 是否公开(前端可获取)
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// 配置分组常量
const (
	ConfigGroupBasic    = "basic"    // 基础配置
	ConfigGroupEmail    = "email"    // 邮件配置
	ConfigGroupUpload   = "upload"   // 上传配置
	ConfigGroupSecurity = "security" // 安全配置
)

// 配置类型常量
const (
	ConfigTypeString = "string"
	ConfigTypeInt    = "int"
	ConfigTypeBool   = "bool"
	ConfigTypeJSON   = "json"
)

// GetConfigByKey 根据key获取配置
func GetConfigByKey(key string) (*SysConfig, error) {
	var config SysConfig
	err := database.DB.Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetConfigsByGroup 根据分组获取配置列表
func GetConfigsByGroup(group string) ([]SysConfig, error) {
	var configs []SysConfig
	err := database.DB.Where("config_group = ?", group).Order("sort ASC, id ASC").Find(&configs).Error
	return configs, err
}

// GetAllConfigs 获取所有配置
func GetAllConfigs() ([]SysConfig, error) {
	var configs []SysConfig
	err := database.DB.Order("config_group ASC, sort ASC, id ASC").Find(&configs).Error
	return configs, err
}

// GetPublicConfigs 获取所有公开配置
func GetPublicConfigs() ([]SysConfig, error) {
	var configs []SysConfig
	err := database.DB.Where("is_public = ?", true).Order("config_group ASC, sort ASC").Find(&configs).Error
	return configs, err
}

// CreateConfig 创建配置
func CreateConfig(config *SysConfig) error {
	return database.DB.Create(config).Error
}

// UpdateConfig 更新配置
func UpdateConfig(config *SysConfig) error {
	return database.DB.Save(config).Error
}

// UpdateConfigValue 只更新配置值
func UpdateConfigValue(key, value string) error {
	return database.DB.Model(&SysConfig{}).Where("config_key = ?", key).Update("config_value", value).Error
}

// DeleteConfig 删除配置
func DeleteConfig(id uint) error {
	return database.DB.Delete(&SysConfig{}, id).Error
}

// BatchUpdateConfigs 批量更新配置值
func BatchUpdateConfigs(configs map[string]string) error {
	tx := database.DB.Begin()
	for key, value := range configs {
		if err := tx.Model(&SysConfig{}).Where("config_key = ?", key).Update("config_value", value).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// ConfigExists 检查配置是否存在
func ConfigExists(key string) bool {
	var count int64
	database.DB.Model(&SysConfig{}).Where("config_key = ?", key).Count(&count)
	return count > 0
}
