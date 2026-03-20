package unknown

import "reflect"

// TypeName 获取类型的字符串名称
func (u UnknownType) TypeName() string {
	if u.Type == nil {
		return "nil"
	}
	return u.Type.String()
}

// KindName 获取 Kind 的字符串名称
func (u UnknownType) KindName() string {
	if u.Type == nil {
		return "nil"
	}
	return u.Type.Kind().String()
}

// Kind 返回值的 reflect.Kind
func (u UnknownType) Kind() reflect.Kind {
	if u.Type == nil {
		return reflect.Invalid
	}
	return u.Type.Kind()
}
