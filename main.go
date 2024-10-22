package main

import (
	es7 "github.com/elastic/go-elasticsearch/v7"
	es7util "github.com/elastic/go-elasticsearch/v7/esutil"
	es8 "github.com/elastic/go-elasticsearch/v8"
	es8util "github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/go-mysql-org/go-mysql/canal"
	gomysql "github.com/go-mysql-org/go-mysql/mysql"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

var (
	RedisClient       redis.UniversalClient
	Es7Client         *es7.Client
	Es7Bi             es7util.BulkIndexer
	Es8Client         *es8.Client
	Es8Bi             es8util.BulkIndexer
	mysqlPosition     gomysql.Position
	includeTableRegex []string
	eventRules        []EventRule
)

func main() {
	mysqlCfg := Config.Mysql
	// 初始化 规则
	InitRules(mysqlCfg)
	cfg := canal.NewDefaultConfig()
	// CREATE USER canal IDENTIFIED BY 'canal';
	// GRANT RELOAD,SELECT, REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO IDENTIFIED BY 'canal' WITH GRANT OPTION;
	// FLUSH PRIVILEGES;
	cfg.Addr = mysqlCfg.Addr
	cfg.ServerID = mysqlCfg.ServerId
	cfg.MaxReconnectAttempts = 5
	cfg.User = mysqlCfg.Username
	cfg.Password = mysqlCfg.Password
	cfg.Logger = SlogAdapter{Adapter: slog.Default()}
	cfg.Dump.ExecutionPath = ""
	if len(includeTableRegex) > 0 {
		cfg.IncludeTableRegex = includeTableRegex
	}
	c, err := canal.NewCanal(cfg)
	if err != nil {
		slog.Error("new canal error", slog.Any("err", err))
	}
	slog.Info("canal table", slog.Any("includeTableRegex", includeTableRegex))
	c.SetEventHandler(&MyEventHandler{Rules: eventRules})
	if err = c.RunFrom(mysqlPosition); err != nil {
		slog.Error("start canal error", slog.Any("err", err))
	}
}

func direct(c1 Consumer, ch chan *EventData) {
	for {
		if d1, ok := <-ch; ok {
			c1.Accept(d1)
		}
	}
}
