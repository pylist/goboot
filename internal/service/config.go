package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"goboot/internal/model"
	"goboot/pkg/database"
	"goboot/pkg/logger"
)

// ConfigService 系统配置服务
type ConfigService struct {
	cache      map[string]*model.SysConfig // 内存缓存
	cacheMutex sync.RWMutex                // 缓存读写锁
}

var (
	configService *ConfigService
	configOnce    sync.Once
)

// GetConfigService 获取配置服务单例
func GetConfigService() *ConfigService {
	configOnce.Do(func() {
		configService = &ConfigService{
			cache: make(map[string]*model.SysConfig),
		}
		// 启动时加载所有配置到内存
		configService.LoadAll()
	})
	return configService
}

// LoadAll 加载所有配置到内存缓存
func (s *ConfigService) LoadAll() error {
	configs, err := model.GetAllConfigs()
	if err != nil {
		logger.Error("加载系统配置失败: " + err.Error())
		return err
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cache = make(map[string]*model.SysConfig)
	for i := range configs {
		s.cache[configs[i].ConfigKey] = &configs[i]
	}

	logger.Info(fmt.Sprintf("已加载 %d 条系统配置", len(configs)))
	return nil
}

// Refresh 刷新单个配置缓存
func (s *ConfigService) Refresh(key string) error {
	config, err := model.GetConfigByKey(key)
	if err != nil {
		// 配置不存在，从缓存中删除
		s.cacheMutex.Lock()
		delete(s.cache, key)
		s.cacheMutex.Unlock()
		return err
	}

	s.cacheMutex.Lock()
	s.cache[key] = config
	s.cacheMutex.Unlock()

	// 同时更新Redis缓存
	s.setRedisCache(key, config.ConfigValue)
	return nil
}

// RefreshGroup 刷新分组配置缓存
func (s *ConfigService) RefreshGroup(group string) error {
	configs, err := model.GetConfigsByGroup(group)
	if err != nil {
		return err
	}

	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	for i := range configs {
		s.cache[configs[i].ConfigKey] = &configs[i]
		s.setRedisCache(configs[i].ConfigKey, configs[i].ConfigValue)
	}
	return nil
}

// Get 获取配置值(字符串)
func (s *ConfigService) Get(key string, defaultValue ...string) string {
	s.cacheMutex.RLock()
	if config, ok := s.cache[key]; ok {
		s.cacheMutex.RUnlock()
		return config.ConfigValue
	}
	s.cacheMutex.RUnlock()

	// 缓存未命中，从数据库加载
	config, err := model.GetConfigByKey(key)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}

	// 更新缓存
	s.cacheMutex.Lock()
	s.cache[key] = config
	s.cacheMutex.Unlock()

	return config.ConfigValue
}

// GetString Get的别名
func (s *ConfigService) GetString(key string, defaultValue ...string) string {
	return s.Get(key, defaultValue...)
}

// GetInt 获取整数配置
func (s *ConfigService) GetInt(key string, defaultValue ...int) int {
	value := s.Get(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return intVal
}

// GetInt64 获取int64配置
func (s *ConfigService) GetInt64(key string, defaultValue ...int64) int64 {
	value := s.Get(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	intVal, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return intVal
}

// GetBool 获取布尔配置
func (s *ConfigService) GetBool(key string, defaultValue ...bool) bool {
	value := s.Get(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}

	// 支持多种格式: true/false, 1/0, yes/no, on/off
	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return false
	}
}

// GetJSON 获取JSON配置并解析到目标结构
func (s *ConfigService) GetJSON(key string, dest interface{}) error {
	value := s.Get(key)
	if value == "" {
		return errors.New("配置不存在")
	}
	return json.Unmarshal([]byte(value), dest)
}

// GetMap 获取JSON配置为map
func (s *ConfigService) GetMap(key string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.GetJSON(key, &result)
	return result, err
}

// Set 设置配置值
func (s *ConfigService) Set(key, value string) error {
	err := model.UpdateConfigValue(key, value)
	if err != nil {
		return err
	}

	// 刷新缓存
	return s.Refresh(key)
}

// SetInt 设置整数配置
func (s *ConfigService) SetInt(key string, value int) error {
	return s.Set(key, strconv.Itoa(value))
}

// SetBool 设置布尔配置
func (s *ConfigService) SetBool(key string, value bool) error {
	if value {
		return s.Set(key, "true")
	}
	return s.Set(key, "false")
}

// SetJSON 设置JSON配置
func (s *ConfigService) SetJSON(key string, value interface{}) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.Set(key, string(jsonBytes))
}

// GetByGroup 获取分组配置列表
func (s *ConfigService) GetByGroup(group string) ([]model.SysConfig, error) {
	return model.GetConfigsByGroup(group)
}

// GetAll 获取所有配置
func (s *ConfigService) GetAll() ([]model.SysConfig, error) {
	return model.GetAllConfigs()
}

// GetPublic 获取所有公开配置
func (s *ConfigService) GetPublic() ([]model.SysConfig, error) {
	return model.GetPublicConfigs()
}

// Create 创建配置
func (s *ConfigService) Create(config *model.SysConfig) error {
	if model.ConfigExists(config.ConfigKey) {
		return errors.New("配置键已存在")
	}

	err := model.CreateConfig(config)
	if err != nil {
		return err
	}

	// 更新缓存
	s.cacheMutex.Lock()
	s.cache[config.ConfigKey] = config
	s.cacheMutex.Unlock()

	return nil
}

// Update 更新配置
func (s *ConfigService) Update(config *model.SysConfig) error {
	err := model.UpdateConfig(config)
	if err != nil {
		return err
	}

	// 刷新缓存
	return s.Refresh(config.ConfigKey)
}

// Delete 删除配置
func (s *ConfigService) Delete(id uint) error {
	// 先获取配置key
	var config model.SysConfig
	if err := database.DB.First(&config, id).Error; err != nil {
		return err
	}

	err := model.DeleteConfig(id)
	if err != nil {
		return err
	}

	// 从缓存中删除
	s.cacheMutex.Lock()
	delete(s.cache, config.ConfigKey)
	s.cacheMutex.Unlock()

	// 删除Redis缓存
	s.deleteRedisCache(config.ConfigKey)

	return nil
}

// BatchUpdate 批量更新配置值
func (s *ConfigService) BatchUpdate(configs map[string]string) error {
	err := model.BatchUpdateConfigs(configs)
	if err != nil {
		return err
	}

	// 刷新缓存
	for key := range configs {
		s.Refresh(key)
	}
	return nil
}

// setRedisCache 设置Redis缓存
func (s *ConfigService) setRedisCache(key, value string) {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	cacheKey := "sys_config:" + key
	database.RDB.Set(ctx, cacheKey, value, 24*time.Hour)
}

// deleteRedisCache 删除Redis缓存
func (s *ConfigService) deleteRedisCache(key string) {
	if database.RDB == nil {
		return
	}
	ctx := context.Background()
	cacheKey := "sys_config:" + key
	database.RDB.Del(ctx, cacheKey)
}

// ============ 邮件配置便捷方法 ============

// EmailConfig 邮件配置结构
type EmailConfig struct {
	Enabled     bool
	Host        string
	Port        int
	Username    string
	Password    string
	FromName    string
	FromAddr    string
	SSL         bool
	ResetURL    string
	ResetExpire int
}

// GetEmailConfig 获取邮件配置
func (s *ConfigService) GetEmailConfig() *EmailConfig {
	return &EmailConfig{
		Enabled:     s.GetBool("email_enabled", false),
		Host:        s.Get("email_host", ""),
		Port:        s.GetInt("email_port", 465),
		Username:    s.Get("email_username", ""),
		Password:    s.Get("email_password", ""),
		FromName:    s.Get("email_from_name", "Goboot"),
		FromAddr:    s.Get("email_from_addr", ""),
		SSL:         s.GetBool("email_ssl", true),
		ResetURL:    s.Get("email_reset_url", ""),
		ResetExpire: s.GetInt("email_reset_expire", 30),
	}
}

// ============ 上传配置便捷方法 ============

// UploadConfig 上传配置结构
type UploadConfigDB struct {
	Enabled      bool
	StorageType  string
	LocalPath    string
	BaseURL      string
	MaxSize      int
	MaxImageSize int
	AllowedExts  []string
	ImageExts    []string
}

// GetUploadConfig 获取上传配置
func (s *ConfigService) GetUploadConfig() *UploadConfigDB {
	var allowedExts, imageExts []string
	s.GetJSON("upload_allowed_exts", &allowedExts)
	s.GetJSON("upload_image_exts", &imageExts)

	// 默认值
	if len(allowedExts) == 0 {
		allowedExts = []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}
	}
	if len(imageExts) == 0 {
		imageExts = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	}

	return &UploadConfigDB{
		Enabled:      s.GetBool("upload_enabled", true),
		StorageType:  s.Get("upload_storage_type", "local"),
		LocalPath:    s.Get("upload_local_path", "./uploads"),
		BaseURL:      s.Get("upload_base_url", "http://127.0.0.1:8080/uploads"),
		MaxSize:      s.GetInt("upload_max_size", 10),
		MaxImageSize: s.GetInt("upload_max_image_size", 5),
		AllowedExts:  allowedExts,
		ImageExts:    imageExts,
	}
}
