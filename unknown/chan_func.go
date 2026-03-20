package unknown

import (
	"reflect"
)

// Chan 转换为通道类型
func (u UnknownType) Chan() interface{} {
	if u.Value == nil {
		return nil
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() == reflect.Chan {
		return rv.Interface()
	}
	return nil
}

// MustChan 转换为通道，转换失败时返回默认值
func (u UnknownType) MustChan(defaultVal ...interface{}) interface{} {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}
	return u.Chan()
}

// Func 转换为函数类型
func (u UnknownType) Func() interface{} {
	if u.Value == nil {
		return nil
	}

	rv := reflect.ValueOf(u.Value)
	if rv.Kind() == reflect.Func {
		return rv.Interface()
	}
	return nil
}

// MustFunc 转换为函数，转换失败时返回默认值
func (u UnknownType) MustFunc(defaultVal ...interface{}) interface{} {
	if u.Value == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}
	return u.Func()
}
