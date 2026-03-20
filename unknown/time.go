package unknown

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// TimeFormats 定义支持的常用时间格式（按优先级排序）
var TimeFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	time.RFC1123Z,
	time.RFC1123,
	time.RFC850,
	time.RFC822Z,
	time.RFC822,
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02T15:04:05.999Z",
	"2006-01-02T15:04:05.999",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05.999999999",
	"2006-01-02 15:04:05.999",
	"2006-01-02 15:04:05",
	"2006/01/02 15:04:05",
	"2006/01/02 15:04",
	"2006/01/02",
	"2006-01-02",
	"01/02/2006 15:04:05",
	"01/02/2006",
	"02/01/2006 15:04:05",
	"02/01/2006",
	"20060102150405",
	"200601021504",
	"20060102",
}

// Time 转换为 time.Time 类型
// 支持: time.Time, *time.Time, 时间戳(int/uint/float/字符串), 常见日期字符串格式
func (u UnknownType) Time() time.Time {
	if u.Value == nil {
		return time.Time{}
	}

	switch v := u.Value.(type) {
	case time.Time:
		return v
	case *time.Time:
		if v != nil {
			return *v
		}
		return time.Time{}
	case string:
		return parseTimeString(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return normalizeTimestamp(i)
		}
		if f, err := v.Float64(); err == nil {
			return normalizeTimestamp(int64(f))
		}
		return time.Time{}
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return normalizeTimestamp(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return normalizeTimestamp(int64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return normalizeTimestamp(int64(rv.Float()))
	case reflect.Interface, reflect.Ptr:
		if !rv.IsNil() {
			return NewUnknown(rv.Elem().Interface()).Time()
		}
	}
	return time.Time{}
}

// parseTimeString 解析时间字符串
func parseTimeString(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	// 先尝试作为格式字符串解析（8位/14位纯数字优先考虑日期格式）
	if len(s) == 8 || len(s) == 14 {
		// 检查是否全为数字
		isAllDigits := true
		for _, c := range s {
			if c < '0' || c > '9' {
				isAllDigits = false
				break
			}
		}
		if isAllDigits {
			// 优先尝试日期格式解析
			format := "20060102"
			if len(s) == 14 {
				format = "20060102150405"
			}
			if t, err := time.Parse(format, s); err == nil {
				return t
			}
		}
	}

	if ts := tryParseTimestamp(s); !ts.IsZero() {
		return ts
	}

	if strings.HasPrefix(s, "/Date(") {
		if ts := parseJSONDate(s); !ts.IsZero() {
			return ts
		}
	}

	for _, format := range TimeFormats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	return time.Time{}
}

// tryParseTimestamp 尝试将字符串解析为时间戳
func tryParseTimestamp(s string) time.Time {
	// 如果包含非数字、小数点、负号之外的内容，不尝试解析为时间戳
	if strings.TrimFunc(s, func(r rune) bool {
		return r >= '0' && r <= '9' || r == '.' || r == '-'
	}) != "" {
		return time.Time{}
	}

	negative := strings.HasPrefix(s, "-")
	numStr := s
	if negative {
		numStr = s[1:]
	}

	dotIdx := strings.Index(numStr, ".")
	var intPart, fracPart string
	if dotIdx >= 0 {
		intPart = numStr[:dotIdx]
		fracPart = numStr[dotIdx+1:]
	} else {
		intPart = numStr
	}

	intPart = strings.TrimSpace(intPart)
	fracPart = strings.TrimSpace(fracPart)

	if intPart == "" {
		return time.Time{}
	}

	i, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return time.Time{}
	}

	if negative {
		i = -i
	}

	abs := i
	if abs < 0 {
		abs = -abs
	}

	// 根据位数判断是否为时间戳
	// 10 位: 秒级时间戳 (1970-2099年左右)
	// 12 位: 毫秒时间戳
	// 13 位: 毫秒时间戳 (带小数部分)
	switch len(intPart) {
	case 10:
		// 秒级时间戳
		return time.Unix(i, 0)
	case 12:
		// 毫秒时间戳
		return time.UnixMilli(i)
	case 13:
		// 毫秒时间戳 (带小数部分)
		if fracPart != "" {
			f, _ := strconv.ParseFloat(s, 64)
			return normalizeTimestamp(int64(f))
		}
		return time.UnixMilli(i)
	default:
		// 其他情况返回零值，让格式解析尝试
		return time.Time{}
	}
}

// parseJSONDate 解析 JSON Date 格式: /Date(毫秒)/
func parseJSONDate(s string) time.Time {
	if len(s) < 9 {
		return time.Time{}
	}

	start := strings.Index(s, "(")
	end := strings.Index(s, ")")
	if start < 0 || end < 0 || end <= start+1 {
		return time.Time{}
	}

	msStr := s[start+1 : end]
	ms, err := strconv.ParseInt(msStr, 10, 64)
	if err != nil {
		return time.Time{}
	}

	return time.UnixMilli(ms)
}

// normalizeTimestamp 根据时间戳范围归一化为 time.Time
func normalizeTimestamp(i int64) time.Time {
	abs := i
	if abs < 0 {
		abs = -abs
	}

	switch {
	case abs < 1e10 && abs > 1e8:
		if i > -6e17 && i < 6e17 {
			return time.Unix(0, i)
		}
		return time.Unix(i, 0)
	case abs >= 1e12 && abs < 1e15:
		return time.UnixMilli(i)
	case abs >= 1e15:
		return time.Unix(0, i)
	default:
		return time.Unix(i, 0)
	}
}

// MustTime 转换为 time.Time，转换失败时返回默认值
func (u UnknownType) MustTime(defaultVal ...time.Time) time.Time {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return time.Time{}
	}
	return u.Time()
}

// Unix 获取 Unix 时间戳（秒）
func (u UnknownType) Unix() int64 {
	return u.Time().Unix()
}

// UnixMilli 获取 Unix 时间戳（毫秒）
func (u UnknownType) UnixMilli() int64 {
	return u.Time().UnixMilli()
}

// UnixNano 获取 Unix 时间戳（纳秒）
func (u UnknownType) UnixNano() int64 {
	return u.Time().UnixNano()
}
