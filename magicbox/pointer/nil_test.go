package pointer_test

import (
	"testing"

	"github.com/aide-family/magicbox/pointer"
)

// TestIsNil_Nil 测试 nil 值
func TestIsNil_Nil(t *testing.T) {
	if !pointer.IsNil(nil) {
		t.Error("IsNil(nil) should return true")
	}
}

// TestIsNil_PointerToNil 测试指向 nil 的指针
func TestIsNil_PointerToNil(t *testing.T) {
	var p *int
	if !pointer.IsNil(p) {
		t.Error("IsNil(*int(nil)) should return true")
	}
}

// TestIsNil_PointerToValue 测试指向值的指针
func TestIsNil_PointerToValue(t *testing.T) {
	x := 42
	p := &x
	if pointer.IsNil(p) {
		t.Error("IsNil(&x) should return false")
	}
}

// TestIsNil_DoublePointerToNil 测试指向 nil 的双重指针
func TestIsNil_DoublePointerToNil(t *testing.T) {
	var p *int
	var pp **int = &p
	if !pointer.IsNil(pp) {
		t.Error("IsNil(**int pointing to nil) should return true")
	}
}

// TestIsNil_DoublePointerToValue 测试指向值的双重指针
func TestIsNil_DoublePointerToValue(t *testing.T) {
	x := 42
	p := &x
	pp := &p
	if pointer.IsNil(pp) {
		t.Error("IsNil(**int pointing to value) should return false")
	}
}

// TestIsNil_TriplePointer 测试三重指针
func TestIsNil_TriplePointer(t *testing.T) {
	var p *int
	var pp **int = &p
	var ppp ***int = &pp
	if !pointer.IsNil(ppp) {
		t.Error("IsNil(***int pointing to nil) should return true")
	}

	x := 42
	p2 := &x
	pp2 := &p2
	ppp2 := &pp2
	if pointer.IsNil(ppp2) {
		t.Error("IsNil(***int pointing to value) should return false")
	}
}

// TestIsNil_Slice 测试 slice
func TestIsNil_Slice(t *testing.T) {
	var s []int
	if !pointer.IsNil(s) {
		t.Error("IsNil([]int(nil)) should return true")
	}

	s2 := []int{1, 2, 3}
	if pointer.IsNil(s2) {
		t.Error("IsNil([]int{1,2,3}) should return false")
	}
}

// TestIsNil_Map 测试 map
func TestIsNil_Map(t *testing.T) {
	var m map[string]int
	if !pointer.IsNil(m) {
		t.Error("IsNil(map[string]int(nil)) should return true")
	}

	m2 := make(map[string]int)
	if pointer.IsNil(m2) {
		t.Error("IsNil(make(map[string]int)) should return false")
	}
}

// TestIsNil_Chan 测试 channel
func TestIsNil_Chan(t *testing.T) {
	var ch chan int
	if !pointer.IsNil(ch) {
		t.Error("IsNil(chan int(nil)) should return true")
	}

	ch2 := make(chan int)
	if pointer.IsNil(ch2) {
		t.Error("IsNil(make(chan int)) should return false")
	}
}

// TestIsNil_Func 测试函数
func TestIsNil_Func(t *testing.T) {
	var f func()
	if !pointer.IsNil(f) {
		t.Error("IsNil(func()(nil)) should return true")
	}

	f2 := func() {}
	if pointer.IsNil(f2) {
		t.Error("IsNil(func(){}) should return false")
	}
}

// TestIsNil_Interface 测试接口
func TestIsNil_Interface(t *testing.T) {
	var i interface{}
	if !pointer.IsNil(i) {
		t.Error("IsNil(interface{}(nil)) should return true")
	}

	var p *int
	var i2 interface{} = p
	if !pointer.IsNil(i2) {
		t.Error("IsNil(interface{} containing nil pointer) should return true")
	}

	x := 42
	var i3 interface{} = &x
	if pointer.IsNil(i3) {
		t.Error("IsNil(interface{} containing non-nil pointer) should return false")
	}
}

// TestIsNil_ValueTypes 测试值类型
func TestIsNil_ValueTypes(t *testing.T) {
	// 值类型不应该被认为是 nil
	if pointer.IsNil(0) {
		t.Error("IsNil(0) should return false")
	}

	if pointer.IsNil("") {
		t.Error("IsNil(\"\") should return false")
	}

	if pointer.IsNil(false) {
		t.Error("IsNil(false) should return false")
	}

	if pointer.IsNil(struct{}{}) {
		t.Error("IsNil(struct{}{}) should return false")
	}
}

// TestIsNil_PointerToInterface 测试指向接口的指针
func TestIsNil_PointerToInterface(t *testing.T) {
	var i interface{}
	var pi *interface{} = &i
	if !pointer.IsNil(pi) {
		t.Error("IsNil(*interface{} pointing to nil interface) should return true")
	}

	x := 42
	var i2 interface{} = x
	var pi2 *interface{} = &i2
	if pointer.IsNil(pi2) {
		t.Error("IsNil(*interface{} pointing to non-nil interface) should return false")
	}
}

// TestIsNil_ComplexTypes 测试复杂类型
func TestIsNil_ComplexTypes(t *testing.T) {
	// 结构体
	type S struct {
		X int
	}
	var s *S
	if !pointer.IsNil(s) {
		t.Error("IsNil(*S(nil)) should return true")
	}

	s2 := &S{X: 1}
	if pointer.IsNil(s2) {
		t.Error("IsNil(&S{X:1}) should return false")
	}

	// 数组（不是指针）
	arr := [3]int{1, 2, 3}
	if pointer.IsNil(arr) {
		t.Error("IsNil([3]int{1,2,3}) should return false")
	}
}

// TestIsNotNil 测试 IsNotNil 函数
func TestIsNotNil(t *testing.T) {
	if pointer.IsNotNil(nil) {
		t.Error("IsNotNil(nil) should return false")
	}

	var p *int
	if pointer.IsNotNil(p) {
		t.Error("IsNotNil(*int(nil)) should return false")
	}

	x := 42
	if !pointer.IsNotNil(&x) {
		t.Error("IsNotNil(&x) should return true")
	}

	s := []int{1, 2, 3}
	if !pointer.IsNotNil(s) {
		t.Error("IsNotNil([]int{1,2,3}) should return true")
	}
}

// TestIsNil_EdgeCases 测试边界情况
func TestIsNil_EdgeCases(t *testing.T) {
	// 空字符串
	if pointer.IsNil("") {
		t.Error("IsNil(\"\") should return false")
	}

	// 零值
	if pointer.IsNil(0) {
		t.Error("IsNil(0) should return false")
	}

	// 空结构体
	if pointer.IsNil(struct{}{}) {
		t.Error("IsNil(struct{}{}) should return false")
	}

	// 空数组
	var arr [0]int
	if pointer.IsNil(arr) {
		t.Error("IsNil([0]int{}) should return false")
	}
}

// TestIsNil_PointerChain 测试指针链
func TestIsNil_PointerChain(t *testing.T) {
	// 测试多层指针链，确保不会无限循环
	var p1 *int
	var p2 **int = &p1
	var p3 ***int = &p2
	var p4 ****int = &p3

	if !pointer.IsNil(p4) {
		t.Error("IsNil(****int pointing to nil) should return true")
	}

	x := 42
	p1v := &x
	p2v := &p1v
	p3v := &p2v
	p4v := &p3v

	if pointer.IsNil(p4v) {
		t.Error("IsNil(****int pointing to value) should return false")
	}
}

// TestIsNil_InterfaceWithPointer 测试包含指针的接口
func TestIsNil_InterfaceWithPointer(t *testing.T) {
	var p *int
	var i interface{} = p
	if !pointer.IsNil(i) {
		t.Error("IsNil(interface{} with nil pointer) should return true")
	}

	x := 42
	var i2 interface{} = &x
	if pointer.IsNil(i2) {
		t.Error("IsNil(interface{} with non-nil pointer) should return false")
	}
}

// BenchmarkIsNil_Nil 基准测试 IsNil(nil)
func BenchmarkIsNil_Nil(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pointer.IsNil(nil)
	}
}

// BenchmarkIsNil_Pointer 基准测试 IsNil(pointer)
func BenchmarkIsNil_Pointer(b *testing.B) {
	var p *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pointer.IsNil(p)
	}
}

// BenchmarkIsNil_Value 基准测试 IsNil(value)
func BenchmarkIsNil_Value(b *testing.B) {
	x := 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pointer.IsNil(x)
	}
}

// BenchmarkIsNil_Slice 基准测试 IsNil(slice)
func BenchmarkIsNil_Slice(b *testing.B) {
	s := []int{1, 2, 3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pointer.IsNil(s)
	}
}

// BenchmarkIsNotNil 基准测试 IsNotNil
func BenchmarkIsNotNil(b *testing.B) {
	x := 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pointer.IsNotNil(&x)
	}
}

// TestIsNil_InterfaceNilValue 测试接口包含 nil 值的情况
func TestIsNil_InterfaceNilValue(t *testing.T) {
	// 接口包含 nil 指针
	var p *int
	var i interface{} = p
	if !pointer.IsNil(i) {
		t.Error("IsNil(interface{} with nil pointer) should return true")
	}

	// 接口包含 nil slice
	var s []int
	var i2 interface{} = s
	if !pointer.IsNil(i2) {
		t.Error("IsNil(interface{} with nil slice) should return true")
	}

	// 接口包含 nil map
	var m map[string]int
	var i3 interface{} = m
	if !pointer.IsNil(i3) {
		t.Error("IsNil(interface{} with nil map) should return true")
	}

	// 接口包含 nil channel
	var ch chan int
	var i4 interface{} = ch
	if !pointer.IsNil(i4) {
		t.Error("IsNil(interface{} with nil channel) should return true")
	}

	// 接口包含 nil function
	var f func()
	var i5 interface{} = f
	if !pointer.IsNil(i5) {
		t.Error("IsNil(interface{} with nil function) should return true")
	}
}

// TestIsNil_DeepPointerChain 测试深层指针链，确保不会无限循环
func TestIsNil_DeepPointerChain(t *testing.T) {
	// 创建 10 层指针链
	var p *int
	var pp1 **int = &p
	var pp2 ***int = &pp1
	var pp3 ****int = &pp2
	var pp4 *****int = &pp3
	var pp5 ******int = &pp4
	var pp6 *******int = &pp5
	var pp7 ********int = &pp6
	var pp8 *********int = &pp7
	var pp9 **********int = &pp8
	var pp10 ***********int = &pp9

	if !pointer.IsNil(pp10) {
		t.Error("IsNil with 10-level pointer chain to nil should return true")
	}

	// 创建指向值的深层指针链
	x := 42
	pv := &x
	pp1v := &pv
	pp2v := &pp1v
	pp3v := &pp2v
	pp4v := &pp3v
	pp5v := &pp4v

	if pointer.IsNil(pp5v) {
		t.Error("IsNil with 5-level pointer chain to value should return false")
	}
}

// TestIsNil_InterfaceWithInterface 测试接口包含接口的情况
func TestIsNil_InterfaceWithInterface(t *testing.T) {
	var i1 interface{}
	var i2 interface{} = i1
	if !pointer.IsNil(i2) {
		t.Error("IsNil(interface{} containing nil interface{}) should return true")
	}

	var p *int
	var i3 interface{} = p
	var i4 interface{} = i3
	if !pointer.IsNil(i4) {
		t.Error("IsNil(interface{} containing interface{} with nil pointer) should return true")
	}
}
