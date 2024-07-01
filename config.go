package main

import (
	"errors"
	"github.com/spf13/viper"
	"log/slog"
)

// Config 全局配置配置文件
var Config *AppConfig

func init() {
	viper.SetDefault("appName", "molly-mysql-canal")
	viper.SetDefault(
		"mysql",
		MysqlConfig{
			Addr:     "127.0.0.1:3306",
			Username: "canal",
			Password: "canal",
			ServerId: 88,
		},
	)
	viper.SetDefault(
		"rules",
		map[string]SyncRule{
			"default": {
				TableRegex:          ".*\\..*",
				SyncTarget:          "",
				SerializationFormat: "json",
				IncludeColumnNames:  []string{},
				FieldNameFormat:     "upperCamelCase",
			},
		},
	)
	viper.SetConfigName("config")                // name of config file (without extension)
	viper.SetConfigType("yaml")                  // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")                     // optionally look for config in the working directory
	if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			slog.Error("viper ReadInConfig ", slog.Any("error", err))
			panic(err)
		}
	}
	if err := viper.Unmarshal(&Config); err != nil {
		slog.Error("viper Unmarshal ", slog.Any("error", err))
		panic(err)
	}
}

type AppConfig struct {
	// 服务名称 ，默认: molly-mysql-canal
	AppName string `yaml:"appName" json:"appName"`

	// mysql 的配置
	Mysql MysqlConfig `yaml:"mysql" json:"mysql"`

	// redis 的配置
	Redis RedisConfig `yaml:"redis" json:"redis"`

	// 同步的规则
	Rules map[string]SyncRule `yaml:"rules" json:"rules"`
}

type MysqlConfig struct {
	Addr string `yaml:"addr" json:"addr" `

	Username string `yaml:"username" json:"username" `

	Password string `yaml:"password" json:"password" `

	ServerId uint32 `yaml:"serverId" json:"serverId" `
}

type SyncRule struct {
	// 表用作ID 的名称
	TableRegex string `yaml:"tableRegex" json:"tableRegex"`

	// 同步的目的地 redis,
	SyncTarget string `yaml:"syncTarget" json:"syncTarget"`

	// 初始化数据
	InitData bool `yaml:"initData" json:"initData"`

	// 清空之前的数据
	ClearBeforeData bool `yaml:"clearBeforeData" json:"clearBeforeData"`

	// 自定义主键
	CustomPKColumn string `yaml:"customPKColumn" json:"customPKColumn"`

	// 序列化格式，支持: msgpack、json、yaml、protobuf
	SerializationFormat string `yaml:"serializationFormat" json:"serializationFormat"`

	// 包含的 表格 行 名称。为空，全部行
	IncludeColumnNames []string `yaml:"includeColumnNames" json:"includeColumnNames"`

	// 排除的 表格 行 名称。为空，全部行
	ExcludeColumnNames []string `yaml:"excludeColumnNames" json:"excludeColumnNames"`

	// 字段名称格式，小驼峰: lowerCamelCase ，大驼峰：upperCamelCase 其他.不处理
	FieldNameFormat string `yaml:"fieldNameFormat" json:"fieldNameFormat"`

	// redis 的配置
	RedisRule SyncRedisRule `yaml:"redisRule" json:"redisRule"`
}

type SyncRedisRule struct {
	// redis 的 key 名称
	KeyName string `yaml:"keyName" json:"keyName"`

	// redis 的 key 类型。string or hash
	KeyType string `yaml:"keyType" json:"keyType"`
}

type RedisConfig struct {
	// redis 地址。默认: 127.0.0.1:6379
	Addrs []string `yaml:"addrs" json:"addrs" `

	// 库索引，默认: 0
	DB int `yaml:"db" json:"db" toml:"db"`

	// 用户名，默认: 空
	Username string `yaml:"username" json:"username" `

	// 密码，默认: 空
	Password string `yaml:"password" json:"password" `

	// Sentinel 模式。
	MasterName string `yaml:"masterName" json:"masterName" `

	// Sentinel username。
	SentinelUsername string `yaml:"sentinelUsername" json:"sentinelUsername" `

	// Sentinel password
	SentinelPassword string `yaml:"sentinelPassword" json:"sentinelPassword" `
}
