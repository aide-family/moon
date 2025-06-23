package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNil(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "nil interface",
			input:    nil,
			expected: true,
		},
		{
			name:     "nil pointer to int",
			input:    (*int)(nil),
			expected: true,
		},
		{
			name:     "nil pointer to string",
			input:    (*string)(nil),
			expected: true,
		},
		{
			name:     "nil pointer to struct",
			input:    (*struct{})(nil),
			expected: true,
		},
		{
			name:     "nil slice",
			input:    []int(nil),
			expected: true,
		},
		{
			name:     "nil map",
			input:    map[string]int(nil),
			expected: true,
		},
		{
			name:     "nil channel",
			input:    (chan int)(nil),
			expected: true,
		},
		{
			name:     "nil function",
			input:    (func())(nil),
			expected: true,
		},
		{
			name:     "nil interface{}",
			input:    (interface{})(nil),
			expected: true,
		},
		{
			name:     "valid int",
			input:    42,
			expected: false,
		},
		{
			name:     "valid string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "valid bool",
			input:    true,
			expected: false,
		},
		{
			name:     "valid float",
			input:    3.14,
			expected: false,
		},
		{
			name:     "zero int",
			input:    0,
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "false bool",
			input:    false,
			expected: false,
		},
		{
			name:     "valid pointer to int",
			input:    func() *int { v := 42; return &v }(),
			expected: false,
		},
		{
			name:     "valid pointer to string",
			input:    func() *string { v := "hello"; return &v }(),
			expected: false,
		},
		{
			name:     "valid slice",
			input:    []int{1, 2, 3},
			expected: false,
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: false,
		},
		{
			name:     "valid map",
			input:    map[string]int{"a": 1},
			expected: false,
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: false,
		},
		{
			name:     "valid channel",
			input:    make(chan int),
			expected: false,
		},
		{
			name:     "valid function",
			input:    func() {},
			expected: false,
		},
		{
			name:     "valid struct",
			input:    struct{ Name string }{Name: "test"},
			expected: false,
		},
		{
			name:     "empty struct",
			input:    struct{}{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNil(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %v", tt.input)
		})
	}
}

func TestIsNotNil(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "nil interface",
			input:    nil,
			expected: false,
		},
		{
			name:     "nil pointer to int",
			input:    (*int)(nil),
			expected: false,
		},
		{
			name:     "nil pointer to string",
			input:    (*string)(nil),
			expected: false,
		},
		{
			name:     "nil slice",
			input:    []int(nil),
			expected: false,
		},
		{
			name:     "nil map",
			input:    map[string]int(nil),
			expected: false,
		},
		{
			name:     "nil channel",
			input:    (chan int)(nil),
			expected: false,
		},
		{
			name:     "nil function",
			input:    (func())(nil),
			expected: false,
		},
		{
			name:     "valid int",
			input:    42,
			expected: true,
		},
		{
			name:     "valid string",
			input:    "hello",
			expected: true,
		},
		{
			name:     "valid bool",
			input:    true,
			expected: true,
		},
		{
			name:     "zero int",
			input:    0,
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "false bool",
			input:    false,
			expected: true,
		},
		{
			name:     "valid pointer to int",
			input:    func() *int { v := 42; return &v }(),
			expected: true,
		},
		{
			name:     "valid slice",
			input:    []int{1, 2, 3},
			expected: true,
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: true,
		},
		{
			name:     "valid map",
			input:    map[string]int{"a": 1},
			expected: true,
		},
		{
			name:     "empty map",
			input:    map[string]int{},
			expected: true,
		},
		{
			name:     "valid channel",
			input:    make(chan int),
			expected: true,
		},
		{
			name:     "valid function",
			input:    func() {},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotNil(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %v", tt.input)
		})
	}
}

func TestIsNilAndIsNotNilRelationship(t *testing.T) {
	// 测试 IsNil 和 IsNotNil 的互补关系
	testCases := []interface{}{
		nil,
		(*int)(nil),
		(*string)(nil),
		[]int(nil),
		map[string]int(nil),
		(chan int)(nil),
		(func())(nil),
		42,
		"hello",
		true,
		false,
		0,
		"",
		[]int{1, 2, 3},
		[]int{},
		map[string]int{"a": 1},
		map[string]int{},
		make(chan int),
		func() {},
		struct{}{},
	}

	for _, tc := range testCases {
		t.Run("complementary_test", func(t *testing.T) {
			isNil := IsNil(tc)
			isNotNil := IsNotNil(tc)

			// IsNil 和 IsNotNil 应该是互补的
			assert.Equal(t, !isNil, isNotNil, "Input: %v, IsNil: %v, IsNotNil: %v", tc, isNil, isNotNil)
		})
	}
}

func TestComplexTypes(t *testing.T) {
	// 测试复杂类型
	type Person struct {
		Name string
		Age  int
	}

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "nil struct pointer",
			input:    (*Person)(nil),
			expected: true,
		},
		{
			name:     "valid struct pointer",
			input:    &Person{Name: "Alice", Age: 30},
			expected: false,
		},
		{
			name:     "struct value",
			input:    Person{Name: "Bob", Age: 25},
			expected: false,
		},
		{
			name:     "nil interface slice",
			input:    []interface{}(nil),
			expected: true,
		},
		{
			name:     "interface slice with nil",
			input:    []interface{}{nil, "hello", 42},
			expected: false,
		},
		{
			name:     "nil map with interface",
			input:    map[string]interface{}(nil),
			expected: true,
		},
		{
			name:     "map with nil values",
			input:    map[string]interface{}{"nil": nil, "str": "hello"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNil(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %v", tt.input)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	// 测试边界情况
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "nil interface{}",
			input:    (interface{})(nil),
			expected: true,
		},
		{
			name:     "interface{} with nil",
			input:    interface{}(nil),
			expected: true,
		},
		{
			name:     "interface{} with value",
			input:    interface{}(42),
			expected: false,
		},
		{
			name:     "nil pointer to interface",
			input:    (*interface{})(nil),
			expected: true,
		},
		{
			name:     "pointer to interface with nil",
			input:    func() *interface{} { var v interface{} = nil; return &v }(),
			expected: true,
		},
		{
			name:     "pointer to interface with value",
			input:    func() *interface{} { var v interface{} = 42; return &v }(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNil(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %v", tt.input)
		})
	}
}

// 性能基准测试

func BenchmarkIsNil_Nil(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = IsNil(nil)
	}
}

func BenchmarkIsNil_ValidInt(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = IsNil(42)
	}
}

func BenchmarkIsNil_NilPointer(t *testing.B) {
	var ptr *int
	for i := 0; i < t.N; i++ {
		_ = IsNil(ptr)
	}
}

func BenchmarkIsNil_ValidPointer(t *testing.B) {
	val := 42
	ptr := &val
	for i := 0; i < t.N; i++ {
		_ = IsNil(ptr)
	}
}

func BenchmarkIsNil_NilSlice(t *testing.B) {
	var slice []int
	for i := 0; i < t.N; i++ {
		_ = IsNil(slice)
	}
}

func BenchmarkIsNil_ValidSlice(t *testing.B) {
	slice := []int{1, 2, 3}
	for i := 0; i < t.N; i++ {
		_ = IsNil(slice)
	}
}

func BenchmarkIsNil_NilMap(t *testing.B) {
	var m map[string]int
	for i := 0; i < t.N; i++ {
		_ = IsNil(m)
	}
}

func BenchmarkIsNil_ValidMap(t *testing.B) {
	m := map[string]int{"a": 1, "b": 2}
	for i := 0; i < t.N; i++ {
		_ = IsNil(m)
	}
}

func BenchmarkIsNil_NilChannel(t *testing.B) {
	var ch chan int
	for i := 0; i < t.N; i++ {
		_ = IsNil(ch)
	}
}

func BenchmarkIsNil_ValidChannel(t *testing.B) {
	ch := make(chan int)
	for i := 0; i < t.N; i++ {
		_ = IsNil(ch)
	}
}

func BenchmarkIsNil_NilFunction(t *testing.B) {
	var fn func()
	for i := 0; i < t.N; i++ {
		_ = IsNil(fn)
	}
}

func BenchmarkIsNil_ValidFunction(t *testing.B) {
	fn := func() {}
	for i := 0; i < t.N; i++ {
		_ = IsNil(fn)
	}
}

func BenchmarkIsNotNil_Nil(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = IsNotNil(nil)
	}
}

func BenchmarkIsNotNil_ValidInt(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = IsNotNil(42)
	}
}

func BenchmarkIsNotNil_NilPointer(t *testing.B) {
	var ptr *int
	for i := 0; i < t.N; i++ {
		_ = IsNotNil(ptr)
	}
}

func BenchmarkIsNotNil_ValidPointer(t *testing.B) {
	val := 42
	ptr := &val
	for i := 0; i < t.N; i++ {
		_ = IsNotNil(ptr)
	}
}

// 复杂类型基准测试

func BenchmarkIsNil_ComplexStruct(t *testing.B) {
	type ComplexStruct struct {
		ID   int
		Name string
		Data []int
		Meta map[string]interface{}
	}

	var ptr *ComplexStruct
	for i := 0; i < t.N; i++ {
		_ = IsNil(ptr)
	}
}

func BenchmarkIsNil_ValidComplexStruct(t *testing.B) {
	type ComplexStruct struct {
		ID   int
		Name string
		Data []int
		Meta map[string]interface{}
	}

	obj := &ComplexStruct{
		ID:   1,
		Name: "test",
		Data: []int{1, 2, 3, 4, 5},
		Meta: map[string]interface{}{"key": "value"},
	}

	for i := 0; i < t.N; i++ {
		_ = IsNil(obj)
	}
}

func BenchmarkIsNil_InterfaceSlice(t *testing.B) {
	var slice []interface{}
	for i := 0; i < t.N; i++ {
		_ = IsNil(slice)
	}
}

func BenchmarkIsNil_ValidInterfaceSlice(t *testing.B) {
	slice := []interface{}{"hello", 42, true, nil}
	for i := 0; i < t.N; i++ {
		_ = IsNil(slice)
	}
}
