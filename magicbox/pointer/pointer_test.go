package pointer_test

import (
	"testing"

	"github.com/aide-family/magicbox/pointer"
)

// TestOf 测试 Of 函数
func TestOf(t *testing.T) {
	// 测试 int 类型
	x := 42
	p := pointer.Of(x)
	if p == nil {
		t.Fatal("Of(42) returned nil pointer")
	}
	if *p != 42 {
		t.Errorf("Of(42) = %d, want 42", *p)
	}

	// 测试 string 类型
	s := "hello"
	ps := pointer.Of(s)
	if ps == nil {
		t.Fatal("Of(\"hello\") returned nil pointer")
	}
	if *ps != "hello" {
		t.Errorf("Of(\"hello\") = %q, want \"hello\"", *ps)
	}

	// 测试 bool 类型
	b := true
	pb := pointer.Of(b)
	if pb == nil {
		t.Fatal("Of(true) returned nil pointer")
	}
	if *pb != true {
		t.Errorf("Of(true) = %v, want true", *pb)
	}

	// 测试 float 类型
	f := 3.14
	pf := pointer.Of(f)
	if pf == nil {
		t.Fatal("Of(3.14) returned nil pointer")
	}
	if *pf != 3.14 {
		t.Errorf("Of(3.14) = %v, want 3.14", *pf)
	}

	// 测试结构体类型
	type S struct {
		X int
		Y string
	}
	ss := S{X: 1, Y: "test"}
	ps2 := pointer.Of(ss)
	if ps2 == nil {
		t.Fatal("Of(S{}) returned nil pointer")
	}
	if ps2.X != 1 || ps2.Y != "test" {
		t.Errorf("Of(S{X:1, Y:\"test\"}) = %+v, want {X:1, Y:\"test\"}", *ps2)
	}

	// 测试零值
	var zero int
	pz := pointer.Of(zero)
	if pz == nil {
		t.Fatal("Of(0) returned nil pointer")
	}
	if *pz != 0 {
		t.Errorf("Of(0) = %d, want 0", *pz)
	}
}

// TestGet 测试 Get 函数
func TestGet(t *testing.T) {
	// 测试非 nil 指针
	x := 42
	p := &x
	if got := pointer.Get(p); got != 42 {
		t.Errorf("Get(&42) = %d, want 42", got)
	}

	// 测试 nil 指针（应该返回零值）
	var p2 *int
	if got := pointer.Get(p2); got != 0 {
		t.Errorf("Get(nil) = %d, want 0", got)
	}

	// 测试 string 类型
	s := "hello"
	ps := &s
	if got := pointer.Get(ps); got != "hello" {
		t.Errorf("Get(&\"hello\") = %q, want \"hello\"", got)
	}

	var ps2 *string
	if got := pointer.Get(ps2); got != "" {
		t.Errorf("Get(nil *string) = %q, want \"\"", got)
	}

	// 测试 bool 类型
	b := true
	pb := &b
	if got := pointer.Get(pb); got != true {
		t.Errorf("Get(&true) = %v, want true", got)
	}

	var pb2 *bool
	if got := pointer.Get(pb2); got != false {
		t.Errorf("Get(nil *bool) = %v, want false", got)
	}

	// 测试结构体类型
	type S struct {
		X int
		Y string
	}
	ss := S{X: 1, Y: "test"}
	ps3 := &ss
	got := pointer.Get(ps3)
	if got.X != 1 || got.Y != "test" {
		t.Errorf("Get(&S{X:1, Y:\"test\"}) = %+v, want {X:1, Y:\"test\"}", got)
	}

	var ps4 *S
	got2 := pointer.Get(ps4)
	if got2.X != 0 || got2.Y != "" {
		t.Errorf("Get(nil *S) = %+v, want {X:0, Y:\"\"}", got2)
	}
}

// TestGetOr 测试 GetOr 函数
func TestGetOr(t *testing.T) {
	// 测试非 nil 指针
	x := 42
	p := &x
	if got := pointer.GetOr(p, 100); got != 42 {
		t.Errorf("GetOr(&42, 100) = %d, want 42", got)
	}

	// 测试 nil 指针（应该返回默认值）
	var p2 *int
	if got := pointer.GetOr(p2, 100); got != 100 {
		t.Errorf("GetOr(nil, 100) = %d, want 100", got)
	}

	// 测试 string 类型
	s := "hello"
	ps := &s
	if got := pointer.GetOr(ps, "default"); got != "hello" {
		t.Errorf("GetOr(&\"hello\", \"default\") = %q, want \"hello\"", got)
	}

	var ps2 *string
	if got := pointer.GetOr(ps2, "default"); got != "default" {
		t.Errorf("GetOr(nil, \"default\") = %q, want \"default\"", got)
	}

	// 测试 bool 类型
	b := true
	pb := &b
	if got := pointer.GetOr(pb, false); got != true {
		t.Errorf("GetOr(&true, false) = %v, want true", got)
	}

	var pb2 *bool
	if got := pointer.GetOr(pb2, true); got != true {
		t.Errorf("GetOr(nil, true) = %v, want true", got)
	}

	// 测试结构体类型
	type S struct {
		X int
		Y string
	}
	ss := S{X: 1, Y: "test"}
	ps3 := &ss
	def := S{X: 0, Y: "default"}
	got := pointer.GetOr(ps3, def)
	if got.X != 1 || got.Y != "test" {
		t.Errorf("GetOr(&S{X:1, Y:\"test\"}, def) = %+v, want {X:1, Y:\"test\"}", got)
	}

	var ps4 *S
	got2 := pointer.GetOr(ps4, def)
	if got2.X != 0 || got2.Y != "default" {
		t.Errorf("GetOr(nil, def) = %+v, want {X:0, Y:\"default\"}", got2)
	}
}

// TestGetOrZero 测试 GetOrZero 函数
func TestGetOrZero(t *testing.T) {
	// 测试非 nil 指针
	x := 42
	p := &x
	got, ok := pointer.GetOrZero(p)
	if !ok {
		t.Error("GetOrZero(&42) should return ok=true")
	}
	if got != 42 {
		t.Errorf("GetOrZero(&42) = (%d, %v), want (42, true)", got, ok)
	}

	// 测试 nil 指针（应该返回零值和 false）
	var p2 *int
	got2, ok2 := pointer.GetOrZero(p2)
	if ok2 {
		t.Error("GetOrZero(nil) should return ok=false")
	}
	if got2 != 0 {
		t.Errorf("GetOrZero(nil) = (%d, %v), want (0, false)", got2, ok2)
	}

	// 测试 string 类型
	s := "hello"
	ps := &s
	got3, ok3 := pointer.GetOrZero(ps)
	if !ok3 {
		t.Error("GetOrZero(&\"hello\") should return ok=true")
	}
	if got3 != "hello" {
		t.Errorf("GetOrZero(&\"hello\") = (%q, %v), want (\"hello\", true)", got3, ok3)
	}

	var ps2 *string
	got4, ok4 := pointer.GetOrZero(ps2)
	if ok4 {
		t.Error("GetOrZero(nil *string) should return ok=false")
	}
	if got4 != "" {
		t.Errorf("GetOrZero(nil *string) = (%q, %v), want (\"\", false)", got4, ok4)
	}

	// 测试 bool 类型
	b := true
	pb := &b
	got5, ok5 := pointer.GetOrZero(pb)
	if !ok5 {
		t.Error("GetOrZero(&true) should return ok=true")
	}
	if got5 != true {
		t.Errorf("GetOrZero(&true) = (%v, %v), want (true, true)", got5, ok5)
	}

	var pb2 *bool
	got6, ok6 := pointer.GetOrZero(pb2)
	if ok6 {
		t.Error("GetOrZero(nil *bool) should return ok=false")
	}
	if got6 != false {
		t.Errorf("GetOrZero(nil *bool) = (%v, %v), want (false, false)", got6, ok6)
	}

	// 测试结构体类型
	type S struct {
		X int
		Y string
	}
	ss := S{X: 1, Y: "test"}
	ps3 := &ss
	got7, ok7 := pointer.GetOrZero(ps3)
	if !ok7 {
		t.Error("GetOrZero(&S{}) should return ok=true")
	}
	if got7.X != 1 || got7.Y != "test" {
		t.Errorf("GetOrZero(&S{X:1, Y:\"test\"}) = (%+v, %v), want ({X:1, Y:\"test\"}, true)", got7, ok7)
	}

	var ps4 *S
	got8, ok8 := pointer.GetOrZero(ps4)
	if ok8 {
		t.Error("GetOrZero(nil *S) should return ok=false")
	}
	if got8.X != 0 || got8.Y != "" {
		t.Errorf("GetOrZero(nil *S) = (%+v, %v), want ({X:0, Y:\"\"}, false)", got8, ok8)
	}
}

// TestOf_ZeroValue 测试 Of 函数处理零值
func TestOf_ZeroValue(t *testing.T) {
	// 测试各种类型的零值
	var zeroInt int
	p1 := pointer.Of(zeroInt)
	if *p1 != 0 {
		t.Errorf("Of(0) = %d, want 0", *p1)
	}

	var zeroString string
	p2 := pointer.Of(zeroString)
	if *p2 != "" {
		t.Errorf("Of(\"\") = %q, want \"\"", *p2)
	}

	var zeroBool bool
	p3 := pointer.Of(zeroBool)
	if *p3 != false {
		t.Errorf("Of(false) = %v, want false", *p3)
	}

	var zeroFloat float64
	p4 := pointer.Of(zeroFloat)
	if *p4 != 0.0 {
		t.Errorf("Of(0.0) = %v, want 0.0", *p4)
	}
}

// TestGet_Slice 测试 Get 函数处理 slice 类型
func TestGet_Slice(t *testing.T) {
	s := []int{1, 2, 3}
	ps := &s
	got := pointer.Get(ps)
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("Get(&[]int{1,2,3}) = %v, want [1,2,3]", got)
	}

	var ps2 *[]int
	got2 := pointer.Get(ps2)
	if got2 != nil {
		t.Errorf("Get(nil *[]int) = %v, want nil", got2)
	}
}

// TestGet_Map 测试 Get 函数处理 map 类型
func TestGet_Map(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	pm := &m
	got := pointer.Get(pm)
	if got["a"] != 1 || got["b"] != 2 {
		t.Errorf("Get(&map) = %v, want map[a:1 b:2]", got)
	}

	var pm2 *map[string]int
	got2 := pointer.Get(pm2)
	if got2 != nil {
		t.Errorf("Get(nil *map) = %v, want nil", got2)
	}
}

// TestGet_Channel 测试 Get 函数处理 channel 类型
func TestGet_Channel(t *testing.T) {
	ch := make(chan int, 1)
	pch := &ch
	got := pointer.Get(pch)
	if got == nil {
		t.Error("Get(&chan) should not return nil")
	}

	var pch2 *chan int
	got2 := pointer.Get(pch2)
	if got2 != nil {
		t.Errorf("Get(nil *chan) = %v, want nil", got2)
	}
}

// TestGet_Function 测试 Get 函数处理 function 类型
func TestGet_Function(t *testing.T) {
	f := func() int { return 42 }
	pf := &f
	got := pointer.Get(pf)
	if got() != 42 {
		t.Error("Get(&func) should return the function")
	}

	var pf2 *func() int
	got2 := pointer.Get(pf2)
	if got2 != nil {
		t.Error("Get(nil *func) should return nil")
	}
}

// TestGetOr_ZeroDefault 测试 GetOr 函数使用零值作为默认值
func TestGetOr_ZeroDefault(t *testing.T) {
	var p *int
	if got := pointer.GetOr(p, 0); got != 0 {
		t.Errorf("GetOr(nil, 0) = %d, want 0", got)
	}

	var ps *string
	if got := pointer.GetOr(ps, ""); got != "" {
		t.Errorf("GetOr(nil, \"\") = %q, want \"\"", got)
	}

	var pb *bool
	if got := pointer.GetOr(pb, false); got != false {
		t.Errorf("GetOr(nil, false) = %v, want false", got)
	}
}

// TestGetOrZero_AllTypes 测试 GetOrZero 函数处理所有类型
func TestGetOrZero_AllTypes(t *testing.T) {
	// int
	x := 42
	px := &x
	v1, ok1 := pointer.GetOrZero(px)
	if !ok1 || v1 != 42 {
		t.Errorf("GetOrZero(&int) = (%d, %v), want (42, true)", v1, ok1)
	}

	var px2 *int
	v2, ok2 := pointer.GetOrZero(px2)
	if ok2 || v2 != 0 {
		t.Errorf("GetOrZero(nil *int) = (%d, %v), want (0, false)", v2, ok2)
	}

	// float64
	f := 3.14
	pf := &f
	v3, ok3 := pointer.GetOrZero(pf)
	if !ok3 || v3 != 3.14 {
		t.Errorf("GetOrZero(&float64) = (%v, %v), want (3.14, true)", v3, ok3)
	}

	var pf2 *float64
	v4, ok4 := pointer.GetOrZero(pf2)
	if ok4 || v4 != 0.0 {
		t.Errorf("GetOrZero(nil *float64) = (%v, %v), want (0.0, false)", v4, ok4)
	}
}

// TestOf_Modification 测试 Of 函数返回的指针是独立的
func TestOf_Modification(t *testing.T) {
	x := 42
	p := pointer.Of(x)
	*p = 100
	// 修改指针指向的值不应该影响原始值
	if x != 42 {
		t.Errorf("Modifying pointer should not affect original value, x = %d, want 42", x)
	}
	if *p != 100 {
		t.Errorf("Pointer value should be modified, *p = %d, want 100", *p)
	}
}

// TestGet_Modification 测试 Get 函数返回的值是副本
func TestGet_Modification(t *testing.T) {
	x := 42
	p := &x
	got := pointer.Get(p)
	got = 100
	// 修改返回值不应该影响原始值
	if x != 42 {
		t.Errorf("Modifying returned value should not affect original, x = %d, want 42", x)
	}
	if got != 100 {
		t.Errorf("Returned value should be modified, got = %d, want 100", got)
	}
}

// BenchmarkOf 基准测试 Of 函数
func BenchmarkOf(b *testing.B) {
	x := 42
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pointer.Of(x)
	}
}

// BenchmarkGet 基准测试 Get 函数
func BenchmarkGet(b *testing.B) {
	x := 42
	p := &x
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pointer.Get(p)
	}
}

// BenchmarkGet_Nil 基准测试 Get 函数（nil 指针）
func BenchmarkGet_Nil(b *testing.B) {
	var p *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pointer.Get(p)
	}
}

// BenchmarkGetOr 基准测试 GetOr 函数
func BenchmarkGetOr(b *testing.B) {
	x := 42
	p := &x
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pointer.GetOr(p, 100)
	}
}

// BenchmarkGetOr_Nil 基准测试 GetOr 函数（nil 指针）
func BenchmarkGetOr_Nil(b *testing.B) {
	var p *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pointer.GetOr(p, 100)
	}
}

// BenchmarkGetOrZero 基准测试 GetOrZero 函数
func BenchmarkGetOrZero(b *testing.B) {
	x := 42
	p := &x
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pointer.GetOrZero(p)
	}
}

// BenchmarkGetOrZero_Nil 基准测试 GetOrZero 函数（nil 指针）
func BenchmarkGetOrZero_Nil(b *testing.B) {
	var p *int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pointer.GetOrZero(p)
	}
}

