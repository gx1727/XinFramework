package flag

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// nullStr 将空字符串转换为 nil,用于数据库 NULL 值处理
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nilIfZero 将零值转换为 nil,用于数据库 NULL 值处理
func nilIfZero(v uint) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

// FormatValidationError 格式化验证错误,返回友好的中文错误提示
func FormatValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok {
		errors := make([]string, 0, len(ve))
		for _, e := range ve {
			msg := formatFieldError(e)
			errors = append(errors, msg)
		}
		return strings.Join(errors, "; ")
	}
	// 其他绑定错误(如 JSON 格式错误)
	return "请求参数格式错误: " + err.Error()
}

// formatFieldError 格式化单个字段的验证错误
func formatFieldError(e validator.FieldError) string {
	fieldName := getFieldName(e.Field())

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s 是必填字段", fieldName)
	case "min":
		return fmt.Sprintf("%s 长度不能小于 %s", fieldName, e.Param())
	case "max":
		return fmt.Sprintf("%s 长度不能超过 %s", fieldName, e.Param())
	case "email":
		return fmt.Sprintf("%s 邮箱格式不正确", fieldName)
	case "url":
		return fmt.Sprintf("%s URL 格式不正确", fieldName)
	case "numeric":
		return fmt.Sprintf("%s 必须是数字", fieldName)
	case "gt":
		return fmt.Sprintf("%s 必须大于 %s", fieldName, e.Param())
	case "gte":
		return fmt.Sprintf("%s 必须大于等于 %s", fieldName, e.Param())
	case "lt":
		return fmt.Sprintf("%s 必须小于 %s", fieldName, e.Param())
	case "lte":
		return fmt.Sprintf("%s 必须小于等于 %s", fieldName, e.Param())
	case "len":
		return fmt.Sprintf("%s 长度必须为 %s", fieldName, e.Param())
	case "oneof":
		return fmt.Sprintf("%s 必须是以下值之一: %s", fieldName, e.Param())
	default:
		return fmt.Sprintf("%s 验证失败", fieldName)
	}
}

// getFieldName 获取字段的友好名称
func getFieldName(field string) string {
	// 可以根据需要添加字段映射
	fieldMap := map[string]string{
		"ID":           "ID",
		"Name":         "名称",
		"Code":         "编码",
		"CategoryID":   "分类",
		"SourceURL":    "源URL",
		"ThumbnailURL": "缩略图URL",
		"PreviewURL":   "预览URL",
		"TemplateURL":  "模板URL",
		"FrameID":      "相框ID",
		"SpaceID":      "空间ID",
		"UserID":       "用户ID",
	}

	if name, ok := fieldMap[field]; ok {
		return name
	}
	return field
}
