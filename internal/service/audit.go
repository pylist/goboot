package service

import (
	"goboot/internal/model"
	"goboot/pkg/logger"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

// Log 记录审计日志
func (s *AuditService) Log(c *gin.Context, action, module, target, detail string, status int) {
	var userID uint
	var username string

	// 获取当前用户信息
	if id, exists := c.Get("userID"); exists {
		userID = id.(uint)
	}
	if name, exists := c.Get("username"); exists {
		username = name.(string)
	}

	log := &model.AuditLog{
		UserID:    userID,
		Username:  username,
		Action:    action,
		Module:    module,
		Target:    target,
		Detail:    detail,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Status:    status,
	}

	// 异步写入数据库，不阻塞主流程
	go func() {
		if err := model.CreateAuditLog(log); err != nil {
			logger.Error("Failed to create audit log", slog.Any("error", err))
		}
	}()
}

// LogSuccess 记录成功操作
func (s *AuditService) LogSuccess(c *gin.Context, action, module, target, detail string) {
	s.Log(c, action, module, target, detail, 1)
}

// LogFail 记录失败操作
func (s *AuditService) LogFail(c *gin.Context, action, module, target, detail string) {
	s.Log(c, action, module, target, detail, 0)
}

// GetLogs 获取审计日志列表
func (s *AuditService) GetLogs(req *AuditLogListRequest) ([]model.AuditLog, int64, error) {
	return model.GetAuditLogs(req.Page, req.PageSize, req.UserID, req.Action, req.Module, req.StartTime, req.EndTime)
}

type AuditLogListRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	UserID    uint   `json:"userId"`
	Action    string `json:"action"`
	Module    string `json:"module"`
	StartTime *time.Time `json:"startTime"`
	EndTime   *time.Time `json:"endTime"`
}
