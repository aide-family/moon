package types

import (
	"reflect"
)

// Of 获取对象指针
func Of[T any](v T) *T {
	if IsNil(v) {
		return nil
	}
	return &v
}

// UnwrapOr 解包指针， 如果为nil则返回指定的默认值(没有指定，则返回go默认值), 否则返回值本身
func UnwrapOr[T any](p *T, fallback ...T) T {
	if !IsNil(p) {
		return *p
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	var t T
	return t
}

// ExtractPointerOr 解包多层指针， 如果为nil则返回指定的默认值(没有指定，则返回go默认值), 否则返回值本身
func ExtractPointerOr[T any](value any, fallback ...T) T {
	val, ok := extractPointer(value).(T)
	if ok {
		return val
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	var t T
	return t
}

func extractPointer(value any) any {
	t := reflect.TypeOf(value)
	v := reflect.ValueOf(value)
	if IsNil(t) {
		return nil
	}
	if t.Kind() != reflect.Pointer {
		return value
	}
	return extractPointer(v.Elem().Interface())
}
