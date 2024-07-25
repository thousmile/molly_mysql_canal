package main

import (
	"bytes"
	"context"
	"encoding/json"
	es7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	es7util "github.com/elastic/go-elasticsearch/v7/esutil"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/samber/lo"
	"io"
	"log/slog"
	"runtime"
	"slices"
	"strings"
	"time"
)

type Elasticsearch7Consumer struct {
	// es 的 index 名称
	IndexName string `yaml:"indexName" json:"indexName"`

	// 自定义主键
	CustomPKColumn string `yaml:"customPKColumn" json:"customPKColumn"`

	// 包含的 表格 行 名称。为空，全部行
	IncludeColumnNames []string `yaml:"includeColumnNames" json:"includeColumnNames"`

	// 排除的 表格 行 名称。为空，全部行
	ExcludeColumnNames []string `yaml:"excludeColumnNames" json:"excludeColumnNames"`

	// 字段名称格式，小驼峰: lowerCamelCase ，大驼峰：upperCamelCase 其他.不处理
	FieldNameFormat string `yaml:"fieldNameFormat" json:"fieldNameFormat"`

	*slog.Logger
}

func (c *Elasticsearch7Consumer) Accept(data *EventData) {
	c.BatchAccept([]*EventData{data})
}

func (c *Elasticsearch7Consumer) BatchAccept(list []*EventData) {
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

func (c *Elasticsearch7Consumer) remove(ids []string) {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": ids,
			},
		},
	}
	jsonBody, _ := json.Marshal(body)
	req := es7api.DeleteByQueryRequest{
		Index: []string{c.IndexName},
		Body:  bytes.NewReader(jsonBody),
	}
	_, err := req.Do(context.Background(), Es7Client)
	if err != nil {
		slog.Error("elasticsearch 7 remove id", slog.Any("err", err))
	}
}

func (c *Elasticsearch7Consumer) insert(list []*EventData) {
	for _, item := range list {
		id := ConvertAnyToString(item.After[c.getPKColumn(item)])
		newMap := make(map[string]interface{}, len(item.After))
		b1 := len(c.IncludeColumnNames) > 0
		b2 := len(c.ExcludeColumnNames) > 0
		for column, value := range item.After {
			// 如果 IncludeColumnNames 不包含 字段。或者 ExcludeColumnNames 包含 字段。
			if b1 && !slices.Contains(c.IncludeColumnNames, column) ||
				b2 && slices.Contains(c.ExcludeColumnNames, column) {
				continue
			} else {
				newMap[ConvertColumn(c.FieldNameFormat, column)] = value
			}
		}
		buf := ConvertSerializationFormat("json", newMap)
		doc := es7util.BulkIndexerItem{Action: "index", Index: c.IndexName, DocumentID: id, Body: bytes.NewReader(buf.Bytes())}
		err := Es7Bi.Add(context.Background(), doc)
		if err != nil {
			slog.Info("bulk upsert Add doc fail,", slog.Any("err", err))
			panic(err)
		}
	}
}

// 获取主键ID
func (c *Elasticsearch7Consumer) getPKColumn(item *EventData) string {
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
func (c *Elasticsearch7Consumer) ClearBeforeData() {
	req := es7api.DeleteByQueryRequest{
		Index: []string{c.IndexName},
		Body:  strings.NewReader(`{"query": {"match_all": {}}}`),
	}
	resp, err := req.Do(context.Background(), Es7Client)
	if err != nil {
		slog.Error("elasticsearch 7 clear before data error", slog.Any("err", err))
	} else {
		slog.Info("elasticsearch 7 clear before data success", slog.Any("indexName", c.IndexName))
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)
	if resp.IsError() {
		slog.Error("elasticsearch 7 response ", slog.Any("err", err))
	}
}

func CreateElasticsearch7Client() {
	esConf := Config.Elasticsearch
	var es7Cfg es7.Config
	if len(esConf.Addrs) > 0 {
		es7Cfg.Addresses = esConf.Addrs
	}
	if len(esConf.Username) > 0 {
		es7Cfg.Username = esConf.Username
	}
	if len(esConf.Password) > 0 {
		es7Cfg.Password = esConf.Password
	}
	esClient, err := es7.NewClient(es7Cfg)
	if err != nil {
		slog.Error("elasticsearch 7 connect error", slog.Any("err", err))
		panic(err)
	}
	info, err := esClient.Info()
	if err != nil {
		slog.Error("elasticsearch 7 info response error", slog.Any("err", err))
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(info.Body)
	if info.IsError() {
		slog.Error("elasticsearch 7 close response error", slog.Any("err", err))
		panic(err)
	}
	slog.Info("elasticsearch 7", slog.Any("info", info.Status()))
	Es7Client = esClient
	flushInterval, err := time.ParseDuration(esConf.FlushInterval)
	if err != nil {
		flushInterval = 1 * time.Second
	}
	Es7Bi, err = es7util.NewBulkIndexer(es7util.BulkIndexerConfig{
		Client:        esClient,         // The Elasticsearch client
		NumWorkers:    runtime.NumCPU(), // The number of worker goroutines
		FlushInterval: flushInterval,    // The periodic flush interval
	})
	if err != nil {
		slog.Error("elasticsearch 7 indexer error", slog.Any("err", err))
		panic(err)
	}
}
