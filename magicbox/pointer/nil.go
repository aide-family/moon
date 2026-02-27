package pointer

import "reflect"

// IsNil checks if the given interface is nil or a pointer to nil.
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	for {
		switch v.Kind() {
		case reflect.Ptr, reflect.Interface:
			if v.IsNil() {
				return true
			}
			v = v.Elem()
		case reflect.Func, reflect.Slice, reflect.Map, reflect.Chan:
			return v.IsNil()
		default:
			return false
		}
	}
}

// IsNotNil checks if the given interface is not nil or a pointer to nil.
func IsNotNil(i interface{}) bool {
	return !IsNil(i)
}
