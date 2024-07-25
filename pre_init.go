package main

import (
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	gomysql "github.com/go-mysql-org/go-mysql/mysql"
	slogGorm "github.com/orandin/slog-gorm"
	"github.com/samber/lo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log/slog"
	"regexp"
	"slices"
	"strings"
)

type MySqlPosition struct {
	File     string `json:"file"`
	Position uint32 `json:"position"`
}

func InitRules(mysqlCfg MysqlConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlCfg.Username, mysqlCfg.Password, mysqlCfg.Addr, "information_schema")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: slogGorm.New()})
	if err != nil {
		slog.Error("connect mysql ", slog.Any("error", err))
		panic(err)
	}
	slog.Info("connect mysql server success")
	var mp MySqlPosition
	err = db.Raw("SHOW MASTER STATUS;").Scan(&mp).Error
	if err != nil {
		slog.Error("execute mysql `show master status` ", slog.Any("error", err))
		panic(err)
	}
	mysqlPosition = gomysql.Position{Name: mp.File, Pos: mp.Position}
	slog.Info("get mysql position", slog.Any("position", mysqlPosition))
	var tableNames []string
	err = db.Raw(`
SELECT
	CONCAT( TABLE_SCHEMA, '.', TABLE_NAME )
FROM
	INFORMATION_SCHEMA.TABLES 
WHERE
	TABLE_SCHEMA NOT IN ( 'information_schema', 'mysql', 'sys', 'performance_schema' );`,
	).Scan(&tableNames).Error
	if err != nil {
		slog.Error("execute mysql `get table name`", slog.Any("error", err))
		panic(err)
	}
	for key, rule := range Config.Rules {
		if !slices.Contains(includeTableRegex, rule.TableRegex) {
			includeTableRegex = append(includeTableRegex, rule.TableRegex)
		}
		reg, err := regexp.Compile(rule.TableRegex)
		if err != nil {
			slog.Error(fmt.Sprintf("%s regexp:", key), slog.Any("error", err))
			panic(err)
		}
		eventRule := EventRule{Reg: reg, Stream: make(chan *EventData, 1024)}
		eventRules = append(eventRules, eventRule)
		switch rule.SyncTarget {
		case "redis":
			if RedisClient == nil {
				CreateRedisClient()
			}
			c1 := &RedisConsumer{
				KeyName:             rule.RedisRule.KeyName,
				KeyType:             rule.RedisRule.KeyType,
				CustomPKColumn:      rule.CustomPKColumn,
				SerializationFormat: rule.SerializationFormat,
				IncludeColumnNames:  rule.IncludeColumnNames,
				ExcludeColumnNames:  rule.ExcludeColumnNames,
				FieldNameFormat:     rule.FieldNameFormat,
				Logger:              slog.Default(),
			}
			go direct(c1, eventRule.Stream)
			// 清空 之前的数据
			if rule.ClearBeforeData {
				c1.ClearBeforeData()
			}
			// 初始化 数据
			if rule.InitData {
				// 初始化数据
				InitData(db, tableNames, reg, c1)
			}
			break
		case "es7":
			if Es7Client == nil {
				CreateElasticsearch7Client()
			}
			c1 := &Elasticsearch7Consumer{
				IndexName:          rule.ElasticsearchRule.IndexName,
				CustomPKColumn:     rule.CustomPKColumn,
				IncludeColumnNames: rule.IncludeColumnNames,
				ExcludeColumnNames: rule.ExcludeColumnNames,
				FieldNameFormat:    rule.FieldNameFormat,
				Logger:             slog.Default(),
			}
			go direct(c1, eventRule.Stream)
			// 清空 之前的数据
			if rule.ClearBeforeData {
				c1.ClearBeforeData()
			}
			// 初始化 数据
			if rule.InitData {
				// 初始化数据
				InitData(db, tableNames, reg, c1)
			}
			break
		case "es8":
			if Es8Client == nil {
				CreateElasticsearch8Client()
			}
			c1 := &Elasticsearch8Consumer{
				IndexName:          rule.ElasticsearchRule.IndexName,
				CustomPKColumn:     rule.CustomPKColumn,
				IncludeColumnNames: rule.IncludeColumnNames,
				ExcludeColumnNames: rule.ExcludeColumnNames,
				FieldNameFormat:    rule.FieldNameFormat,
				Logger:             slog.Default(),
			}
			go direct(c1, eventRule.Stream)
			// 清空 之前的数据
			if rule.ClearBeforeData {
				c1.ClearBeforeData()
			}
			// 初始化 数据
			if rule.InitData {
				// 初始化数据
				InitData(db, tableNames, reg, c1)
			}
			break
		default:
			c1 := &ConsoleConsumer{Logger: slog.Default()}
			go direct(c1, eventRule.Stream)
			break
		}
	}
}

func InitData(db *gorm.DB, tableNames []string, reg *regexp.Regexp, c1 Consumer) {
	newTableNames := lo.Uniq(
		lo.Filter(tableNames, func(item string, index int) bool {
			return reg.MatchString(item)
		}),
	)
	for _, tableName := range newTableNames {
		s1 := strings.Split(tableName, ".")
		var pkColumns []string
		err := db.Raw(`SELECT
	COLUMN_NAME 
FROM
	INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
WHERE
	TABLE_SCHEMA = ? 
	AND TABLE_NAME = ? 
	AND CONSTRAINT_NAME = 'PRIMARY';`, s1[0], s1[1]).
			Scan(&pkColumns).
			Error
		if err != nil {
			slog.Error("execute mysql `show master status` ", slog.Any("error", err))
			panic(err)
		}
		var count int64
		db.Table(tableName).Count(&count)
		if count < 1 {
			return
		}
		slog.Info("init data", slog.String("tableName", tableName), slog.Int64("count", count))
		pageSize := int64(10000)
		pageTotal := (count + pageSize - 1) / pageSize
		for pageIndex := int64(1); pageIndex <= pageTotal; pageIndex++ {
			var result []map[string]interface{}
			err := db.Table(tableName).
				Scopes(Paginate(pageIndex, pageSize)).
				Find(&result).
				Error
			if err != nil {
				slog.Error("init failed ", slog.String("tableName", tableName), slog.Any("error", err))
			} else {
				data := lo.Map(result, func(after map[string]interface{}, index int) *EventData {
					return &EventData{
						Action:    canal.InsertAction,
						TableName: tableName,
						PKColumns: pkColumns,
						After:     after,
					}
				})
				c1.BatchAccept(data)
			}
		}
	}
}

// Paginate 分页封装
func Paginate(pageIndex int64, pageSize int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if pageIndex == 0 {
			pageIndex = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		offset := (pageIndex - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}
