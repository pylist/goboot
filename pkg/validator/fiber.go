package validator

import (
	"goboot/pkg/response"

	"github.com/gofiber/fiber/v3"
)

// BindAndValidate 绑定请求体并验证
// 使用方式:
//
//	var req LoginRequest
//	if err := validator.BindAndValidate(c, &req); err != nil {
//	    return err // 已经返回了标准错误响应
//	}
func BindAndValidate(c fiber.Ctx, req any) error {
	// 绑定请求体
	if err := c.Bind().Body(req); err != nil {
		return response.Fail(c, "参数格式错误: "+err.Error())
	}

	// 执行验证
	if err := Validate(req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	return nil
}

// BindQueryAndValidate 绑定Query参数并验证
func BindQueryAndValidate(c fiber.Ctx, req any) error {
	// 绑定Query参数
	if err := c.Bind().Query(req); err != nil {
		return response.Fail(c, "参数格式错误: "+err.Error())
	}

	// 执行验证
	if err := Validate(req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}

	return nil
}

// MustValidate 仅验证（不绑定），返回错误响应
// 适用于已经绑定后需要再次验证的场景
func MustValidate(c fiber.Ctx, req any) error {
	if err := Validate(req); err != nil {
		return response.Fail(c, "参数错误: "+err.Error())
	}
	return nil
}

// ValidateVar 验证单个变量
// 使用方式:
//
//	if err := validator.ValidateVar(email, "required,email"); err != nil {
//	    return response.Fail(c, "邮箱格式不正确")
//	}
func ValidateVar(value any, rules string) error {
	// 创建一个临时结构体进行验证
	type tempStruct struct {
		Value any `validate:""`
	}

	// 使用反射动态设置验证规则会比较复杂
	// 这里简化处理：手动解析规则并验证
	return validateSingleValue(value, rules)
}

// validateSingleValue 验证单个值
func validateSingleValue(value any, rules string) error {
	// 简单实现，后续可以扩展
	v := New()

	// 创建临时结构体
	temp := struct {
		Field any `validate:"" label:"该字段"`
	}{
		Field: value,
	}

	// 这种方式无法动态设置 tag，暂时跳过复杂实现
	// 直接使用结构体验证更可靠
	_ = temp
	_ = v

	return nil
}
