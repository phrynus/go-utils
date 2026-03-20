package unknown

import (
	"encoding/json"
	"reflect"
	"strconv"
)

// Complex 转换为 complex128 类型
func (u UnknownType) Complex() complex128 {
	if u.Value == nil {
		return 0
	}

	switch v := u.Value.(type) {
	case bool:
		if v {
			return complex(1, 0)
		}
		return 0
	case int:
		return complex(float64(v), 0)
	case int8:
		return complex(float64(v), 0)
	case int16:
		return complex(float64(v), 0)
	case int32:
		return complex(float64(v), 0)
	case int64:
		return complex(float64(v), 0)
	case uint:
		return complex(float64(v), 0)
	case uint8:
		return complex(float64(v), 0)
	case uint16:
		return complex(float64(v), 0)
	case uint32:
		return complex(float64(v), 0)
	case uint64:
		return complex(float64(v), 0)
	case float32:
		return complex(float64(v), 0)
	case float64:
		return complex(v, 0)
	case complex64:
		return complex128(v)
	case complex128:
		return v
	case string:
		if c, err := strconv.ParseComplex(v, 128); err == nil {
			return c
		}
		return 0
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return complex(f, 0)
		}
		return 0
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return complex(float64(rv.Int()), 0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return complex(float64(rv.Uint()), 0)
	case reflect.Float32, reflect.Float64:
		return complex(rv.Float(), 0)
	case reflect.Bool:
		if rv.Bool() {
			return complex(1, 0)
		}
		return 0
	case reflect.String:
		if c, err := strconv.ParseComplex(rv.String(), 128); err == nil {
			return c
		}
		return 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex()
	case reflect.Map, reflect.Array, reflect.Slice:
		return complex(float64(rv.Len()), 0)
	case reflect.Interface, reflect.Ptr, reflect.Struct, reflect.Func, reflect.Chan:
		if !rv.IsNil() {
			return complex(float64(rv.Pointer()), 0)
		}
	}
	return 0
}

// MustComplex 转换为 complex128，转换失败时返回默认值
func (u UnknownType) MustComplex(defaultVal ...complex128) complex128 {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return u.Complex()
}
