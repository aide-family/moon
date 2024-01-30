package types

import (
	"reflect"
)

type Int interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Uint interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type String interface {
	~string
}

type Bool interface {
	~bool
}

type Number interface {
	Int | Uint | Float
}

func IsNil(v interface{}) bool {
	// 判断类型+值
	t := reflect.TypeOf(v)
	return t == nil || reflect.ValueOf(v).IsNil()
}
