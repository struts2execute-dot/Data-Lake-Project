package schema

import (
	"reflect"
	"strings"
)

type Schema struct {
	Type   string  `json:"type"`
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Field    string `json:"field"`
	Type     string `json:"type"`
	Optional bool   `json:"optional,omitempty"`
}

type Envelope struct {
	Schema  Schema      `json:"schema"`
	Payload interface{} `json:"payload"`
}

func BuildSchemaFromStruct(t reflect.Type, name string) Schema {
	fields := make([]Field, 0)
	var walk func(rt reflect.Type)
	walk = func(rt reflect.Type) {
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			// 匿名嵌入（比如嵌入 Payload）
			if f.Anonymous {
				walk(f.Type)
				continue
			}
			// 非导出字段跳过
			if f.PkgPath != "" {
				continue
			}
			tag := f.Tag.Get("json")
			if tag == "-" {
				continue
			}
			jsonName := strings.Split(tag, ",")[0]
			if jsonName == "" {
				jsonName = f.Name
			}
			// optional 规则：指针类型 或 tag 里含 omitempty
			isOptional := false
			ft := f.Type
			if ft.Kind() == reflect.Ptr {
				isOptional = true
				ft = ft.Elem()
			}
			if strings.Contains(tag, "omitempty") {
				isOptional = true
			}
			var typeStr string
			switch ft.Kind() {
			case reflect.String:
				typeStr = "string"
			case reflect.Int32:
				typeStr = "int32"
			case reflect.Int, reflect.Int64:
				typeStr = "int64"
			case reflect.Float32:
				typeStr = "float"
			case reflect.Float64:
				typeStr = "double"
			case reflect.Bool:
				typeStr = "boolean"
			default:
				// 复杂类型后面需要的话再拓展
				continue
			}
			fields = append(fields, Field{
				Field:    jsonName,
				Type:     typeStr,
				Optional: isOptional,
			})
		}
	}
	walk(t)
	return Schema{
		Type:   "struct",
		Name:   name,
		Fields: fields,
	}
}
