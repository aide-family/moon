package cache

import (
	"context"
	"sort"
	"testing"

	"k8s.io/client-go/tools/cache"
)

func TestReader_Get(t *testing.T) {

	indexer := cache.NewIndexer(ts.KeyFunc, cache.Indexers{})
	reader := &reader{
		indexer: indexer,
		kind:    "TestStruct",
	}

	err := indexer.Add(&TestStruct{Field1: "key1", Field2: 1})
	if err != nil {
		t.Errorf("Add returned unexpected error: %v", err)
	}

	var out TestStruct

	// Test case: successful get
	err = reader.Get(context.Background(), "key1", &out)
	if err != nil {
		t.Errorf("Get returned unexpected error: %v", err)
	}
	if out.Field1 != "key1" || out.Field2 != 1 {
		t.Errorf("Get did not return the expected object, got: %+v", out)
	}

	// Test case: object not found
	err = reader.Get(context.Background(), "key2", &out)
	if err == nil || err.Error() != "object key2 of kind TestStruct not found" {
		t.Errorf("Get did not return the expected error for missing key, got: %v", err)
	}

	// Test case: incorrect output type
	var invalidOut string
	err = reader.Get(context.Background(), "key1", &invalidOut)
	if err == nil {
		t.Errorf("Get did not return the expected error for incorrect output type, got: %v", err)
	} else {
		t.Log(err)
	}
}

func TestReader_List(t *testing.T) {

	indexer := cache.NewIndexer(ts.KeyFunc, cache.Indexers{})
	reader := &reader{
		indexer: indexer,
		kind:    "TestStruct",
	}
	err := indexer.Add(&TestStruct{Field1: "key1", Field2: 1})
	if err != nil {
		t.Errorf("Add returned unexpected error: %v", err)
	}
	err = indexer.Add(&TestStruct{Field1: "key2", Field2: 2})
	if err != nil {
		t.Errorf("Add returned unexpected error: %v", err)
	}

	var out []TestStruct

	// Test case: successful list
	err = reader.List(context.Background(), &out)
	if err != nil {
		t.Errorf("List returned unexpected error: %v", err)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Field2 < out[j].Field2
	})

	if len(out) != 2 || out[0].Field1 != "key1" || out[1].Field1 != "key2" {
		t.Errorf("List did not return the expected objects, got: %+v", out)
	}

	// Test case: incorrect output type
	var invalidOut string
	err = reader.List(context.Background(), &invalidOut)
	if err == nil {
		t.Errorf("List did not return the expected error for incorrect output type, got: %v", err)
	} else {
		t.Log(err)
	}
}
