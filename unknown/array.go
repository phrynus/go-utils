package unknown

import (
	"reflect"
)

// Array 将值转换为 []interface{} 切片
// 支持: slice, array, map(返回keys), string(返回runes), 标量(返回单元素切片), struct(返回字段值)
func (u UnknownType) Array() []interface{} {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case []interface{}:
		return v
	case []string:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	case []int:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	case []int64:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	case []float64:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	case []bool:
		result := make([]interface{}, len(v))
		for i := range v {
			result[i] = v[i]
		}
		return result
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		length := rv.Len()
		result := make([]interface{}, length)
		for i := 0; i < length; i++ {
			result[i] = rv.Index(i).Interface()
		}
		return result
	case reflect.Map:
		keys := rv.MapKeys()
		result := make([]interface{}, 0, len(keys))
		for _, key := range keys {
			result = append(result, key.Interface())
		}
		return result
	case reflect.String:
		runes := []rune(rv.String())
		result := make([]interface{}, len(runes))
		for i, r := range runes {
			result[i] = r
		}
		return result
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return []interface{}{u.Float()}
	case reflect.Bool:
		return []interface{}{u.Bool()}
	case reflect.Complex64, reflect.Complex128:
		return []interface{}{u.Complex()}
	case reflect.Struct:
		if rv.NumField() == 0 {
			return nil
		}
		result := make([]interface{}, rv.NumField())
		for i := 0; i < rv.NumField(); i++ {
			result[i] = rv.Field(i).Interface()
		}
		return result
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		return NewUnknown(rv.Elem().Interface()).Array()
	case reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return NewUnknown(rv.Elem().Interface()).Array()
	case reflect.Chan, reflect.Func:
		if rv.IsNil() {
			return nil
		}
		return []interface{}{rv.Pointer()}
	}
	return nil
}

// Slice 与 Array 相同（Go 中切片更常用）
func (u UnknownType) Slice() []interface{} {
	return u.Array()
}

// MustArray 转换为 []interface{}，转换失败时返回默认值
func (u UnknownType) MustArray(defaultVal ...[]interface{}) []interface{} {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}
	return u.Array()
}

// Len 获取长度（适用于切片、数组、Map、字符串等）
func (u UnknownType) Len() int {
	if u.Value == nil {
		return 0
	}

	switch v := u.Value.(type) {
	case string:
		return len(v)
	case []interface{}:
		return len(v)
	case []byte:
		return len(v)
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan, reflect.String:
		return rv.Len()
	}
	return 0
}
