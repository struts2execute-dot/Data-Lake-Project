package model

import (
	"go-test/model/payload"
	"go-test/model/schema"
	"reflect"
)

// SchemaRegistry 业务类型 -> 对应的 Schema
var SchemaRegistry = map[string]schema.Schema{}

func RegisterSchema(eventType string, sample interface{}) {
	t := reflect.TypeOf(sample)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	s := schema.BuildSchemaFromStruct(t, eventType+"_record")
	SchemaRegistry[eventType] = s
}

func init() {
	// event 类型直接用 Event 字段的值
	RegisterSchema("recharge", payload.Recharge{})
	RegisterSchema("bet", payload.Bet{})
}
