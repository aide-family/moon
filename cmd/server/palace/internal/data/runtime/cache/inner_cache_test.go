package cache

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInnerCache_GetInformer(t *testing.T) {
	cache := NewCache(testScheme)

	obj := &TestStruct{Field1: "key1", Field2: 1}
	informer, err := cache.GetInformer(context.Background(), obj)
	assert.NoError(t, err)
	assert.NotNil(t, informer)
}

func TestInnerCache_GetInformerForKind(t *testing.T) {
	cache := NewCache(testScheme)

	informer, err := cache.GetInformerForKind(context.Background(), "TestStruct")
	assert.NoError(t, err)
	assert.NotNil(t, informer)
}

func TestInnerCache_Add(t *testing.T) {
	cache := NewCache(testScheme)

	obj := &TestStruct{Field1: "key1", Field2: 1}
	err := cache.Add(context.Background(), obj)
	assert.NoError(t, err)

	// Verify object was added
	in := &TestStruct{}
	err = cache.Get(context.Background(), "key1", in)
	assert.NoError(t, err)
	assert.Equal(t, obj, in)
}

func TestInnerCache_Update(t *testing.T) {
	cache := NewCache(testScheme)

	obj := &TestStruct{Field1: "key1", Field2: 1}
	err := cache.Add(context.Background(), obj)
	assert.NoError(t, err)
	// Update object
	obj.Field2 = 2
	err = cache.Update(context.Background(), obj)
	assert.NoError(t, err)

	// Verify object was updated
	in := &TestStruct{}
	err = cache.Get(context.Background(), "key1", in)
	assert.NoError(t, err)
	assert.Equal(t, obj, in)
	assert.Equal(t, 2, in.Field2)
}

func TestInnerCache_Replace(t *testing.T) {
	cache := NewCache(testScheme)

	obj1 := TestStruct{Field1: "key1", Field2: 1}
	obj2 := TestStruct{Field1: "key2", Field2: 2}
	cache.Add(context.Background(), obj1)

	err := cache.Replace(context.Background(), []TestStruct{obj2})
	assert.NoError(t, err)

	// Verify old object was replaced and new object exists
	out1 := &TestStruct{}
	err = cache.Get(context.Background(), "key1", out1)
	assert.Error(t, err) // Should return error as object with "key1" should be replaced

	out2 := &TestStruct{}
	err = cache.Get(context.Background(), "key2", out2)
	assert.NoError(t, err)
	assert.Equal(t, &obj2, out2)
}

func TestInnerCache_Delete(t *testing.T) {
	cache := NewCache(testScheme)

	obj := &TestStruct{Field1: "key1", Field2: 1}
	cache.Add(context.Background(), obj)

	err := cache.Delete(context.Background(), obj)
	assert.NoError(t, err)

	// Verify object was deleted
	in := &TestStruct{}
	err = cache.Get(context.Background(), "key1", in)
	assert.Error(t, err) // Should return error as object should be deleted
}

func TestInnerCache_List(t *testing.T) {
	cache := NewCache(testScheme)

	obj1 := &TestStruct{Field1: "key1", Field2: 1}
	obj2 := &TestStruct{Field1: "key2", Field2: 2}
	obj3 := &TestStruct{Field1: "key3", Field2: 3}

	err := cache.Add(context.Background(), obj1)
	err = cache.Add(context.Background(), obj2)
	err = cache.Add(context.Background(), obj3)
	assert.NoError(t, err)

	var out []TestStruct

	// 调用 List 方法
	err = cache.List(context.Background(), &out)
	assert.NoError(t, err)

	// 验证 List 输出的内容
	assert.Len(t, out, 3)
	assert.Contains(t, out, *obj1)
	assert.Contains(t, out, *obj2)
	assert.Contains(t, out, *obj3)
}
