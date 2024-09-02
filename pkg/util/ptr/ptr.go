package ptr

import (
	"fmt"
	"reflect"
)

// EnforcePtr takes an interface `obj` and returns the dereferenced value
// if `obj` is a non-nil pointer. It returns an error if `obj` is not a pointer,
// is nil, or is an invalid type.
func EnforcePtr(obj any) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Pointer {
		if v.Kind() == reflect.Invalid {
			return reflect.Value{}, fmt.Errorf("invalid type: expected pointer")
		}
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v", v.Type())
	}
	if v.IsNil() {
		return reflect.Value{}, fmt.Errorf("nil pointer received")
	}
	return v.Elem(), nil
}

// GenerateElementPtrBySlice takes a slice or a pointer to a slice as input
// and returns a pointer to a new zero-valued element of the slice's element type.
// It returns an error if the input is not a slice or a pointer to a slice.
func GenerateElementPtrBySlice(slice any) (any, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Slice {
		elementType := v.Type().Elem()
		return reflect.New(elementType).Interface(), nil
	} else if v.Kind() == reflect.Pointer {
		elementType := v.Elem().Type().Elem()
		return reflect.New(elementType).Interface(), nil
	}
	return nil, fmt.Errorf("input must be a slice or pointer to a slice")
}

// GenerateElementBySlice takes a slice or a pointer to a slice as input
// and returns a new zero-valued element of the slice's element type.
// It returns an error if the input is not a slice or a pointer to a slice.
func GenerateElementBySlice(slice any) (any, error) {
	v := reflect.ValueOf(slice)
	if v.Kind() == reflect.Slice {
		elementType := v.Type().Elem()
		return reflect.New(elementType).Elem().Interface(), nil
	} else if v.Kind() == reflect.Pointer {
		elementType := v.Elem().Type().Elem()
		return reflect.New(elementType).Elem().Interface(), nil
	}
	return nil, fmt.Errorf("input must be a slice or pointer to a slice")
}
