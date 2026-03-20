package unknown

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// Uint 转换为 uint64 类型
func (u UnknownType) Uint() uint64 {
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
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int8:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int16:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int32:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int64:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case uintptr:
		return uint64(v)
	case float32:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case float64:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case string:
		if i, err := strconv.ParseUint(v, 10, 64); err == nil {
			return i
		}
		return 0
	case json.Number:
		if i, err := v.Int64(); err == nil && i >= 0 {
			return uint64(i)
		}
		if f, err := v.Float64(); err == nil && f >= 0 {
			return uint64(f)
		}
		return 0
	case time.Time:
		return uint64(v.Unix())
	case time.Duration:
		return uint64(v.Milliseconds())
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if rv.Int() < 0 {
			return 0
		}
		return uint64(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint()
	case reflect.Float32, reflect.Float64:
		if rv.Float() < 0 {
			return 0
		}
		return uint64(rv.Float())
	case reflect.Bool:
		if rv.Bool() {
			return 1
		}
		return 0
	case reflect.String:
		if i, err := strconv.ParseUint(rv.String(), 10, 64); err == nil {
			return i
		}
		return 0
	case reflect.Map, reflect.Array, reflect.Slice:
		return uint64(rv.Len())
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Ptr, reflect.Struct:
		if !rv.IsNil() {
			return uint64(rv.Pointer())
		}
	}
	return 0
}

// MustUint 转换为 uint64，转换失败时返回默认值
func (u UnknownType) MustUint(defaultVal ...uint64) uint64 {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return u.Uint()
}

// Uint8 转换为 uint8
func (u UnknownType) Uint8() uint8 {
	v := u.Uint()
	if v > 255 {
		return 255
	}
	return uint8(v)
}

// Uint16 转换为 uint16
func (u UnknownType) Uint16() uint16 {
	v := u.Uint()
	if v > 65535 {
		return 65535
	}
	return uint16(v)
}

// Uint32 转换为 uint32
func (u UnknownType) Uint32() uint32 {
	v := u.Uint()
	if v > 4294967295 {
		return 4294967295
	}
	return uint32(v)
}
