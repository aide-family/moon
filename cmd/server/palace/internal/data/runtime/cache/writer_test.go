package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter_Add(t *testing.T) {
	indexer := NewMockIndexer()
	w := &writer{indexer: indexer}

	obj1 := &TestStruct{Field1: "key1", Field2: 1}
	obj2 := &TestStruct{Field1: "key2", Field2: 2}

	err := w.Add(context.TODO(), obj1)
	assert.NoError(t, err)
	err = w.Add(context.TODO(), obj2)
	assert.NoError(t, err)

	// Verify that objects are added
	key1, _ := ts.KeyFunc(obj1)
	_, exists1, _ := indexer.GetByKey(key1)
	assert.True(t, exists1)

	key2, _ := ts.KeyFunc(obj2)
	_, exists2, _ := indexer.GetByKey(key2)
	assert.True(t, exists2)
}

func TestWriter_Update(t *testing.T) {
	indexer := NewMockIndexer()
	w := &writer{indexer: indexer}

	obj := &TestStruct{Field1: "key1", Field2: 1}
	err := indexer.Add(obj)
	assert.NoError(t, err)

	// Update the object
	obj.Field2 = 2
	err = w.Update(context.TODO(), obj)
	assert.NoError(t, err)

	// Verify the update
	key, _ := ts.KeyFunc(obj)
	updatedObj, exists, _ := indexer.GetByKey(key)
	assert.True(t, exists)
	assert.Equal(t, 2, updatedObj.(*TestStruct).Field2)
}

func TestWriter_Delete(t *testing.T) {
	indexer := NewMockIndexer()
	w := &writer{indexer: indexer}

	obj := &TestStruct{Field1: "key1", Field2: 1}
	err := indexer.Add(obj)
	assert.NoError(t, err)

	err = w.Delete(context.TODO(), obj)
	assert.NoError(t, err)

	// Verify the deletion
	key, _ := ts.KeyFunc(obj)
	_, exists, _ := indexer.GetByKey(key)
	assert.False(t, exists)
}

func TestWriter_Replace(t *testing.T) {
	indexer := NewMockIndexer()
	w := &writer{indexer: indexer}

	obj1 := &TestStruct{Field1: "key1", Field2: 1}
	obj2 := &TestStruct{Field1: "key2", Field2: 2}

	err := w.Replace(context.TODO(), []TestStruct{*obj1, *obj2})
	assert.NoError(t, err)

	// Verify the replacement
	key1, _ := ts.KeyFunc(obj1)
	_, exists1, _ := indexer.GetByKey(key1)
	assert.True(t, exists1)

	key2, _ := ts.KeyFunc(obj2)
	_, exists2, _ := indexer.GetByKey(key2)
	assert.True(t, exists2)

	// Ensure that only the objects in the replace call exist in the indexer
	if keys := indexer.ListKeys(); len(keys) != 2 {
		t.Fatalf("expected 2 keys, but got %d", len(keys))
	}
}
