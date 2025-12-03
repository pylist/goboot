package handler

import (
	"fmt"

	"goboot/internal/model"
	"goboot/internal/service"
	"goboot/pkg/response"

	"github.com/gofiber/fiber/v3"
)

type ConfigHandler struct {
	configService *service.ConfigService
	auditService  *service.AuditService
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configService: service.GetConfigService(),
		auditService:  service.NewAuditService(),
	}
}

// GetAllConfigs 获取所有配置(管理员)
func (h *ConfigHandler) GetAllConfigs(c fiber.Ctx) error {
	configs, err := h.configService.GetAll()
	if err != nil {
		return response.Fail(c, "获取配置失败: "+err.Error())
	}

	// 按分组整理
	grouped := make(map[string][]model.SysConfig)
	for _, cfg := range configs {
		grouped[cfg.ConfigGroup] = append(grouped[cfg.ConfigGroup], cfg)
	}

	return response.Success(c, grouped)
}

// GetConfigsByGroup 按分组获取配置
func (h *ConfigHandler) GetConfigsByGroup(c fiber.Ctx) error {
	group := c.Query("group")
	if group == "" {
		return response.Fail(c, "分组参数不能为空")
	}

	configs, err := h.configService.GetByGroup(group)
	if err != nil {
		return response.Fail(c, "获取配置失败: "+err.Error())
	}

	return response.Success(c, configs)
}

// GetPublicConfigs 获取公开配置(无需登录)
func (h *ConfigHandler) GetPublicConfigs(c fiber.Ctx) error {
	configs, err := h.configService.GetPublic()
	if err != nil {
		return response.Fail(c, "获取配置失败: "+err.Error())
	}

	// 转换为简单的key-value格式
	result := make(map[string]string)
	for _, cfg := range configs {
		result[cfg.ConfigKey] = cfg.ConfigValue
	}

	return response.Success(c, result)
}

// CreateConfigRequest 创建配置请求
type CreateConfigRequest struct {
	ConfigKey   string `json:"configKey" validate:"required"`
	ConfigValue string `json:"configValue"`
	ConfigType  string `json:"configType"`
	ConfigGroup string `json:"configGroup"`
	Name        string `json:"name"`
	Remark      string `json:"remark"`
	Sort        int    `json:"sort"`
	IsPublic    bool   `json:"isPublic"`
}

// CreateConfig 创建配置
func (h *ConfigHandler) CreateConfig(c fiber.Ctx) error {
	var req CreateConfigRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ConfigKey == "" {
		return response.Fail(c, "配置键不能为空")
	}

	// 设置默认值
	if req.ConfigType == "" {
		req.ConfigType = model.ConfigTypeString
	}
	if req.ConfigGroup == "" {
		req.ConfigGroup = model.ConfigGroupBasic
	}

	config := &model.SysConfig{
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType:  req.ConfigType,
		ConfigGroup: req.ConfigGroup,
		Name:        req.Name,
		Remark:      req.Remark,
		Sort:        req.Sort,
		IsPublic:    req.IsPublic,
	}

	if err := h.configService.Create(config); err != nil {
		h.auditService.LogFail(c, model.ActionCreate, model.ModuleConfig, req.ConfigKey, err.Error())
		return response.Fail(c, "创建配置失败: "+err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionCreate, model.ModuleConfig, req.ConfigKey, "创建系统配置")
	return response.Success(c, config)
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	ID          uint   `json:"id" validate:"required"`
	ConfigKey   string `json:"configKey"`
	ConfigValue string `json:"configValue"`
	ConfigType  string `json:"configType"`
	ConfigGroup string `json:"configGroup"`
	Name        string `json:"name"`
	Remark      string `json:"remark"`
	Sort        int    `json:"sort"`
	IsPublic    bool   `json:"isPublic"`
}

// UpdateConfig 更新配置
func (h *ConfigHandler) UpdateConfig(c fiber.Ctx) error {
	var req UpdateConfigRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ID == 0 {
		return response.Fail(c, "配置ID不能为空")
	}

	config := &model.SysConfig{
		ID:          req.ID,
		ConfigKey:   req.ConfigKey,
		ConfigValue: req.ConfigValue,
		ConfigType:  req.ConfigType,
		ConfigGroup: req.ConfigGroup,
		Name:        req.Name,
		Remark:      req.Remark,
		Sort:        req.Sort,
		IsPublic:    req.IsPublic,
	}

	if err := h.configService.Update(config); err != nil {
		h.auditService.LogFail(c, model.ActionUpdate, model.ModuleConfig, req.ConfigKey, err.Error())
		return response.Fail(c, "更新配置失败: "+err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionUpdate, model.ModuleConfig, req.ConfigKey, "更新系统配置")
	return response.SuccessWithMessage(c, "更新成功", config)
}

// BatchUpdateRequest 批量更新请求
type BatchUpdateRequest struct {
	Configs map[string]string `json:"configs" validate:"required"`
}

// BatchUpdateConfigs 批量更新配置值
func (h *ConfigHandler) BatchUpdateConfigs(c fiber.Ctx) error {
	var req BatchUpdateRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if len(req.Configs) == 0 {
		return response.Fail(c, "配置数据不能为空")
	}

	if err := h.configService.BatchUpdate(req.Configs); err != nil {
		h.auditService.LogFail(c, model.ActionUpdate, model.ModuleConfig, "", err.Error())
		return response.Fail(c, "批量更新失败: "+err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionUpdate, model.ModuleConfig, "", "批量更新系统配置")
	return response.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteConfigRequest 删除配置请求
type DeleteConfigRequest struct {
	ID uint `json:"id" validate:"required"`
}

// DeleteConfig 删除配置
func (h *ConfigHandler) DeleteConfig(c fiber.Ctx) error {
	var req DeleteConfigRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	if req.ID == 0 {
		return response.Fail(c, "配置ID不能为空")
	}

	if err := h.configService.Delete(req.ID); err != nil {
		h.auditService.LogFail(c, model.ActionDelete, model.ModuleConfig, "", err.Error())
		return response.Fail(c, "删除配置失败: "+err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionDelete, model.ModuleConfig, "", "删除系统配置")
	return response.SuccessWithMessage(c, "删除成功", nil)
}

// RefreshCache 刷新配置缓存
func (h *ConfigHandler) RefreshCache(c fiber.Ctx) error {
	if err := h.configService.LoadAll(); err != nil {
		return response.Fail(c, "刷新缓存失败: "+err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionUpdate, model.ModuleConfig, "", "刷新配置缓存")
	return response.SuccessWithMessage(c, "缓存刷新成功", nil)
}

// GetEmailConfig 获取邮件配置
func (h *ConfigHandler) GetEmailConfig(c fiber.Ctx) error {
	configs, err := h.configService.GetByGroup(model.ConfigGroupEmail)
	if err != nil {
		return response.Fail(c, "获取配置失败: "+err.Error())
	}
	return response.Success(c, configs)
}

// UpdateEmailConfigRequest 更新邮件配置请求
type UpdateEmailConfigRequest struct {
	Enabled     bool   `json:"enabled"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromName    string `json:"fromName"`
	FromAddr    string `json:"fromAddr"`
	SSL         bool   `json:"ssl"`
	ResetURL    string `json:"resetUrl"`
	ResetExpire int    `json:"resetExpire"`
}

// UpdateEmailConfig 更新邮件配置
func (h *ConfigHandler) UpdateEmailConfig(c fiber.Ctx) error {
	var req UpdateEmailConfigRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	configs := map[string]string{
		"email_enabled":      boolToString(req.Enabled),
		"email_host":         req.Host,
		"email_port":         intToString(req.Port),
		"email_username":     req.Username,
		"email_password":     req.Password,
		"email_from_name":    req.FromName,
		"email_from_addr":    req.FromAddr,
		"email_ssl":          boolToString(req.SSL),
		"email_reset_url":    req.ResetURL,
		"email_reset_expire": intToString(req.ResetExpire),
	}

	if err := h.configService.BatchUpdate(configs); err != nil {
		h.auditService.LogFail(c, model.ActionUpdate, model.ModuleConfig, "email", err.Error())
		return response.Fail(c, "更新邮件配置失败: "+err.Error())
	}

	h.auditService.LogSuccess(c, model.ActionUpdate, model.ModuleConfig, "email", "更新邮件配置")
	return response.SuccessWithMessage(c, "邮件配置更新成功", nil)
}

// 辅助函数
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func intToString(i int) string {
	return fmt.Sprintf("%d", i)
}
