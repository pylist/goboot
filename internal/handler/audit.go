package handler

import (
	"goboot/internal/service"
	"goboot/pkg/response"
	"time"

	"github.com/gofiber/fiber/v3"
)

type AuditHandler struct {
	auditService *service.AuditService
}

func NewAuditHandler() *AuditHandler {
	return &AuditHandler{
		auditService: service.NewAuditService(),
	}
}

type AuditLogListRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	UserID    uint   `json:"userId"`
	Action    string `json:"action"`
	Module    string `json:"module"`
	StartTime string `json:"startTime"` // 格式: 2006-01-02 15:04:05
	EndTime   string `json:"endTime"`
}

// GetAuditLogs 获取审计日志列表
func (h *AuditHandler) GetAuditLogs(c fiber.Ctx) error {
	var req AuditLogListRequest
	if err := c.Bind().Body(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 解析时间
	var startTime, endTime *time.Time
	if req.StartTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, time.Local)
		if err == nil {
			startTime = &t
		}
	}
	if req.EndTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndTime, time.Local)
		if err == nil {
			endTime = &t
		}
	}

	serviceReq := &service.AuditLogListRequest{
		Page:      req.Page,
		PageSize:  req.PageSize,
		UserID:    req.UserID,
		Action:    req.Action,
		Module:    req.Module,
		StartTime: startTime,
		EndTime:   endTime,
	}

	logs, total, err := h.auditService.GetLogs(serviceReq)
	if err != nil {
		return response.Fail(c, err.Error())
	}

	return response.SuccessWithPage(c, logs, total, req.Page, req.PageSize)
}
