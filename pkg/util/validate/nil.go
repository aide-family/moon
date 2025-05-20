package validate

import (
	"reflect"
)

// IsNil checks if the given interface is nil or a pointer to nil.
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Ptr, reflect.Func:
		return v.IsNil()
	default:
		return false
	}
}

// IsNotNil checks if the given interface is not nil or a pointer to nil.
func IsNotNil(i interface{}) bool {
	return !IsNil(i)
}
