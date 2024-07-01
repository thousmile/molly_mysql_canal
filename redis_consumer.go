package main

import (
	"context"
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"log/slog"
	"slices"
	"strings"
	"time"
)

type RedisConsumer struct {
	// redis 的 key 名称
	KeyName string `yaml:"keyName" json:"keyName"`

	// redis 的 key 类型。string or hash
	KeyType string `yaml:"keyType" json:"keyType"`

	// 自定义主键
	CustomPKColumn string `yaml:"customPKColumn" json:"customPKColumn"`

	// 序列化格式，仅支持: msgpack , json
	SerializationFormat string `yaml:"serializationFormat" json:"serializationFormat"`

	// 包含的 表格 行 名称。为空，全部行
	IncludeColumnNames []string `yaml:"includeColumnNames" json:"includeColumnNames"`

	// 排除的 表格 行 名称。为空，全部行
	ExcludeColumnNames []string `yaml:"excludeColumnNames" json:"excludeColumnNames"`

	// 字段名称格式，小驼峰: lowerCamelCase ，大驼峰：upperCamelCase 其他.不处理
	FieldNameFormat string `yaml:"fieldNameFormat" json:"fieldNameFormat"`

	*slog.Logger
}

func (c *RedisConsumer) Accept(data *EventData) {
	c.BatchAccept([]*EventData{data})
}

func (c *RedisConsumer) BatchAccept(list []*EventData) {
	ids := lo.Map(
		lo.Filter(list, func(item *EventData, index int) bool {
			return item.Action == canal.UpdateAction || item.Action == canal.DeleteAction
		}),
		func(item *EventData, index int) string {
			return ConvertAnyToString(item.Before[c.getPKColumn(item)])
		},
	)
	if len(ids) > 0 {
		c.remove(ids)
	}
	newList := lo.Filter(list, func(item *EventData, index int) bool {
		return item.Action == canal.UpdateAction || item.Action == canal.InsertAction
	})
	if len(newList) > 0 {
		c.insert(newList)
	}
}

func (c *RedisConsumer) remove(ids []string) {
	ctx := context.Background()
	switch c.KeyType {
	case "hash":
		_, err := RedisClient.HDel(ctx, c.KeyName, ids...).Result()
		if err != nil {
			slog.Error("redis hash delete", slog.Any("err", err))
		}
		break
	default:
		newIds := lo.Map(ids, func(item string, index int) string {
			return fmt.Sprintf("%s:%s", c.KeyName, item)
		})
		_, err := RedisClient.Del(ctx, newIds...).Result()
		if err != nil {
			slog.Error("redis string delete", slog.Any("err", err))
		}
		break
	}
}

func (c *RedisConsumer) insert(list []*EventData) {
	newMap := lo.Associate(list, func(item *EventData) (string, string) {
		id := ConvertAnyToString(item.After[c.getPKColumn(item)])
		// 如果 IncludeColumnNames 只有一个 属性
		if len(c.IncludeColumnNames) == 1 {
			return id, ConvertAnyToString(item.After[c.IncludeColumnNames[0]])
		}
		newMap := make(map[string]interface{}, len(item.After))
		b1 := len(c.IncludeColumnNames) > 0
		b2 := len(c.ExcludeColumnNames) > 0
		for column, value := range item.After {
			if b1 && slices.Contains(c.IncludeColumnNames, column) {
				newMap[ConvertColumn(c.FieldNameFormat, column)] = value
			} else {
				if b2 && slices.Contains(c.ExcludeColumnNames, column) {
					continue
				}
				newMap[ConvertColumn(c.FieldNameFormat, column)] = value
			}
		}
		buf := ConvertSerializationFormat(c.SerializationFormat, newMap)
		if c.KeyType == "hash" {
			return id, buf.String()
		} else {
			return fmt.Sprintf("%s:%s", c.KeyName, id), buf.String()
		}
	})
	ctx := context.Background()
	switch c.KeyType {
	case "hash":
		_, err := RedisClient.HMSet(ctx, c.KeyName, newMap).Result()
		if err != nil {
			slog.Error("redis hash Set", slog.Any("err", err))
		}
		break
	default:
		_, err := RedisClient.MSet(context.Background(), newMap).Result()
		if err != nil {
			slog.Error("redis string Set", slog.Any("err", err))
		}
		break
	}
}

// 获取主键ID
func (c *RedisConsumer) getPKColumn(item *EventData) string {
	if len(c.CustomPKColumn) > 0 {
		return c.CustomPKColumn
	}
	if len(item.PKColumns) > 0 {
		return item.PKColumns[0]
	} else {
		c.Error("database table has no primary key", slog.String("table", item.TableName))
		return ""
	}
}

// ClearBeforeData 清除之前的数据
func (c *RedisConsumer) ClearBeforeData() {
	var rdsId string
	switch c.KeyType {
	case "hash":
		rdsId = c.KeyName
		break
	default:
		rdsId = fmt.Sprintf("%s:*", c.KeyName)
		break
	}
	_, err := RedisClient.Del(context.Background(), rdsId).Result()
	if err != nil {
		slog.Error("redis clear before data error", slog.Any("err", err))
	} else {
		slog.Info("redis clear before data success", slog.Any("redisKey", rdsId))
	}
}

func CreateRedisClient() {
	rdsConf := Config.Redis
	options := redis.UniversalOptions{
		Addrs:        rdsConf.Addrs,
		DB:           rdsConf.DB,
		WriteTimeout: time.Second * 2,
		ReadTimeout:  time.Second * 2,
		DialTimeout:  time.Second * 3,
		PoolSize:     4,
	}
	if len(strings.TrimSpace(rdsConf.Password)) > 0 {
		options.Password = rdsConf.Password
	}
	if len(strings.TrimSpace(rdsConf.Username)) > 0 {
		options.Username = rdsConf.Username
	}
	if len(strings.TrimSpace(rdsConf.MasterName)) > 0 {
		options.MasterName = rdsConf.MasterName
	}
	if len(strings.TrimSpace(rdsConf.SentinelUsername)) > 0 {
		options.SentinelUsername = rdsConf.SentinelUsername
	}
	if len(strings.TrimSpace(rdsConf.SentinelUsername)) > 0 {
		options.SentinelUsername = rdsConf.SentinelUsername
	}
	RedisClient = redis.NewUniversalClient(&options)
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		slog.Error("redis ping", slog.Any("err", err))
		panic(err)
	}
}
