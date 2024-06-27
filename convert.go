package main

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strconv"
	"strings"
)

var (
	causer = cases.Title(language.English)
)

// ConvertColumn 字段转换
func ConvertColumn(fieldNameFormat, column string) string {
	switch fieldNameFormat {
	case "lowerCamelCase":
		return lowerCamelCase(column)
	case "upperCamelCase":
		return upperCamelCase(column)
	default:
		return column
	}
}

// lowerCamelCase 小驼峰
func lowerCamelCase(input string) string {
	parts := strings.Split(input, "_")
	if len(parts) <= 1 {
		parts = strings.Split(input, "-")
	}
	var result []string
	for i, part := range parts {
		if i == 0 {
			result = append(result, strings.ToLower(part))
		} else {
			result = append(result, causer.String(part))
		}
	}
	return strings.Join(result, "")
}

// upperCamelCase 大驼峰
func upperCamelCase(input string) string {
	parts := strings.Split(input, "_")
	if len(parts) <= 1 {
		parts = strings.Split(input, "-")
	}
	var result []string
	for _, part := range parts {
		result = append(result, causer.String(part))
	}
	return strings.Join(result, "")
}

// ConvertInt64 string 转 int64
func ConvertInt64(str string) int64 {
	v, _ := strconv.ParseInt(str, 10, 64)
	return v
}

// ConvertAnyToString interface 转 字符串
func ConvertAnyToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	default:
		return fmt.Sprintf("%v", value)
	}
}
