package response

import (
	"github.com/gofiber/fiber/v3"
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

func Result(c fiber.Ctx, code int, message string, data interface{}) error {
	return c.JSON(Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Success(c fiber.Ctx, data interface{}) error {
	return Result(c, SUCCESS, "success", data)
}

func SuccessWithMessage(c fiber.Ctx, message string, data interface{}) error {
	return Result(c, SUCCESS, message, data)
}

func Fail(c fiber.Ctx, message string) error {
	return Result(c, ERROR, message, nil)
}

func FailWithCode(c fiber.Ctx, code int, message string) error {
	return Result(c, code, message, nil)
}

// Unauthorized 认证失败 HTTP 401
func Unauthorized(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(Response{
		Code:    fiber.StatusUnauthorized,
		Message: message,
		Data:    nil,
	})
}

// Forbidden 权限不足 HTTP 403
func Forbidden(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(Response{
		Code:    fiber.StatusForbidden,
		Message: message,
		Data:    nil,
	})
}

// TooManyRequests 请求过于频繁 HTTP 429
func TooManyRequests(c fiber.Ctx, message string) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(Response{
		Code:    fiber.StatusTooManyRequests,
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

func SuccessWithPage(c fiber.Ctx, items interface{}, total int64, page, pageSize int) error {
	return Success(c, PageResult{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
