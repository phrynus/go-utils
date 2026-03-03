package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type UnknownType struct {
	Value interface{}
	Type  reflect.Type
}

func NewUnknownType(data interface{}) UnknownType {
	return UnknownType{
		Value: data,
		Type:  reflect.TypeOf(data),
	}
}

// Bool 布尔类型
// Int 整数类型
// Int8 整数类型 8 位
// Int16 整数类型 16 位
// Int32 整数类型 32 位
// Int64 整数类型 64 位
// Uint 无符号整数类型
// Uint8 无符号整数类型 8 位
// Uint16 无符号整数类型 16 位
// Uint32 无符号整数类型 32 位
// Uint64 无符号整数类型 64 位
// Uintptr 无符号整数类型
// Float32 浮点数类型 32 位
// Float64 浮点数类型 64 位
// Complex64 复数类型 64 位
// Complex128 复数类型 128 位
// Array 数组类型
// Chan 通道类型
// Func 函数类型
// Interface 接口类型
// Map 映射类型
// Pointer 指针类型
// Slice 切片类型
// String 字符串类型
// Struct 结构体类型
// UnsafePointer 不安全指针类型

// Bool 转换为布尔类型
func (u UnknownType) ToBool() bool {
	switch u.Type.Kind() {
	case reflect.Bool:
		return u.Value.(bool)
	case reflect.String:
		b, _ := strconv.ParseBool(u.Value.(string))
		return b
	case reflect.Int:
		return u.Value.(int) != 0
	case reflect.Int8:
		return u.Value.(int8) != 0
	case reflect.Int16:
		return u.Value.(int16) != 0
	case reflect.Int32:
		return u.Value.(int32) != 0
	case reflect.Int64:
		return u.Value.(int64) != 0
	case reflect.Uint:
		return u.Value.(uint) != 0
	case reflect.Uint8:
		return u.Value.(uint8) != 0
	case reflect.Uint16:
		return u.Value.(uint16) != 0
	case reflect.Uint32:
		return u.Value.(uint32) != 0
	case reflect.Uint64:
		return u.Value.(uint64) != 0
	case reflect.Float32:
		return u.Value.(float32) != 0
	case reflect.Float64:
		return u.Value.(float64) != 0
	}
	return false
}

// ToInt 转换为int类型
func (u UnknownType) ToInt() int {
	return int(u.ToInt64())
}

// ToInt8 转换为int8类型
func (u UnknownType) ToInt8() int8 {
	return int8(u.ToInt64())
}

// ToInt16 转换为int16类型
func (u UnknownType) ToInt16() int16 {
	return int16(u.ToInt64())
}

// ToInt32 转换为int32类型
func (u UnknownType) ToInt32() int32 {
	return int32(u.ToInt64())
}

// ToInt64 转换为int64类型
func (u UnknownType) ToInt64() int64 {
	switch u.Type.Kind() {
	case reflect.Int:
		return int64(u.Value.(int))
	case reflect.Int8:
		return int64(u.Value.(int8))
	case reflect.Int16:
		return int64(u.Value.(int16))
	case reflect.Int32:
		return int64(u.Value.(int32))
	case reflect.Int64:
		return u.Value.(int64)
	case reflect.String:
		i, _ := strconv.ParseInt(u.Value.(string), 10, 64)
		return i
	case reflect.Float32:
		return int64(u.Value.(float32))
	case reflect.Float64:
		return int64(u.Value.(float64))
	case reflect.Bool:
		if u.Value.(bool) {
			return 1
		}
		return 0
	case reflect.Uint:
		return int64(u.Value.(uint))
	case reflect.Uint8:
		return int64(u.Value.(uint8))
	case reflect.Uint16:
		return int64(u.Value.(uint16))
	case reflect.Uint32:
		return int64(u.Value.(uint32))
	case reflect.Uint64:
		return int64(u.Value.(uint64))
	}
	return 0
}

// ToUint 转换为uint类型
func (u UnknownType) ToUint() uint {
	return uint(u.ToUint64())
}

// ToUint8 转换为uint8类型
func (u UnknownType) ToUint8() uint8 {
	return uint8(u.ToUint64())
}

// ToUint16 转换为uint16类型
func (u UnknownType) ToUint16() uint16 {
	return uint16(u.ToUint64())
}

// ToUint32 转换为uint32类型
func (u UnknownType) ToUint32() uint32 {
	return uint32(u.ToUint64())
}

// ToUint64 转换为uint64类型
func (u UnknownType) ToUint64() uint64 {
	switch u.Type.Kind() {
	case reflect.Uint:
		return uint64(u.Value.(uint))
	case reflect.Uint8:
		return uint64(u.Value.(uint8))
	case reflect.Uint16:
		return uint64(u.Value.(uint16))
	case reflect.Uint32:
		return uint64(u.Value.(uint32))
	case reflect.Uint64:
		return u.Value.(uint64)
	case reflect.String:
		i, _ := strconv.ParseUint(u.Value.(string), 10, 64)
		return i
	case reflect.Int:
		return uint64(u.Value.(int))
	case reflect.Int8:
		return uint64(u.Value.(int8))
	case reflect.Int16:
		return uint64(u.Value.(int16))
	case reflect.Int32:
		return uint64(u.Value.(int32))
	case reflect.Int64:
		return uint64(u.Value.(int64))
	case reflect.Float32:
		return uint64(u.Value.(float32))
	case reflect.Float64:
		return uint64(u.Value.(float64))
	case reflect.Bool:
		if u.Value.(bool) {
			return 1
		}
		return 0
	}
	return 0
}

// ToUintptr 转换为uintptr类型
func (u UnknownType) ToUintptr() uintptr {
	return uintptr(u.ToUint64())
}

// ToFloat32 转换为float32类型
func (u UnknownType) ToFloat32() float32 {
	return float32(u.ToFloat64())
}

// ToFloat64 转换为float64类型
func (u UnknownType) ToFloat64() float64 {
	switch u.Type.Kind() {
	case reflect.Float32:
		return float64(u.Value.(float32))
	case reflect.Float64:
		return u.Value.(float64)
	case reflect.String:
		f, _ := strconv.ParseFloat(u.Value.(string), 64)
		return f
	case reflect.Int:
		return float64(u.Value.(int))
	case reflect.Int8:
		return float64(u.Value.(int8))
	case reflect.Int16:
		return float64(u.Value.(int16))
	case reflect.Int32:
		return float64(u.Value.(int32))
	case reflect.Int64:
		return float64(u.Value.(int64))
	case reflect.Uint:
		return float64(u.Value.(uint))
	case reflect.Uint8:
		return float64(u.Value.(uint8))
	case reflect.Uint16:
		return float64(u.Value.(uint16))
	case reflect.Uint32:
		return float64(u.Value.(uint32))
	case reflect.Uint64:
		return float64(u.Value.(uint64))
	case reflect.Bool:
		if u.Value.(bool) {
			return 1
		}
		return 0
	}
	return 0
}

// ToComplex64 转换为complex64类型
func (u UnknownType) ToComplex64() complex64 {
	return complex64(u.ToComplex128())
}

// ToComplex128 转换为complex128类型
func (u UnknownType) ToComplex128() complex128 {
	switch u.Type.Kind() {
	case reflect.Complex64:
		return complex128(u.Value.(complex64))
	case reflect.Complex128:
		return u.Value.(complex128)
	case reflect.String:
		// 尝试解析字符串为复数
		var c complex128
		fmt.Sscanf(u.Value.(string), "%g", &c)
		return c
	case reflect.Int:
		return complex(float64(u.Value.(int)), 0)
	case reflect.Int8:
		return complex(float64(u.Value.(int8)), 0)
	case reflect.Int16:
		return complex(float64(u.Value.(int16)), 0)
	case reflect.Int32:
		return complex(float64(u.Value.(int32)), 0)
	case reflect.Int64:
		return complex(float64(u.Value.(int64)), 0)
	case reflect.Uint:
		return complex(float64(u.Value.(uint)), 0)
	case reflect.Uint8:
		return complex(float64(u.Value.(uint8)), 0)
	case reflect.Uint16:
		return complex(float64(u.Value.(uint16)), 0)
	case reflect.Uint32:
		return complex(float64(u.Value.(uint32)), 0)
	case reflect.Uint64:
		return complex(float64(u.Value.(uint64)), 0)
	case reflect.Float32:
		return complex(float64(u.Value.(float32)), 0)
	case reflect.Float64:
		return complex(u.Value.(float64), 0)
	}
	return 0
}

// ToArray 转换为数组
func (u UnknownType) ToArray() interface{} {
	switch u.Type.Kind() {
	case reflect.Array, reflect.Slice:
		return u.Value
	case reflect.String:
		return []rune(u.Value.(string))
	}
	return nil
}

// ToChan 转换为通道
func (u UnknownType) ToChan() interface{} {
	if u.Type.Kind() == reflect.Chan {
		return u.Value
	}
	return nil
}

// ToFunc 转换为函数
func (u UnknownType) ToFunc() interface{} {
	if u.Type.Kind() == reflect.Func {
		return u.Value
	}
	return nil
}

// ToInterface 转换为interface{}
func (u UnknownType) ToInterface() interface{} {
	return u.Value
}

// ToMap 转换为map
func (u UnknownType) ToMap() map[string]interface{} {
	if u.Type.Kind() == reflect.Map {
		return u.Value.(map[string]interface{})
	}
	return nil
}

// ToPointer 转换为指针
func (u UnknownType) ToPointer() interface{} {
	if u.Type.Kind() == reflect.Ptr {
		return u.Value
	}
	return nil
}

// ToSlice 转换为切片
func (u UnknownType) ToSlice() interface{} {
	switch u.Type.Kind() {
	case reflect.Slice:
		return u.Value
	case reflect.Array:
		return u.Value
	}
	return nil
}

// ToString 转换为字符串
func (u UnknownType) ToString() string {
	switch u.Type.Kind() {
	case reflect.String:
		return u.Value.(string)
	case reflect.Int:
		return strconv.FormatInt(int64(u.Value.(int)), 10)
	case reflect.Int8:
		return strconv.FormatInt(int64(u.Value.(int8)), 10)
	case reflect.Int16:
		return strconv.FormatInt(int64(u.Value.(int16)), 10)
	case reflect.Int32:
		return strconv.FormatInt(int64(u.Value.(int32)), 10)
	case reflect.Int64:
		return strconv.FormatInt(u.Value.(int64), 10)
	case reflect.Uint:
		return strconv.FormatUint(uint64(u.Value.(uint)), 10)
	case reflect.Uint8:
		return strconv.FormatUint(uint64(u.Value.(uint8)), 10)
	case reflect.Uint16:
		return strconv.FormatUint(uint64(u.Value.(uint16)), 10)
	case reflect.Uint32:
		return strconv.FormatUint(uint64(u.Value.(uint32)), 10)
	case reflect.Uint64:
		return strconv.FormatUint(u.Value.(uint64), 10)
	case reflect.Float32:
		return strconv.FormatFloat(float64(u.Value.(float32)), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(u.Value.(float64), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(u.Value.(bool))
	case reflect.Complex64:
		return fmt.Sprintf("%v", complex128(u.Value.(complex64)))
	case reflect.Complex128:
		return fmt.Sprintf("%v", u.Value.(complex128))
	case reflect.Interface:
		return fmt.Sprintf("%v", u.Value)
	case reflect.Ptr:
		if u.Value != nil {
			return fmt.Sprintf("%v", u.Value)
		}
		return ""
	case reflect.Struct:
		return fmt.Sprintf("%v", u.Value)
	case reflect.Map:
		return fmt.Sprintf("%v", u.Value)
	case reflect.Slice:
		return fmt.Sprintf("%v", u.Value)
	case reflect.Array:
		return fmt.Sprintf("%v", u.Value)
	case reflect.Chan:
		return fmt.Sprintf("%v", u.Value)
	case reflect.Func:
		return fmt.Sprintf("%v", u.Value)
	}
	return fmt.Sprintf("%v", u.Value)
}

// ToStruct 转换为结构体
func (u UnknownType) ToStruct() interface{} {
	if u.Type.Kind() == reflect.Struct {
		return u.Value
	}
	return nil
}

// ToUnsafePointer 转换为unsafe.Pointer
func (u UnknownType) ToUnsafePointer() interface{} {
	if u.Type.Kind() == reflect.UnsafePointer {
		return u.Value
	}
	return nil
}

// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------

// unmarshalWithTypeConversion 根据目标类型进行智能转换
func (u UnknownType) SmartUnmarshal(v interface{}) error {
	data := u.Value

	// 获取目标值的反射
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("目标必须是非空指针")
	}

	// 填充数据
	return setValue(rv.Elem(), data)
}

// fillStruct 递归填充结构体
func fillStruct(target reflect.Value, data interface{}) error {
	if !target.CanSet() {
		return nil
	}

	switch target.Kind() {
	case reflect.Struct:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		targetType := target.Type()
		for i := 0; i < target.NumField(); i++ {
			field := target.Field(i)
			fieldType := targetType.Field(i)

			// 跳过未导出的字段
			if !fieldType.IsExported() {
				continue
			}

			// 获取JSON标签
			jsonTag := fieldType.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				jsonTag = fieldType.Name
			} else {
				// 处理json标签的逗号分隔(如 `json:"name,omitempty"`)
				if idx := bytes.IndexByte([]byte(jsonTag), ','); idx != -1 {
					jsonTag = jsonTag[:idx]
				}
			}

			// 获取对应的数据
			if rawValue, exists := dataMap[jsonTag]; exists {
				if err := setValue(field, rawValue); err != nil {
					return err
				}
			}
		}

	case reflect.Slice:
		dataSlice, ok := data.([]interface{})
		if !ok {
			return nil
		}

		slice := reflect.MakeSlice(target.Type(), len(dataSlice), len(dataSlice))
		for i, item := range dataSlice {
			if err := setValue(slice.Index(i), item); err != nil {
				return err
			}
		}
		target.Set(slice)

	case reflect.Map:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		}

		mapValue := reflect.MakeMap(target.Type())
		for key, value := range dataMap {
			keyValue := reflect.ValueOf(key)
			elemValue := reflect.New(target.Type().Elem()).Elem()
			if err := setValue(elemValue, value); err != nil {
				return err
			}
			mapValue.SetMapIndex(keyValue, elemValue)
		}
		target.Set(mapValue)
	}

	return nil
}

// fillArray 填充数组
func fillArray(target reflect.Value, data interface{}) error {
	if !target.CanSet() {
		return nil
	}

	dataSlice, ok := data.([]interface{})
	if !ok {
		return nil
	}

	// 如果目标数组长度小于数据长度，只填充到数组长度
	length := target.Len()
	if len(dataSlice) < length {
		length = len(dataSlice)
	}

	for i := 0; i < length; i++ {
		if err := setValue(target.Index(i), dataSlice[i]); err != nil {
			return err
		}
	}

	return nil
}

// handleTimeStruct 处理 time.Time 结构
func handleTimeStruct(target reflect.Value, data interface{}) error {
	// 尝试解析各种时间格式
	var timeStr string
	switch v := data.(type) {
	case string:
		timeStr = v
	case json.Number:
		timeStr = v.String()
	default:
		return nil
	}

	// 尝试多种时间格式解析
	timeFormats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, timeStr); err == nil {
			target.Set(reflect.ValueOf(t))
			return nil
		}
	}

	return nil
}

// setValue 根据目标类型设置值
func setValue(target reflect.Value, data interface{}) error {
	if !target.CanSet() {
		return nil
	}

	// 处理空值
	if data == nil {
		return nil
	}

	targetKind := target.Kind()

	// 使用 UnknownType 进行智能类型转换
	unknown := NewUnknownType(data)

	switch targetKind {
	case reflect.String:
		target.SetString(unknown.ToString())

	case reflect.Int:
		target.SetInt(unknown.ToInt64())

	case reflect.Int8:
		target.SetInt(int64(unknown.ToInt8()))

	case reflect.Int16:
		target.SetInt(int64(unknown.ToInt16()))

	case reflect.Int32:
		target.SetInt(int64(unknown.ToInt32()))

	case reflect.Int64:
		target.SetInt(unknown.ToInt64())

	case reflect.Uint:
		target.SetUint(unknown.ToUint64())

	case reflect.Uint8:
		target.SetUint(uint64(unknown.ToUint8()))

	case reflect.Uint16:
		target.SetUint(uint64(unknown.ToUint16()))

	case reflect.Uint32:
		target.SetUint(uint64(unknown.ToUint32()))

	case reflect.Uint64:
		target.SetUint(unknown.ToUint64())

	case reflect.Uintptr:
		target.SetUint(uint64(unknown.ToUintptr()))

	case reflect.Float32:
		target.SetFloat(unknown.ToFloat64())

	case reflect.Float64:
		target.SetFloat(unknown.ToFloat64())

	case reflect.Bool:
		target.SetBool(unknown.ToBool())

	case reflect.Complex64:
		target.SetComplex(complex128(unknown.ToComplex64()))

	case reflect.Complex128:
		target.SetComplex(unknown.ToComplex128())

	case reflect.Ptr:
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return setValue(target.Elem(), data)

	case reflect.Struct:
		// 尝试处理 time.Time
		if target.Type().String() == "time.Time" {
			return handleTimeStruct(target, data)
		}
		return fillStruct(target, data)

	case reflect.Slice:
		// 处理字节切片
		if target.Type().Elem().Kind() == reflect.Uint8 {
			if bytes, ok := data.([]byte); ok {
				target.SetBytes(bytes)
				return nil
			}
		}
		return fillStruct(target, data)

	case reflect.Map:
		return fillStruct(target, data)

	case reflect.Interface:
		if target.NumMethod() == 0 {
			target.Set(reflect.ValueOf(data))
		}
	case reflect.Array:
		return fillArray(target, data)
	case reflect.Chan:
		// 通道不支持直接设置
	case reflect.Func:
		// 函数不支持直接设置
	case reflect.UnsafePointer:
		// 不安全指针不支持直接设置
	}

	return nil
}

// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------
// --------------------------------------------------------------------

// JSONSort JSON排序输出
func (u UnknownType) JSONSort() ([]byte, error) {
	// 使用 Marshal 进行 JSON 序列化，会自动按 key 排序
	return json.Marshal(u.Value)
}

// JSONSortIndent JSON排序并格式化输出
func (u UnknownType) JSONSortIndent() ([]byte, error) {
	// 使用 MarshalIndent 进行格式化 JSON 序列化
	return json.MarshalIndent(u.Value, "", "  ")
}

// JSONToMap 将值转换为 map[string]interface{}（支持多层嵌套）
func (u UnknownType) JSONToMap() (map[string]interface{}, error) {
	if u.Type.Kind() == reflect.Map {
		result := convertToMapInterface(u.Value)
		if m, ok := result.(map[string]interface{}); ok {
			return m, nil
		}
		return nil, fmt.Errorf("无法转换为 map[string]interface{}")
	}
	return nil, fmt.Errorf("值不是 map 类型")
}

// convertToMapInterface 递归转换任意 map 为 map[string]interface{}
func convertToMapInterface(val interface{}) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		// 已是最底层，直接返回
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = convertToMapInterface(value)
		}
		return result

	case map[interface{}]interface{}:
		// 处理 map[interface{}]interface{} 类型（JSON 解析可能产生）
		result := make(map[string]interface{})
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertToMapInterface(value)
		}
		return result

	case []interface{}:
		// 处理切片中的嵌套 map
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertToMapInterface(item)
		}
		return result

	default:
		return val
	}
}

// JSONToSlice 将值转换为 []interface{}（支持多层嵌套）
func (u UnknownType) JSONToSlice() ([]interface{}, error) {
	if u.Type.Kind() == reflect.Slice {
		if s, ok := u.Value.([]interface{}); ok {
			return convertSliceInterface(s), nil
		}
	}
	return nil, fmt.Errorf("值不是切片类型")
}

// convertSliceInterface 递归转换切片中的元素
func convertSliceInterface(val []interface{}) []interface{} {
	result := make([]interface{}, len(val))
	for i, item := range val {
		result[i] = convertToMapInterface(item)
	}
	return result
}
