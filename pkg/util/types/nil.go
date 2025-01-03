package types

import (
	"reflect"
)

// IsNil 判断是否为nil
func IsNil(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}

// IsNotNil 判断是否不为nil
func IsNotNil(i interface{}) bool {
	return !IsNil(i)
}
