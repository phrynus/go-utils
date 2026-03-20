package unknown

import (
	"fmt"
	"reflect"
	"strings"
)

// Map 转换为 map[string]interface{} 类型
func (u UnknownType) Map() map[string]interface{} {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]interface{}:
		return v
	case map[string]string:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result
	case map[string]int:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result
	case map[string]int64:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result
	case map[string]float64:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result
	case map[string]bool:
		result := make(map[string]interface{}, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() != reflect.Map {
		return nil
	}

	result := make(map[string]interface{})
	iter := rv.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		key := fmt.Sprintf("%v", k.Interface())
		result[key] = v.Interface()
	}
	return result
}

// MustMap 转换为 map[string]interface{}，转换失败时返回默认值
func (u UnknownType) MustMap(defaultVal ...map[string]interface{}) map[string]interface{} {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}
	return u.Map()
}

// StringMap 转换为 map[string]string 类型
func (u UnknownType) StringMap() map[string]string {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]string:
		return v
	case map[string]interface{}:
		result := make(map[string]string, len(v))
		for k, val := range v {
			result[k] = NewUnknown(val).String()
		}
		return result
	}

	m := u.Map()
	if m == nil {
		return nil
	}

	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = NewUnknown(v).String()
	}
	return result
}

// IntMap 转换为 map[string]int64 类型
func (u UnknownType) IntMap() map[string]int64 {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]int64:
		return v
	case map[string]int:
		result := make(map[string]int64, len(v))
		for k, val := range v {
			result[k] = int64(val)
		}
		return result
	case map[string]int32:
		result := make(map[string]int64, len(v))
		for k, val := range v {
			result[k] = int64(val)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]int64, len(v))
		for k, val := range v {
			result[k] = NewUnknown(val).Int()
		}
		return result
	}

	m := u.Map()
	if m == nil {
		return nil
	}

	result := make(map[string]int64, len(m))
	for k, v := range m {
		result[k] = NewUnknown(v).Int()
	}
	return result
}

// FloatMap 转换为 map[string]float64 类型
func (u UnknownType) FloatMap() map[string]float64 {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]float64:
		return v
	case map[string]interface{}:
		result := make(map[string]float64, len(v))
		for k, val := range v {
			result[k] = NewUnknown(val).Float()
		}
		return result
	}

	m := u.Map()
	if m == nil {
		return nil
	}

	result := make(map[string]float64, len(m))
	for k, v := range m {
		result[k] = NewUnknown(v).Float()
	}
	return result
}

// BoolMap 转换为 map[string]bool 类型
func (u UnknownType) BoolMap() map[string]bool {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]bool:
		return v
	case map[string]interface{}:
		result := make(map[string]bool, len(v))
		for k, val := range v {
			result[k] = NewUnknown(val).Bool()
		}
		return result
	}

	m := u.Map()
	if m == nil {
		return nil
	}

	result := make(map[string]bool, len(m))
	for k, v := range m {
		result[k] = NewUnknown(v).Bool()
	}
	return result
}

// MapKeys 获取 Map 的所有 Key
func (u UnknownType) MapKeys() []interface{} {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]interface{}:
		keys := make([]interface{}, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys
	case map[string]string:
		keys := make([]interface{}, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() != reflect.Map {
		return nil
	}

	keys := rv.MapKeys()
	result := make([]interface{}, len(keys))
	for i, key := range keys {
		result[i] = key.Interface()
	}
	return result
}

// MapValues 获取 Map 的所有 Value
func (u UnknownType) MapValues() []interface{} {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case map[string]interface{}:
		values := make([]interface{}, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values
	case map[string]string:
		values := make([]interface{}, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() != reflect.Map {
		return nil
	}

	result := make([]interface{}, 0, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		result = append(result, iter.Value().Interface())
	}
	return result
}

// GetMapValue 获取 Map 中指定 key 的值，支持嵌套 key（用 . 分隔）
// 例如: GetMapValue("user.profile.name")
func (u UnknownType) GetMapValue(key string) interface{} {
	if u.Value == nil {
		return nil
	}

	keys := strings.Split(key, ".")
	var current interface{} = u.Value

	for _, k := range keys {
		if current == nil {
			return nil
		}

		switch m := current.(type) {
		case map[string]interface{}:
			if v, ok := m[k]; ok {
				current = v
			} else {
				return nil
			}
		case map[string]string:
			if v, ok := m[k]; ok {
				current = v
			} else {
				return nil
			}
		default:
			rv := reflect.ValueOf(m)
			if rv.Kind() == reflect.Map {
				mapKey := reflect.ValueOf(k)
				value := rv.MapIndex(mapKey)
				if value.IsValid() {
					current = value.Interface()
				} else {
					return nil
				}
			} else {
				return nil
			}
		}
	}

	return current
}
