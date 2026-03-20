package unknown

import (
	"reflect"
)

// Pointer 提取指针/接口指向的值
func (u UnknownType) Pointer() interface{} {
	if u.Value == nil {
		return nil
	}

	rv := reflect.ValueOf(u.Value)
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return nil
		}
		return rv.Elem().Interface()
	case reflect.Interface:
		if rv.IsNil() {
			return nil
		}
		return rv.Elem().Interface()
	}
	return nil
}

// Deref 解引用，返回实际指向的值（递归解引用直到非指针）
func (u UnknownType) Deref() interface{} {
	if u.Value == nil {
		return nil
	}

	rv := reflect.ValueOf(u.Value)
	kind := rv.Kind()

	for kind == reflect.Ptr || kind == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
		kind = rv.Kind()
	}

	return rv.Interface()
}
