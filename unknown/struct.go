package unknown

import (
	"reflect"
	"strings"
)

// Struct 转换为结构体类型
// 支持: 本身是 struct, map[string]interface{} 转换为 struct
func (u UnknownType) Struct() interface{} {
	if u.Type != nil && u.Type.Kind() == reflect.Struct {
		return u.Value
	}

	m := u.Map()
	if m != nil {
		return m
	}

	return nil
}

// StructField 获取结构体字段值（通过字段名或 json tag）
func (u UnknownType) StructField(name string) interface{} {
	if u.Value == nil {
		return nil
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Name == name {
			return rv.Field(i).Interface()
		}
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName == name {
				return rv.Field(i).Interface()
			}
		}
	}
	return nil
}

// Fields 获取结构体或 map 的所有字段名和值
func (u UnknownType) Fields() map[string]interface{} {
	if u.Value == nil {
		return nil
	}

	switch m := u.Value.(type) {
	case map[string]interface{}:
		return m
	case map[string]string:
		result := make(map[string]interface{}, len(m))
		for k, v := range m {
			result[k] = v
		}
		return result
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	result := make(map[string]interface{})
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		name := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			name = strings.Split(jsonTag, ",")[0]
		}

		result[name] = rv.Field(i).Interface()
	}
	return result
}
