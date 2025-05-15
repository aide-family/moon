package validate

import (
	"reflect"
)

// IsNil checks if the given interface is nil or a pointer to nil.
func IsNil(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}

// IsNotNil checks if the given interface is not nil or a pointer to nil.
func IsNotNil(i interface{}) bool {
	return !IsNil(i)
}
