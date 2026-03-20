package unknown

import (
	"encoding/json"
	"reflect"
	"time"
)

// IsBool 检查值是否为布尔类型
func (u UnknownType) IsBool() bool {
	_, ok := u.Value.(bool)
	return ok
}

// IsInt 检查值是否为整数类型
func (u UnknownType) IsInt() bool {
	switch u.Value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, json.Number:
		return true
	}
	return false
}

// IsFloat 检查值是否为浮点数类型
func (u UnknownType) IsFloat() bool {
	switch u.Value.(type) {
	case float32, float64:
		return true
	}
	return false
}

// IsNumber 检查值是否为数字类型（整数或浮点）
func (u UnknownType) IsNumber() bool {
	return u.IsInt() || u.IsFloat()
}

// IsString 检查值是否为字符串类型
func (u UnknownType) IsString() bool {
	_, ok := u.Value.(string)
	return ok
}

// IsSlice 检查值是否为切片类型
func (u UnknownType) IsSlice() bool {
	if u.Value == nil {
		return false
	}
	rv := reflect.ValueOf(u.Value)
	return rv.Kind() == reflect.Slice
}

// IsMap 检查值是否为 Map 类型
func (u UnknownType) IsMap() bool {
	if u.Value == nil {
		return false
	}
	rv := reflect.ValueOf(u.Value)
	return rv.Kind() == reflect.Map
}

// IsStruct 检查值是否为结构体类型
func (u UnknownType) IsStruct() bool {
	if u.Value == nil {
		return false
	}
	rv := reflect.ValueOf(u.Value)
	return rv.Kind() == reflect.Struct
}

// IsPtr 检查值是否为指针类型
func (u UnknownType) IsPtr() bool {
	if u.Value == nil {
		return false
	}
	rv := reflect.ValueOf(u.Value)
	return rv.Kind() == reflect.Ptr
}

// IsArray 检查值是否为数组或切片类型
func (u UnknownType) IsArray() bool {
	if u.Value == nil {
		return false
	}
	rv := reflect.ValueOf(u.Value)
	return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
}

// IsTime 检查值是否为时间类型
func (u UnknownType) IsTime() bool {
	_, ok := u.Value.(time.Time)
	return ok
}

// IsDuration 检查值是否为 Duration 类型
func (u UnknownType) IsDuration() bool {
	_, ok := u.Value.(time.Duration)
	return ok
}

// IsNil 检查值是否为 nil
func (u UnknownType) IsNil() bool {
	if u.Value == nil {
		return true
	}
	rv := reflect.ValueOf(u.Value)
	kind := rv.Kind()
	return kind == reflect.Ptr || kind == reflect.Interface ||
		kind == reflect.Map || kind == reflect.Slice ||
		kind == reflect.Chan || kind == reflect.Func
}

// IsZero 检查值是否为零值
func (u UnknownType) IsZero() bool {
	if u.Value == nil {
		return true
	}
	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() == 0
	case reflect.String:
		return rv.String() == ""
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice,
		reflect.Array, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}
