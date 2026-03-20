package unknown

import (
	"reflect"
)

// Bytes 转换为 []byte 类型
// 支持: []byte, string, []rune, 数字序列化
func (u UnknownType) Bytes() []byte {
	if u.Value == nil {
		return nil
	}

	switch v := u.Value.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	case []rune:
		return []byte(string(v))
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.Uint8 {
		return rv.Bytes()
	}

	return []byte(u.String())
}

// MustBytes 转换为 []byte，转换失败时返回默认值
func (u UnknownType) MustBytes(defaultVal ...[]byte) []byte {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}
	return u.Bytes()
}
