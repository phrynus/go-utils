package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// SmartUnmarshal 智能JSON解析,根据目标类型进行自动转换
// 支持:
// - 字符串数字 → 数字类型
// - 数字 → 字符串类型
// - 空字符串 → 0/""
// - 任意类型之间的智能转换
func SmartUnmarshal(data []byte, v interface{}) error {
	// // 先尝试标准JSON解析
	// if err := json.Unmarshal(data, v); err == nil {
	// 	return nil
	// }

	// 如果失败,使用智能转换
	return unmarshalWithTypeConversion(data, v)
}

// unmarshalWithTypeConversion 根据目标类型进行智能转换
func unmarshalWithTypeConversion(data []byte, v interface{}) error {
	// 解析JSON到interface{}
	var rawData interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&rawData); err != nil {
		return err
	}

	// 获取目标值的反射
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("目标必须是非空指针")
	}

	// 填充数据
	return setValue(rv.Elem(), rawData)
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

	// 转换json.Number为具体类型
	if num, ok := data.(json.Number); ok {
		data = convertJSONNumber(num, targetKind)
	}

	switch targetKind {
	case reflect.String:
		target.SetString(ToString(data))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(ToInt64(data))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		target.SetUint(uint64(ToInt64(data)))

	case reflect.Float32, reflect.Float64:
		target.SetFloat(ToFloat64(data))

	case reflect.Bool:
		target.SetBool(ToBool(data))

	case reflect.Ptr:
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return setValue(target.Elem(), data)

	case reflect.Struct:
		return fillStruct(target, data)

	case reflect.Slice:
		return fillStruct(target, data)

	case reflect.Map:
		return fillStruct(target, data)

	case reflect.Interface:
		if target.NumMethod() == 0 {
			target.Set(reflect.ValueOf(data))
		}
	}

	return nil
}

// convertJSONNumber 根据目标类型转换json.Number
func convertJSONNumber(num json.Number, targetKind reflect.Kind) interface{} {
	switch targetKind {
	case reflect.String:
		return num.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := num.Int64(); err == nil {
			return i
		}
	case reflect.Float32, reflect.Float64:
		if f, err := num.Float64(); err == nil {
			return f
		}
	}
	// 默认尝试返回float64
	if f, err := num.Float64(); err == nil {
		return f
	}
	return num.String()
}

// 类型转换辅助函数

// toString 将任意类型转换为字符串
func ToString(data interface{}) string {
	switch v := data.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(v, 10)
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// toInt64 将任意类型转换为int64
func ToInt64(data interface{}) int64 {
	switch v := data.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case float64:
		return int64(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
		if f, err := v.Float64(); err == nil {
			return int64(f)
		}
	case string:
		if v == "" {
			return 0
		}
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int64(f)
		}
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return 0
}

// toInt 将任意类型转换为int
func ToInt(data interface{}) int {
	switch v := data.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
		if f, err := v.Float64(); err == nil {
			return int(f)
		}
	case string:
		if v == "" {
			return 0
		}
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return int(i)
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return int(f)
		}
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return 0
}

// toFloat64 将任意类型转换为float64
func ToFloat64(data interface{}) float64 {
	switch v := data.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f
		}
	case string:
		if v == "" {
			return 0
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	case bool:
		if v {
			return 1
		}
		return 0
	}
	return 0
}

// toBool 将任意类型转换为bool
func ToBool(data interface{}) bool {
	switch v := data.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	case int64:
		return v != 0
	case int:
		return v != 0
	case float64:
		return v != 0
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i != 0
		}
	}
	return false
}
