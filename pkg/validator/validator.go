package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`   // 字段名
	Tag     string `json:"tag"`     // 验证规则
	Value   any    `json:"value"`   // 字段值
	Message string `json:"message"` // 错误信息
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ValidationErrors 验证错误集合
type ValidationErrors []*ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	return e[0].Message
}

// First 返回第一个错误
func (e ValidationErrors) First() *ValidationError {
	if len(e) > 0 {
		return e[0]
	}
	return nil
}

// All 返回所有错误信息
func (e ValidationErrors) All() []string {
	msgs := make([]string, len(e))
	for i, err := range e {
		msgs[i] = err.Message
	}
	return msgs
}

// Validator 验证器
type Validator struct {
	tagName    string            // 标签名称，默认 "validate"
	labelTag   string            // 字段标签名，默认 "label"
	messages   map[string]string // 自定义错误消息
	validators map[string]ValidatorFunc
}

// ValidatorFunc 自定义验证函数
type ValidatorFunc func(field reflect.Value, param string) bool

// New 创建验证器
func New() *Validator {
	v := &Validator{
		tagName:    "validate",
		labelTag:   "label",
		messages:   defaultMessages(),
		validators: make(map[string]ValidatorFunc),
	}
	return v
}

// 默认验证器实例
var defaultValidator = New()

// Validate 使用默认验证器验证结构体
func Validate(s any) error {
	return defaultValidator.Validate(s)
}

// RegisterValidator 注册自定义验证器
func RegisterValidator(name string, fn ValidatorFunc) {
	defaultValidator.RegisterValidator(name, fn)
}

// SetMessage 设置错误消息
func SetMessage(tag, message string) {
	defaultValidator.SetMessage(tag, message)
}

// Validate 验证结构体
func (v *Validator) Validate(s any) error {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validator: expected struct, got %s", val.Kind())
	}

	var errors ValidationErrors
	v.validateStruct(val, &errors)

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// RegisterValidator 注册自定义验证器
func (v *Validator) RegisterValidator(name string, fn ValidatorFunc) {
	v.validators[name] = fn
}

// SetMessage 设置错误消息
func (v *Validator) SetMessage(tag, message string) {
	v.messages[tag] = message
}

// validateStruct 验证结构体
func (v *Validator) validateStruct(val reflect.Value, errors *ValidationErrors) {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过未导出字段
		if !field.CanInterface() {
			continue
		}

		// 处理嵌套结构体
		if field.Kind() == reflect.Struct && fieldType.Anonymous {
			v.validateStruct(field, errors)
			continue
		}

		// 获取验证规则
		tagValue := fieldType.Tag.Get(v.tagName)
		if tagValue == "" || tagValue == "-" {
			continue
		}

		// 获取字段标签（中文名）
		label := fieldType.Tag.Get(v.labelTag)
		if label == "" {
			// 尝试从 json 标签获取
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag != "" && jsonTag != "-" {
				label = strings.Split(jsonTag, ",")[0]
			} else {
				label = fieldType.Name
			}
		}

		// 解析验证规则
		rules := strings.Split(tagValue, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			// 解析规则名和参数
			tag, param := parseRule(rule)

			// 执行验证
			if !v.validateField(field, tag, param) {
				msg := v.formatMessage(tag, label, param)
				*errors = append(*errors, &ValidationError{
					Field:   fieldType.Name,
					Tag:     tag,
					Value:   field.Interface(),
					Message: msg,
				})
				break // 一个字段只报告第一个错误
			}
		}
	}
}

// parseRule 解析规则
func parseRule(rule string) (tag, param string) {
	parts := strings.SplitN(rule, "=", 2)
	tag = parts[0]
	if len(parts) > 1 {
		param = parts[1]
	}
	return
}

// validateField 验证字段
func (v *Validator) validateField(field reflect.Value, tag, param string) bool {
	// 先检查自定义验证器
	if fn, ok := v.validators[tag]; ok {
		return fn(field, param)
	}

	// 内置验证器
	switch tag {
	case "required":
		return validateRequired(field)
	case "min":
		return validateMin(field, param)
	case "max":
		return validateMax(field, param)
	case "len":
		return validateLen(field, param)
	case "range":
		return validateRange(field, param)
	case "email":
		return validateEmail(field)
	case "phone":
		return validatePhone(field)
	case "url":
		return validateURL(field)
	case "ip":
		return validateIP(field)
	case "alpha":
		return validateAlpha(field)
	case "alphanum":
		return validateAlphaNum(field)
	case "numeric":
		return validateNumeric(field)
	case "number":
		return validateNumber(field)
	case "lowercase":
		return validateLowercase(field)
	case "uppercase":
		return validateUppercase(field)
	case "contains":
		return validateContains(field, param)
	case "startswith":
		return validateStartsWith(field, param)
	case "endswith":
		return validateEndsWith(field, param)
	case "regex":
		return validateRegex(field, param)
	case "eq":
		return validateEq(field, param)
	case "ne":
		return validateNe(field, param)
	case "gt":
		return validateGt(field, param)
	case "gte":
		return validateGte(field, param)
	case "lt":
		return validateLt(field, param)
	case "lte":
		return validateLte(field, param)
	case "oneof":
		return validateOneOf(field, param)
	case "username":
		return validateUsername(field)
	case "password":
		return validatePassword(field, param)
	case "idcard":
		return validateIDCard(field)
	default:
		return true // 未知规则默认通过
	}
}

// formatMessage 格式化错误消息
func (v *Validator) formatMessage(tag, label, param string) string {
	msg, ok := v.messages[tag]
	if !ok {
		msg = "验证失败"
	}

	// 替换占位符
	msg = strings.ReplaceAll(msg, "{field}", label)
	msg = strings.ReplaceAll(msg, "{param}", param)

	// 处理 range 参数
	if tag == "range" && strings.Contains(param, "-") {
		parts := strings.Split(param, "-")
		if len(parts) == 2 {
			msg = strings.ReplaceAll(msg, "{min}", parts[0])
			msg = strings.ReplaceAll(msg, "{max}", parts[1])
		}
	}

	return msg
}

// defaultMessages 默认错误消息
func defaultMessages() map[string]string {
	return map[string]string{
		"required":   "{field}不能为空",
		"min":        "{field}长度不能小于{param}",
		"max":        "{field}长度不能超过{param}",
		"len":        "{field}长度必须为{param}",
		"range":      "{field}长度必须在{min}-{max}之间",
		"email":      "{field}必须是有效的邮箱地址",
		"phone":      "{field}必须是有效的手机号",
		"url":        "{field}必须是有效的URL",
		"ip":         "{field}必须是有效的IP地址",
		"alpha":      "{field}只能包含字母",
		"alphanum":   "{field}只能包含字母和数字",
		"numeric":    "{field}只能包含数字",
		"number":     "{field}必须是数字",
		"lowercase":  "{field}只能包含小写字母",
		"uppercase":  "{field}只能包含大写字母",
		"contains":   "{field}必须包含{param}",
		"startswith": "{field}必须以{param}开头",
		"endswith":   "{field}必须以{param}结尾",
		"regex":      "{field}格式不正确",
		"eq":         "{field}必须等于{param}",
		"ne":         "{field}不能等于{param}",
		"gt":         "{field}必须大于{param}",
		"gte":        "{field}必须大于或等于{param}",
		"lt":         "{field}必须小于{param}",
		"lte":        "{field}必须小于或等于{param}",
		"oneof":      "{field}必须是以下值之一: {param}",
		"username":   "{field}只能包含字母、数字和下划线",
		"password":   "{field}必须包含字母和数字，长度至少{param}位",
		"idcard":     "{field}必须是有效的身份证号",
	}
}

// ==================== 内置验证函数 ====================

// validateRequired 必填验证
func validateRequired(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		return strings.TrimSpace(field.String()) != ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() > 0
	case reflect.Ptr, reflect.Interface:
		return !field.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return field.Float() != 0
	case reflect.Bool:
		return true // bool 类型始终视为有值
	default:
		return !field.IsZero()
	}
}

// validateMin 最小长度/值验证
func validateMin(field reflect.Value, param string) bool {
	min, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return utf8.RuneCountInString(field.String()) >= min
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() >= min
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() >= int64(min)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() >= uint64(min)
	case reflect.Float32, reflect.Float64:
		return field.Float() >= float64(min)
	default:
		return false
	}
}

// validateMax 最大长度/值验证
func validateMax(field reflect.Value, param string) bool {
	max, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return utf8.RuneCountInString(field.String()) <= max
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() <= max
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() <= int64(max)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() <= uint64(max)
	case reflect.Float32, reflect.Float64:
		return field.Float() <= float64(max)
	default:
		return false
	}
}

// validateLen 精确长度验证
func validateLen(field reflect.Value, param string) bool {
	length, err := strconv.Atoi(param)
	if err != nil {
		return false
	}

	switch field.Kind() {
	case reflect.String:
		return utf8.RuneCountInString(field.String()) == length
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() == length
	default:
		return false
	}
}

// validateRange 范围验证
func validateRange(field reflect.Value, param string) bool {
	parts := strings.Split(param, "-")
	if len(parts) != 2 {
		return false
	}
	return validateMin(field, parts[0]) && validateMax(field, parts[1])
}

// 正则表达式预编译
var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex    = regexp.MustCompile(`^1[3-9]\d{9}$`)
	urlRegex      = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	ipRegex       = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	alphaRegex    = regexp.MustCompile(`^[a-zA-Z]+$`)
	alphaNumRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	numericRegex  = regexp.MustCompile(`^[0-9]+$`)
	numberRegex   = regexp.MustCompile(`^-?[0-9]+\.?[0-9]*$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	idcardRegex   = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)
)

// validateEmail 邮箱验证
func validateEmail(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true // 空值由 required 验证
	}
	return emailRegex.MatchString(s)
}

// validatePhone 手机号验证
func validatePhone(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return phoneRegex.MatchString(s)
}

// validateURL URL验证
func validateURL(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return urlRegex.MatchString(s)
}

// validateIP IP地址验证
func validateIP(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	if !ipRegex.MatchString(s) {
		return false
	}
	// 验证每个数字段在 0-255 范围内
	parts := strings.Split(s, ".")
	for _, part := range parts {
		num, _ := strconv.Atoi(part)
		if num > 255 {
			return false
		}
	}
	return true
}

// validateAlpha 纯字母验证
func validateAlpha(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return alphaRegex.MatchString(s)
}

// validateAlphaNum 字母数字验证
func validateAlphaNum(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return alphaNumRegex.MatchString(s)
}

// validateNumeric 纯数字字符串验证
func validateNumeric(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return numericRegex.MatchString(s)
}

// validateNumber 数字验证（包含负数和小数）
func validateNumber(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return numberRegex.MatchString(s)
}

// validateLowercase 小写字母验证
func validateLowercase(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return s == strings.ToLower(s)
}

// validateUppercase 大写字母验证
func validateUppercase(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return s == strings.ToUpper(s)
}

// validateContains 包含验证
func validateContains(field reflect.Value, param string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	return strings.Contains(field.String(), param)
}

// validateStartsWith 前缀验证
func validateStartsWith(field reflect.Value, param string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return strings.HasPrefix(s, param)
}

// validateEndsWith 后缀验证
func validateEndsWith(field reflect.Value, param string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return strings.HasSuffix(s, param)
}

// validateRegex 正则验证
func validateRegex(field reflect.Value, param string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	re, err := regexp.Compile(param)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

// validateEq 相等验证
func validateEq(field reflect.Value, param string) bool {
	switch field.Kind() {
	case reflect.String:
		return field.String() == param
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, _ := strconv.ParseInt(param, 10, 64)
		return field.Int() == p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p, _ := strconv.ParseUint(param, 10, 64)
		return field.Uint() == p
	case reflect.Float32, reflect.Float64:
		p, _ := strconv.ParseFloat(param, 64)
		return field.Float() == p
	default:
		return false
	}
}

// validateNe 不等验证
func validateNe(field reflect.Value, param string) bool {
	return !validateEq(field, param)
}

// validateGt 大于验证
func validateGt(field reflect.Value, param string) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, _ := strconv.ParseInt(param, 10, 64)
		return field.Int() > p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p, _ := strconv.ParseUint(param, 10, 64)
		return field.Uint() > p
	case reflect.Float32, reflect.Float64:
		p, _ := strconv.ParseFloat(param, 64)
		return field.Float() > p
	default:
		return false
	}
}

// validateGte 大于等于验证
func validateGte(field reflect.Value, param string) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, _ := strconv.ParseInt(param, 10, 64)
		return field.Int() >= p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p, _ := strconv.ParseUint(param, 10, 64)
		return field.Uint() >= p
	case reflect.Float32, reflect.Float64:
		p, _ := strconv.ParseFloat(param, 64)
		return field.Float() >= p
	default:
		return false
	}
}

// validateLt 小于验证
func validateLt(field reflect.Value, param string) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, _ := strconv.ParseInt(param, 10, 64)
		return field.Int() < p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p, _ := strconv.ParseUint(param, 10, 64)
		return field.Uint() < p
	case reflect.Float32, reflect.Float64:
		p, _ := strconv.ParseFloat(param, 64)
		return field.Float() < p
	default:
		return false
	}
}

// validateLte 小于等于验证
func validateLte(field reflect.Value, param string) bool {
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		p, _ := strconv.ParseInt(param, 10, 64)
		return field.Int() <= p
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		p, _ := strconv.ParseUint(param, 10, 64)
		return field.Uint() <= p
	case reflect.Float32, reflect.Float64:
		p, _ := strconv.ParseFloat(param, 64)
		return field.Float() <= p
	default:
		return false
	}
}

// validateOneOf 枚举验证
func validateOneOf(field reflect.Value, param string) bool {
	if field.Kind() != reflect.String {
		// 对于非字符串类型，转换后比较
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val := strconv.FormatInt(field.Int(), 10)
			for _, v := range strings.Split(param, " ") {
				if val == v {
					return true
				}
			}
			return false
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val := strconv.FormatUint(field.Uint(), 10)
			for _, v := range strings.Split(param, " ") {
				if val == v {
					return true
				}
			}
			return false
		default:
			return false
		}
	}

	s := field.String()
	if s == "" {
		return true
	}
	for _, v := range strings.Split(param, " ") {
		if s == v {
			return true
		}
	}
	return false
}

// validateUsername 用户名验证
func validateUsername(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return usernameRegex.MatchString(s)
}

// validatePassword 密码强度验证
func validatePassword(field reflect.Value, param string) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}

	minLen := 6
	if param != "" {
		minLen, _ = strconv.Atoi(param)
	}

	if len(s) < minLen {
		return false
	}

	// 检查是否包含字母和数字
	hasLetter := false
	hasDigit := false
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

// validateIDCard 身份证号验证
func validateIDCard(field reflect.Value) bool {
	if field.Kind() != reflect.String {
		return false
	}
	s := field.String()
	if s == "" {
		return true
	}
	return idcardRegex.MatchString(s)
}
