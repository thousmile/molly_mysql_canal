package main

import (
	"fmt"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/schema"
	"regexp"
)

type EventData struct {
	// "insert" "update" "delete"
	Action string
	// 表名称
	TableName string
	// 主键名称
	PKColumns []string
	// 执行动作之前
	Before map[string]interface{}
	// 执行动作之后
	After map[string]interface{}
}

type EventRule struct {
	// 正则表达式
	Reg *regexp.Regexp
	// 管道
	Stream chan *EventData
}

type MyEventHandler struct {
	canal.DummyEventHandler
	Rules []EventRule
}

func (h *MyEventHandler) OnRow(e *canal.RowsEvent) error {
	fullTableName := fmt.Sprintf("%s.%s", e.Table.Schema, e.Table.Name)
	for _, rule := range h.Rules {
		ms := rule.Reg.MatchString(fullTableName)
		if ms {
			data := &EventData{
				Action:    e.Action,
				TableName: fullTableName,
				PKColumns: getPKColumns(e.Table),
			}
			switch e.Action {
			case canal.UpdateAction:
				data.Before = anyToObj(e.Rows[0], e.Table)
				data.After = anyToObj(e.Rows[1], e.Table)
				break
			case canal.InsertAction:
				data.After = anyToObj(e.Rows[0], e.Table)
				break
			case canal.DeleteAction:
				data.Before = anyToObj(e.Rows[0], e.Table)
			default:
				break
			}
			rule.Stream <- data
		}
	}
	return nil
}

func anyToObj(row []interface{}, table *schema.Table) map[string]interface{} {
	obj := make(map[string]interface{}, len(row))
	for i, column := range table.Columns {
		if i < len(row) {
			obj[column.Name] = row[i]
		}
	}
	return obj
}

func getPKColumns(table *schema.Table) []string {
	var pkColumns []string
	for _, ind := range table.PKColumns {
		pkColumns = append(pkColumns, table.Columns[ind].Name)
	}
	return pkColumns
}

func (h *MyEventHandler) String() string {
	return "MyEventHandler"
}
