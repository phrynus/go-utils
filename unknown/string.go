package unknown

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// String 转换为字符串
// 支持: string, error, fmt.Stringer, 所有可格式化的类型
func (u UnknownType) String() string {
	if u.Value == nil {
		return ""
	}

	switch v := u.Value.(type) {
	case string:
		return v
	case error:
		return v.Error()
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uintptr:
		return strconv.FormatUint(uint64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case complex64:
		c := complex128(v)
		return fmt.Sprintf("(%s+%si)",
			strconv.FormatFloat(real(c), 'f', -1, 64),
			strconv.FormatFloat(imag(c), 'f', -1, 64))
	case complex128:
		c := v
		return fmt.Sprintf("(%s+%si)",
			strconv.FormatFloat(real(c), 'f', -1, 64),
			strconv.FormatFloat(imag(c), 'f', -1, 64))
	case time.Time:
		return v.Format(time.RFC3339)
	case time.Duration:
		return v.String()
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	case reflect.Complex64, reflect.Complex128:
		c := rv.Complex()
		return fmt.Sprintf("(%s+%si)",
			strconv.FormatFloat(real(c), 'f', -1, 64),
			strconv.FormatFloat(imag(c), 'f', -1, 64))
	case reflect.Interface, reflect.Ptr:
		if rv.IsNil() {
			return ""
		}
		return NewUnknown(rv.Elem().Interface()).String()
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			t := rv.Interface().(time.Time)
			return t.Format(time.RFC3339)
		}
		b, _ := json.Marshal(u.Value)
		return string(b)
	case reflect.Map:
		b, _ := json.Marshal(u.Value)
		return string(b)
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return string(rv.Bytes())
		}
		b, _ := json.Marshal(u.Value)
		return string(b)
	case reflect.Array:
		b, _ := json.Marshal(u.Value)
		return string(b)
	case reflect.Chan, reflect.Func:
		if rv.IsNil() {
			return ""
		}
		return fmt.Sprintf("0x%x", rv.Pointer())
	}
	return ""
}

// MustString 转换为字符串，转换失败时返回默认值
func (u UnknownType) MustString(defaultVal ...string) string {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
	return u.String()
}

// Quote 返回带引号的字符串
func (u UnknownType) Quote() string {
	return fmt.Sprintf("%q", u.String())
}
