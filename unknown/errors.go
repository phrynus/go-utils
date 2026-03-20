package unknown

import (
	"errors"
	"reflect"
)

var (
	ErrNotConvertible = errors.New("值无法转换为目标类型")
	ErrNilValue       = errors.New("值为 nil")
)

// UnknownType 通用类型封装，支持任意类型到目标类型的智能转换
type UnknownType struct {
	Value any
	Type  reflect.Type
}

// NewUnknown 创建 UnknownType 实例
func NewUnknown(data ...any) UnknownType {
	if len(data) == 1 {
		return UnknownType{
			Value: data[0],
			Type:  reflect.TypeOf(data[0]),
		}
	}
	return UnknownType{
		Value: data,
		Type:  reflect.TypeOf(data),
	}
}
