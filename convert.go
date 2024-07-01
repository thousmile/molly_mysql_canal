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
	var buf bytes.Buffer
	switch format {
	case "msgpack":
		_ = msgpack.NewEncoder(&buf).Encode(data)
		break
	case "yaml":
		_ = yaml.NewEncoder(&buf).Encode(data)
		break
	case "protobuf":
		for key, val := range data {
			switch newVal := val.(type) {
			case int8:
				data[key] = int32(newVal)
				break
			case int16:
				data[key] = int32(newVal)
				break
			case time.Time:
				data[key] = newVal.Format(time.RFC3339)
				break
			}
		}
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
