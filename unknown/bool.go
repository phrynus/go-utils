package unknown

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Bool 转换为布尔类型
// 支持: bool, int/uint/float(非0为true), string("true"/"false"/"1"/"0"/"yes"/"on"), complex, 指针/接口(非nil为true)
func (u UnknownType) Bool() bool {
	if u.Value == nil {
		return false
	}

	switch v := u.Value.(type) {
	case bool:
		return v
	case string:
		s := strings.ToLower(strings.TrimSpace(v))
		return s == "true" || s == "1" || s == "yes" || s == "on"
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return i != 0
		}
		f, err := v.Float64()
		return err == nil && f != 0
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() != 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() != 0
	case reflect.Interface, reflect.Ptr, reflect.Struct, reflect.Map,
		reflect.Slice, reflect.Array, reflect.Chan, reflect.Func:
		return !rv.IsNil()
	}
	return false
}

// MustBool 转换为布尔类型，转换失败时返回默认值
func (u UnknownType) MustBool(defaultVal ...bool) bool {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
	return u.Bool()
}
