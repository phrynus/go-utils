package unknown

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// SmartUnmarshal 根据目标类型进行智能转换
func (u UnknownType) SmartUnmarshal(v interface{}) error {
	data := u.Value

	if v == nil {
		return ErrNilValue
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("目标必须是非空指针: %w", ErrNotConvertible)
	}

	if str, ok := data.(string); ok {
		if err := json.Unmarshal([]byte(str), &data); err != nil {
			// 非 JSON 字符串，保留原值
		}
	}

	return setValue(rv.Elem(), data)
}

// --------------------------------------------------------------------
// setValue 内部实现
// --------------------------------------------------------------------

// fillSlice 填充切片类型
func fillSlice(target reflect.Value, data []interface{}) error {
	if len(data) == 0 {
		if target.Kind() == reflect.Slice {
			target.Set(reflect.MakeSlice(target.Type(), 0, 0))
		}
		return nil
	}

	slice := reflect.MakeSlice(target.Type(), len(data), len(data))
	for i, item := range data {
		if err := setValue(slice.Index(i), item); err != nil {
			return err
		}
	}
	target.Set(slice)
	return nil
}

// fillMap 填充映射类型
func fillMap(target reflect.Value, data map[string]interface{}) error {
	mapValue := reflect.MakeMap(target.Type())
	elemType := target.Type().Elem()

	for key, value := range data {
		elemValue := reflect.New(elemType).Elem()
		if err := setValue(elemValue, value); err != nil {
			return err
		}
		mapValue.SetMapIndex(reflect.ValueOf(key), elemValue)
	}

	if target.CanSet() {
		target.Set(mapValue)
	}
	return nil
}

// fillStruct 递归填充结构体字段
func fillStruct(target reflect.Value, data map[string]interface{}) error {
	if !target.IsValid() {
		return fmt.Errorf("目标结构体无效: %w", ErrNotConvertible)
	}

	targetType := target.Type()

	if target.Kind() == reflect.Ptr {
		if target.IsNil() {
			target.Set(reflect.New(targetType.Elem()))
		}
		target = target.Elem()
		targetType = target.Type()
	}

	if target.Kind() != reflect.Struct {
		return fmt.Errorf("目标不是结构体: %w", ErrNotConvertible)
	}

	for i := 0; i < target.NumField(); i++ {
		field := target.Field(i)
		fieldType := targetType.Field(i)

		if !fieldType.IsExported() {
			continue
		}

		jsonTag := fieldType.Tag.Get("json")
		fieldName := fieldType.Name
		if jsonTag != "" && jsonTag != "-" {
			idx := strings.Index(jsonTag, ",")
			if idx > 0 {
				fieldName = jsonTag[:idx]
			} else {
				fieldName = jsonTag
			}
		}

		var rawValue interface{}
		var exists bool

		if rawValue, exists = data[fieldName]; !exists {
			if rawValue, exists = data[strings.ToLower(fieldName)]; !exists {
				if rawValue, exists = data[strings.ToLower(fieldType.Name)]; !exists {
					snakeName := toSnakeCase(fieldName)
					if rawValue, exists = data[snakeName]; !exists {
						rawValue, exists = data[toSnakeCase(fieldType.Name)]
					}
				}
			}
		}

		if exists {
			if err := setValue(field, rawValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// toSnakeCase 将 CamelCase 转换为 snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// setValue 根据目标类型设置值
func setValue(target reflect.Value, data interface{}) error {
	if !target.CanSet() {
		return nil
	}

	if data == nil {
		switch target.Kind() {
		case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
			target.Set(reflect.Zero(target.Type()))
		}
		return nil
	}

	unknown := NewUnknown(data)

	switch target.Kind() {
	case reflect.String:
		target.SetString(unknown.String())
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		target.SetInt(unknown.Int())
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		target.SetUint(unknown.Uint())
		return nil

	case reflect.Float32, reflect.Float64:
		target.SetFloat(unknown.Float())
		return nil

	case reflect.Bool:
		target.SetBool(unknown.Bool())
		return nil

	case reflect.Complex64, reflect.Complex128:
		target.SetComplex(unknown.Complex())
		return nil

	case reflect.Ptr:
		elemType := target.Type().Elem()

		if s, ok := data.(string); ok && elemType.Kind() == reflect.String && s == "" {
			return nil
		}

		if elemType.Kind() != reflect.String && reflect.DeepEqual(reflect.Zero(elemType).Interface(), data) {
			return nil
		}

		if target.IsNil() {
			target.Set(reflect.New(elemType))
		}
		return setValue(target.Elem(), data)

	case reflect.Struct:
		if target.Type() == reflect.TypeOf(time.Time{}) {
			t := unknown.Time()
			if !t.IsZero() || unknown.IsZero() {
				target.Set(reflect.ValueOf(t))
			}
			return nil
		}

		if target.Type() == reflect.TypeOf(time.Duration(0)) {
			target.Set(reflect.ValueOf(unknown.Duration()))
			return nil
		}

		if m, ok := data.(map[string]interface{}); ok {
			return fillStruct(target, m)
		}
		return nil

	case reflect.Slice:
		if b, ok := data.([]byte); ok {
			if target.Type().Elem().Kind() == reflect.Uint8 {
				target.SetBytes(b)
			} else {
				slice := reflect.MakeSlice(target.Type(), len(b), len(b))
				for i, v := range b {
					slice.Index(i).SetUint(uint64(v))
				}
				target.Set(slice)
			}
			return nil
		}

		if s, ok := data.([]interface{}); ok {
			return fillSlice(target, s)
		}

		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Slice {
			length := rv.Len()
			slice := reflect.MakeSlice(target.Type(), length, length)
			for i := 0; i < length; i++ {
				if err := setValue(slice.Index(i), rv.Index(i).Interface()); err != nil {
					return err
				}
			}
			target.Set(slice)
		}
		return nil

	case reflect.Map:
		if m, ok := data.(map[string]interface{}); ok {
			return fillMap(target, m)
		}
		return nil

	case reflect.Interface:
		if target.NumMethod() == 0 {
			target.Set(reflect.ValueOf(data))
		}
		return nil

	case reflect.Array:
		rv := reflect.ValueOf(data)
		if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
			length := target.Len()
			srcLen := rv.Len()
			if srcLen < length {
				length = srcLen
			}
			for i := 0; i < length; i++ {
				if err := setValue(target.Index(i), rv.Index(i).Interface()); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return nil
}
