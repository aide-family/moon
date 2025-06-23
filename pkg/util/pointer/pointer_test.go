package pointer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOf(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "int pointer",
			input:    42,
			expected: 42,
		},
		{
			name:     "string pointer",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "bool pointer",
			input:    true,
			expected: true,
		},
		{
			name:     "float pointer",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "zero value int",
			input:    0,
			expected: 0,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.input.(type) {
			case int:
				ptr := Of(v)
				assert.NotNil(t, ptr)
				assert.Equal(t, tt.expected, *ptr)
			case string:
				ptr := Of(v)
				assert.NotNil(t, ptr)
				assert.Equal(t, tt.expected, *ptr)
			case bool:
				ptr := Of(v)
				assert.NotNil(t, ptr)
				assert.Equal(t, tt.expected, *ptr)
			case float64:
				ptr := Of(v)
				assert.NotNil(t, ptr)
				assert.Equal(t, tt.expected, *ptr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		ptr      interface{}
		expected interface{}
	}{
		{
			name:     "nil int pointer",
			ptr:      (*int)(nil),
			expected: 0,
		},
		{
			name:     "nil string pointer",
			ptr:      (*string)(nil),
			expected: "",
		},
		{
			name:     "nil bool pointer",
			ptr:      (*bool)(nil),
			expected: false,
		},
		{
			name:     "valid int pointer",
			ptr:      Of(42),
			expected: 42,
		},
		{
			name:     "valid string pointer",
			ptr:      Of("hello"),
			expected: "hello",
		},
		{
			name:     "valid bool pointer",
			ptr:      Of(true),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch ptr := tt.ptr.(type) {
			case *int:
				result := Get(ptr)
				assert.Equal(t, tt.expected, result)
			case *string:
				result := Get(ptr)
				assert.Equal(t, tt.expected, result)
			case *bool:
				result := Get(ptr)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetOr(t *testing.T) {
	tests := []struct {
		name     string
		ptr      interface{}
		def      interface{}
		expected interface{}
	}{
		{
			name:     "nil int pointer with default",
			ptr:      (*int)(nil),
			def:      100,
			expected: 100,
		},
		{
			name:     "nil string pointer with default",
			ptr:      (*string)(nil),
			def:      "default",
			expected: "default",
		},
		{
			name:     "valid int pointer with default",
			ptr:      Of(42),
			def:      100,
			expected: 42,
		},
		{
			name:     "valid string pointer with default",
			ptr:      Of("hello"),
			def:      "default",
			expected: "hello",
		},
		{
			name:     "nil bool pointer with default",
			ptr:      (*bool)(nil),
			def:      true,
			expected: true,
		},
		{
			name:     "valid bool pointer with default",
			ptr:      Of(false),
			def:      true,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch ptr := tt.ptr.(type) {
			case *int:
				def := tt.def.(int)
				result := GetOr(ptr, def)
				assert.Equal(t, tt.expected, result)
			case *string:
				def := tt.def.(string)
				result := GetOr(ptr, def)
				assert.Equal(t, tt.expected, result)
			case *bool:
				def := tt.def.(bool)
				result := GetOr(ptr, def)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetOrZero(t *testing.T) {
	tests := []struct {
		name     string
		ptr      interface{}
		expected interface{}
		ok       bool
	}{
		{
			name:     "nil int pointer",
			ptr:      (*int)(nil),
			expected: 0,
			ok:       false,
		},
		{
			name:     "nil string pointer",
			ptr:      (*string)(nil),
			expected: "",
			ok:       false,
		},
		{
			name:     "nil bool pointer",
			ptr:      (*bool)(nil),
			expected: false,
			ok:       false,
		},
		{
			name:     "valid int pointer",
			ptr:      Of(42),
			expected: 42,
			ok:       true,
		},
		{
			name:     "valid string pointer",
			ptr:      Of("hello"),
			expected: "hello",
			ok:       true,
		},
		{
			name:     "valid bool pointer",
			ptr:      Of(true),
			expected: true,
			ok:       true,
		},
		{
			name:     "zero value int pointer",
			ptr:      Of(0),
			expected: 0,
			ok:       true,
		},
		{
			name:     "empty string pointer",
			ptr:      Of(""),
			expected: "",
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch ptr := tt.ptr.(type) {
			case *int:
				result, ok := GetOrZero(ptr)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.ok, ok)
			case *string:
				result, ok := GetOrZero(ptr)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.ok, ok)
			case *bool:
				result, ok := GetOrZero(ptr)
				assert.Equal(t, tt.expected, result)
				assert.Equal(t, tt.ok, ok)
			}
		})
	}
}

func TestComplexTypes(t *testing.T) {
	// 测试结构体指针
	type Person struct {
		Name string
		Age  int
	}

	person := Person{Name: "Alice", Age: 30}
	ptr := Of(person)
	assert.Equal(t, person, *ptr)

	// 测试切片指针
	slice := []int{1, 2, 3}
	slicePtr := Of(slice)
	assert.Equal(t, slice, *slicePtr)

	// 测试映射指针
	m := map[string]int{"a": 1, "b": 2}
	mapPtr := Of(m)
	assert.Equal(t, m, *mapPtr)

	// 测试接口指针
	var iface interface{} = "interface value"
	ifacePtr := Of(iface)
	assert.Equal(t, iface, *ifacePtr)
}

func TestNilHandling(t *testing.T) {
	// 测试各种类型的 nil 指针
	var intPtr *int
	var stringPtr *string
	var boolPtr *bool
	var slicePtr *[]int
	var mapPtr *map[string]int

	// Get 函数测试
	assert.Equal(t, 0, Get(intPtr))
	assert.Equal(t, "", Get(stringPtr))
	assert.Equal(t, false, Get(boolPtr))
	assert.Equal(t, []int(nil), Get(slicePtr))
	assert.Equal(t, map[string]int(nil), Get(mapPtr))

	// GetOr 函数测试
	assert.Equal(t, 100, GetOr(intPtr, 100))
	assert.Equal(t, "default", GetOr(stringPtr, "default"))
	assert.Equal(t, true, GetOr(boolPtr, true))
	assert.Equal(t, []int{1, 2, 3}, GetOr(slicePtr, []int{1, 2, 3}))
	assert.Equal(t, map[string]int{"a": 1}, GetOr(mapPtr, map[string]int{"a": 1}))

	// GetOrZero 函数测试
	val, ok := GetOrZero(intPtr)
	assert.Equal(t, 0, val)
	assert.False(t, ok)

	valStr, ok := GetOrZero(stringPtr)
	assert.Equal(t, "", valStr)
	assert.False(t, ok)

	valBool, ok := GetOrZero(boolPtr)
	assert.Equal(t, false, valBool)
	assert.False(t, ok)
}

func TestEdgeCases(t *testing.T) {
	// 测试零值
	zeroInt := 0
	zeroString := ""
	zeroBool := false

	ptr1 := Of(zeroInt)
	assert.Equal(t, 0, *ptr1)
	assert.Equal(t, 0, Get(ptr1))
	assert.Equal(t, 0, GetOr(ptr1, 100))
	val, ok := GetOrZero(ptr1)
	assert.Equal(t, 0, val)
	assert.True(t, ok)

	ptr2 := Of(zeroString)
	assert.Equal(t, "", *ptr2)
	assert.Equal(t, "", Get(ptr2))
	assert.Equal(t, "", GetOr(ptr2, "default"))
	valStr, ok := GetOrZero(ptr2)
	assert.Equal(t, "", valStr)
	assert.True(t, ok)

	ptr3 := Of(zeroBool)
	assert.Equal(t, false, *ptr3)
	assert.Equal(t, false, Get(ptr3))
	assert.Equal(t, false, GetOr(ptr3, true))
	valBool, ok := GetOrZero(ptr3)
	assert.Equal(t, false, valBool)
	assert.True(t, ok)
}

// 性能基准测试

func BenchmarkOf_Int(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Of(42)
	}
}

func BenchmarkOf_String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Of("hello world")
	}
}

func BenchmarkOf_Bool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Of(true)
	}
}

func BenchmarkGet_ValidInt(b *testing.B) {
	ptr := Of(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(ptr)
	}
}

func BenchmarkGet_NilInt(b *testing.B) {
	var ptr *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(ptr)
	}
}

func BenchmarkGetOr_ValidInt(b *testing.B) {
	ptr := Of(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetOr(ptr, 100)
	}
}

func BenchmarkGetOr_NilInt(b *testing.B) {
	var ptr *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetOr(ptr, 100)
	}
}

func BenchmarkGetOrZero_ValidInt(b *testing.B) {
	ptr := Of(42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetOrZero(ptr)
	}
}

func BenchmarkGetOrZero_NilInt(b *testing.B) {
	var ptr *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetOrZero(ptr)
	}
}

func BenchmarkGet_ValidString(b *testing.B) {
	ptr := Of("hello world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(ptr)
	}
}

func BenchmarkGet_NilString(b *testing.B) {
	var ptr *string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(ptr)
	}
}

func BenchmarkGetOr_ValidString(b *testing.B) {
	ptr := Of("hello world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetOr(ptr, "default")
	}
}

func BenchmarkGetOr_NilString(b *testing.B) {
	var ptr *string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetOr(ptr, "default")
	}
}

func BenchmarkGetOrZero_ValidString(b *testing.B) {
	ptr := Of("hello world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetOrZero(ptr)
	}
}

func BenchmarkGetOrZero_NilString(b *testing.B) {
	var ptr *string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetOrZero(ptr)
	}
}

// 复杂类型基准测试

func BenchmarkOf_Struct(b *testing.B) {
	type TestStruct struct {
		ID   int
		Name string
		Data []int
	}

	obj := TestStruct{
		ID:   1,
		Name: "test",
		Data: []int{1, 2, 3, 4, 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Of(obj)
	}
}

func BenchmarkGet_ValidStruct(b *testing.B) {
	type TestStruct struct {
		ID   int
		Name string
		Data []int
	}

	obj := TestStruct{
		ID:   1,
		Name: "test",
		Data: []int{1, 2, 3, 4, 5},
	}
	ptr := Of(obj)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(ptr)
	}
}

func BenchmarkGet_NilStruct(b *testing.B) {
	type TestStruct struct {
		ID   int
		Name string
		Data []int
	}

	var ptr *TestStruct

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Get(ptr)
	}
}
