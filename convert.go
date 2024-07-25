package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
	"log/slog"
	"strconv"
	"strings"
	"time"
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

// ConvertSerializationFormat 转数据格式
func ConvertSerializationFormat(format string, data map[string]interface{}) bytes.Buffer {
	for key, val := range data {
		switch newVal := val.(type) {
		case int8:
			data[key] = int32(newVal)
			break
		case []uint8:
			data[key] = string(newVal)
			break
		case int16:
			data[key] = int32(newVal)
			break
		case time.Time:
			data[key] = ConvertTimeToString(newVal)
			break
		}
	}
	var buf bytes.Buffer
	switch format {
	case "msgpack":
		_ = msgpack.NewEncoder(&buf).Encode(data)
		break
	case "yaml":
		_ = yaml.NewEncoder(&buf).Encode(data)
		break
	case "protobuf":
		pbVal, err := structpb.NewStruct(data)
		if err != nil {
			slog.Error("structpb.NewStruct protobuf value ", slog.Any("error", err))
			return buf
		}
		pbBytes, err := proto.Marshal(pbVal)
		if err != nil {
			slog.Error("proto.Marshal converting protobuf data ", slog.Any("error", err))
			return buf
		}
		buf.Write(pbBytes)
		break
	default:
		_ = json.NewEncoder(&buf).Encode(data)
		break
	}
	return buf
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

// ConvertAnyToString interface 转 字符串
func ConvertAnyToString(value interface{}) string {
	switch v := value.(type) {
	case int8:
		return strconv.Itoa(int(v))
	case []uint8:
		return string(v)
	case int16:
		return strconv.Itoa(int(v))
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

// ConvertTimeToString time 转 字符串
func ConvertTimeToString(value time.Time) string {
	if value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 {
		return value.Format(time.DateOnly)
	} else if value.Year() == 0 && value.Month() == 0 && value.Day() == 0 {
		return value.Format(time.TimeOnly)
	} else {
		return value.Format(time.RFC3339)
	}
}
