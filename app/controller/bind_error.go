package controller

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func bindErrorMessage(err error) string {
	if err == nil {
		return "请求参数错误"
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) && len(validationErrors) > 0 {
		return "请求参数错误: " + validationFieldErrorMessage(validationErrors[0])
	}

	return "请求参数错误: " + err.Error()
}

func validationFieldErrorMessage(fieldError validator.FieldError) string {
	fieldName := validationFieldName(fieldError.Field())

	switch fieldError.Tag() {
	case "required":
		return fieldName + "不能为空"
	case "url":
		if fieldError.Field() == "FallbackURL" {
			return fieldName + " 格式无效，请输入完整 URL，例如 https://example.com，或留空"
		}
		return fieldName + " 格式无效，请输入完整 URL，例如 https://example.com"
	case "oneof":
		return fieldName + "取值无效，仅支持 " + strings.ReplaceAll(fieldError.Param(), " ", "、")
	case "min":
		return fieldName + "不能小于 " + fieldError.Param()
	case "max":
		return fieldName + "不能大于 " + fieldError.Param()
	default:
		return fieldName + "参数无效"
	}
}

func validationFieldName(field string) string {
	names := map[string]string{
		"OriginalURL":    "原始 URL",
		"FallbackURL":    "兜底地址 URL",
		"RedirectCode":   "跳转状态码",
		"Page":           "页码",
		"PageSize":       "每页数量",
		"SecurityStatus": "安全状态",
		"RoutingStatus":  "路由状态",
		"URLs":           "URL 列表",
	}
	if name, ok := names[field]; ok {
		return name
	}
	return field
}
