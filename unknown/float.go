package unknown

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// Float 转换为 float64 类型
func (u UnknownType) Float() float64 {
	if u.Value == nil {
		return 0
	}

	switch v := u.Value.(type) {
	case bool:
		if v {
			return 1.0
		}
		return 0.0
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case uintptr:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return 0.0
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
		return 0.0
	case time.Time:
		return float64(v.UnixNano())
	case time.Duration:
		return float64(v.Nanoseconds()) / 1e9
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return rv.Float()
	case reflect.Bool:
		if rv.Bool() {
			return 1.0
		}
		return 0.0
	case reflect.String:
		if f, err := strconv.ParseFloat(rv.String(), 64); err == nil {
			return f
		}
		return 0.0
	case reflect.Map, reflect.Array, reflect.Slice:
		return float64(rv.Len())
	case reflect.Interface, reflect.Ptr, reflect.Struct, reflect.Func, reflect.Chan:
		if !rv.IsNil() {
			return float64(rv.Pointer())
		}
	}
	return 0
}

// MustFloat 转换为 float64，转换失败时返回默认值
func (u UnknownType) MustFloat(defaultVal ...float64) float64 {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return u.Float()
}

// Float32 转换为 float32
func (u UnknownType) Float32() float32 {
	return float32(u.Float())
}
