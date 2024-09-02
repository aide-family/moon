package cache

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aide-family/moon/pkg/util/ptr"

	"k8s.io/client-go/tools/cache"
)

var _ Reader = &reader{}

type reader struct {
	indexer cache.Indexer
	kind    string
}

func (r *reader) Get(_ context.Context, key string, out any) error {
	obj, exists, err := r.indexer.GetByKey(key)
	if err != nil {
		return fmt.Errorf("failed to get object by key %s: %w", key, err)
	}
	if !exists {
		return fmt.Errorf("object %s of kind %s not found", key, r.kind)
	}

	if err := r.assignObject(obj, out); err != nil {
		return fmt.Errorf("failed to assign object to output: %w", err)
	}
	return nil
}

func (r *reader) List(_ context.Context, out any) error {
	objs := r.indexer.List()

	v, err := ptr.EnforcePtr(out)
	if err != nil {
		return fmt.Errorf("output parameter must be a pointer: %w", err)
	}

	// Check if the output is a slice
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("output parameter must be a slice, but got %v", v.Kind())
	}

	// Get the type of the slice elements
	elemType := v.Type().Elem()
	resultSlice := reflect.MakeSlice(reflect.SliceOf(elemType), len(objs), len(objs))
	for i, obj := range objs {
		objVal := reflect.ValueOf(obj)

		// Check if the object type is assignable to the slice element type
		if objVal.Type().AssignableTo(elemType) {
			resultSlice.Index(i).Set(objVal)
		} else if objVal.Kind() == reflect.Ptr && objVal.Elem().Type().AssignableTo(elemType) {
			// If the object is a pointer, but the slice expects a non-pointer type
			resultSlice.Index(i).Set(objVal.Elem())
		} else {
			return fmt.Errorf("cache contained type %s, but slice expects %s", objVal.Type(), elemType)
		}
	}
	v.Set(resultSlice)
	return nil
}

// assignObject is a helper function to assign the object to the output using reflection
func (r *reader) assignObject(obj any, out any) error {
	outVal := reflect.ValueOf(out)
	objVal := reflect.ValueOf(obj)

	// Determine whether the output object types are the same
	if !objVal.Type().AssignableTo(outVal.Type()) {
		return fmt.Errorf("cache had type %s, but %s was asked for", objVal.Type(), outVal.Type())
	}
	// Perform the assignment
	outVal.Elem().Set(reflect.Indirect(objVal))
	return nil
}
