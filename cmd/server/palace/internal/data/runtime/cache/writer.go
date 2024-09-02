package cache

import (
	"context"
	"fmt"
	"reflect"

	"k8s.io/client-go/tools/cache"
)

var _ Writer = &writer{}

type writer struct {
	indexer cache.Indexer
}

func (w *writer) AddIndexers(indexers cache.Indexers) error {
	return w.indexer.AddIndexers(indexers)
}

func (w *writer) Add(ctx context.Context, object any) error {
	return w.Update(ctx, object)
}

func (w *writer) Update(_ context.Context, object any) error {
	return w.indexer.Update(object)
}

func (w *writer) Delete(_ context.Context, object any) error {
	return w.indexer.Delete(object)
}

func (w *writer) Replace(ctx context.Context, objects any) error {
	v := reflect.ValueOf(objects)

	// Check if the output is a slice
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("objects must be a slice, but got %v", v.Kind())
	}
	err := w.indexer.Replace([]any{}, "")
	if err != nil {
		return err
	}
	for i := 0; i < v.Len(); i++ {
		ptr := reflect.New(v.Index(i).Type()) // Create a new pointer to the element type
		ptr.Elem().Set(v.Index(i))
		err = w.Update(ctx, ptr.Interface())
		if err != nil {
			return err
		}
	}
	return nil
}
