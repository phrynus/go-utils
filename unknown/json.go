package unknown

import (
	"encoding/json"
)

// Json JSON序列化输出（自动按 key 排序）
func (u UnknownType) Json() []byte {
	jsonBytes, err := json.Marshal(u.Value)
	if err != nil {
		return []byte{}
	}
	return jsonBytes
}

// JsonString JSON序列化输出为字符串（自动按 key 排序）
func (u UnknownType) JsonString() string {
	return string(u.Json())
}

// JsonIndent JSON格式化输出（自动按 key 排序）
func (u UnknownType) JsonIndent() []byte {
	jsonBytes, err := json.MarshalIndent(u.Value, "", "  ")
	if err != nil {
		return []byte{}
	}
	return jsonBytes
}

// JsonIndentString JSON格式化输出为字符串（自动按 key 排序）
func (u UnknownType) JsonIndentString() string {
	return string(u.JsonIndent())
}

// MustJson JSON序列化，失败时返回默认值
func (u UnknownType) MustJson(defaultVal ...[]byte) []byte {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return []byte("null")
	}
	return u.Json()
}
