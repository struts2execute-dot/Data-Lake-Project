package model

import (
	"go-test/model/schema"
	"reflect"
)

func BuildEnvelope(data interface{}) *schema.Envelope {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// 所有业务 struct 都嵌入了 Payload，有 Event 字段
	field := v.FieldByName("Event")
	if !field.IsValid() || field.Kind() != reflect.String {
		panic("BuildEnvelope: data has no string field `Event`")
	}
	eventType := field.String()

	s, ok := SchemaRegistry[eventType]
	if !ok {
		panic("BuildEnvelope: schema not registered for event = " + eventType)
	}

	return &schema.Envelope{
		Schema:  s,
		Payload: data,
	}
}
