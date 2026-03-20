package unknown

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// Int 转换为 int64 类型
// 支持: bool(1/0), int/uint/float, string(解析数字), complex(实部), 切片长度, time.Time(Unix时间戳), time.Duration(毫秒)
func (u UnknownType) Int() int64 {
	if u.Value == nil {
		return 0
	}

	switch v := u.Value.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		return int64(v)
	case uintptr:
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int64(f)
		}
		return 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
		if f, err := v.Float64(); err == nil {
			return int64(f)
		}
		return 0
	case time.Time:
		return v.Unix()
	case time.Duration:
		return v.Milliseconds()
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float())
	case reflect.Complex64, reflect.Complex128:
		return int64(real(rv.Complex()))
	case reflect.Bool:
		if rv.Bool() {
			return 1
		}
		return 0
	case reflect.String:
		if i, err := strconv.ParseInt(rv.String(), 10, 64); err == nil {
			return i
		}
		return 0
	case reflect.Array, reflect.Map, reflect.Chan, reflect.Slice:
		return int64(rv.Len())
	case reflect.Interface, reflect.Ptr, reflect.Struct, reflect.Func:
		if !rv.IsNil() {
			return int64(rv.Pointer())
		}
	}
	return 0
}

// MustInt 转换为 int64，转换失败时返回默认值
func (u UnknownType) MustInt(defaultVal ...int64) int64 {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return u.Int()
}

// Int8 转换为 int8
func (u UnknownType) Int8() int8 {
	v := u.Int()
	if v < -128 {
		return -128
	}
	if v > 127 {
		return 127
	}
	return int8(v)
}

// Int16 转换为 int16
func (u UnknownType) Int16() int16 {
	v := u.Int()
	if v < -32768 {
		return -32768
	}
	if v > 32767 {
		return 32767
	}
	return int16(v)
}

// Int32 转换为 int32
func (u UnknownType) Int32() int32 {
	v := u.Int()
	if v < -2147483648 {
		return -2147483648
	}
	if v > 2147483647 {
		return 2147483647
	}
	return int32(v)
}
