package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

const (
	SUCCESS = 0
	ERROR   = 1
)

func Result(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Success(c *gin.Context, data interface{}) {
	Result(c, SUCCESS, "success", data)
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	Result(c, SUCCESS, message, data)
}

func Fail(c *gin.Context, message string) {
	Result(c, ERROR, message, nil)
}

func FailWithCode(c *gin.Context, code int, message string) {
	Result(c, code, message, nil)
}

// Unauthorized 认证失败 HTTP 401
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Code:    http.StatusUnauthorized,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 权限不足 HTTP 403
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Code:    http.StatusForbidden,
		Message: message,
		Data:    nil,
	})
}

// TooManyRequests 请求过于频繁 HTTP 429
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{
		Code:    http.StatusTooManyRequests,
		Message: message,
		Data:    nil,
	})
}

type PageResult struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

func SuccessWithPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	Success(c, PageResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
