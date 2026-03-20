package unknown

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"
)

// Duration 转换为 time.Duration 类型
// 支持: time.Duration, int/float(毫秒), string(解析Duration字符串或数字), int64(纳秒)
func (u UnknownType) Duration() time.Duration {
	if u.Value == nil {
		return 0
	}

	switch v := u.Value.(type) {
	case time.Duration:
		return v
	case int:
		return time.Duration(v) * time.Millisecond
	case int8:
		return time.Duration(v) * time.Millisecond
	case int16:
		return time.Duration(v) * time.Millisecond
	case int32:
		return time.Duration(v) * time.Millisecond
	case int64:
		return time.Duration(v)
	case uint:
		return time.Duration(v) * time.Millisecond
	case uint8:
		return time.Duration(v) * time.Millisecond
	case uint16:
		return time.Duration(v) * time.Millisecond
	case uint32:
		return time.Duration(v) * time.Millisecond
	case uint64:
		return time.Duration(v)
	case float32:
		return time.Duration(int64(v * 1e6))
	case float64:
		return time.Duration(int64(v * 1e6))
	case string:
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return time.Duration(int64(f * 1e6))
		}
		return 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return time.Duration(i)
		}
		return 0
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return time.Duration(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return time.Duration(rv.Uint())
	case reflect.Float32, reflect.Float64:
		return time.Duration(int64(rv.Float() * 1e6))
	}
	return 0
}

// MustDuration 转换为 time.Duration，转换失败时返回默认值
func (u UnknownType) MustDuration(defaultVal ...time.Duration) time.Duration {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return u.Duration()
}
