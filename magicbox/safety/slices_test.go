package safety_test

import (
	"encoding"
	"encoding/json"
	"sync"
	"testing"

	"github.com/aide-family/magicbox/safety"
)

// TestNewSlice 测试 NewSlice 函数
func TestNewSlice(t *testing.T) {
	// 测试创建空 slice
	emptySlice := safety.NewSlice[int](nil)
	if emptySlice == nil {
		t.Fatal("NewSlice(nil) returned nil")
	}
	if emptySlice.Len() != 0 {
		t.Errorf("NewSlice(nil).Len() = %d, want 0", emptySlice.Len())
	}

	// 测试创建非空 slice
	original := []int{1, 2, 3}
	newSlice := safety.NewSlice(original)
	if newSlice == nil {
		t.Fatal("NewSlice returned nil")
	}
	if newSlice.Len() != 3 {
		t.Errorf("NewSlice().Len() = %d, want 3", newSlice.Len())
	}

	// 验证是克隆，不是引用
	original[0] = 10
	if newSlice.Get(0) != 1 {
		t.Errorf("After modifying original, NewSlice().Get(0) = %d, want 1", newSlice.Get(0))
	}

	// 验证值正确
	if v := newSlice.Get(0); v != 1 {
		t.Errorf("Get(0) = %d, want 1", v)
	}
}

// TestSlice_Get 测试 Get 方法
func TestSlice_Get(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})

	// 测试有效索引
	if v := s.Get(0); v != 1 {
		t.Errorf("Get(0) = %d, want 1", v)
	}
	if v := s.Get(2); v != 3 {
		t.Errorf("Get(2) = %d, want 3", v)
	}

	// 测试索引越界（应该 panic）
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get with out-of-bounds index should panic")
		}
	}()
	s.Get(10)
}

// TestSlice_Set 测试 Set 方法
func TestSlice_Set(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})

	// 测试设置有效索引
	s.Set(0, 10)
	if v := s.Get(0); v != 10 {
		t.Errorf("After Set(0, 10), Get(0) = %d, want 10", v)
	}

	// 测试链式调用
	s.Set(1, 20).Set(2, 30)
	if s.Get(1) != 20 || s.Get(2) != 30 {
		t.Error("Chained Set did not work correctly")
	}

	// 测试索引越界（应该 panic）
	defer func() {
		if r := recover(); r == nil {
			t.Error("Set with out-of-bounds index should panic")
		}
	}()
	s.Set(10, 100)
}

// TestSlice_Append 测试 Append 方法
func TestSlice_Append(t *testing.T) {
	s := safety.NewSlice([]int{1, 2})

	// 测试追加单个元素
	s.Append(3)
	if s.Len() != 3 {
		t.Errorf("After Append, Len() = %d, want 3", s.Len())
	}
	if v := s.Get(2); v != 3 {
		t.Errorf("After Append, Get(2) = %d, want 3", v)
	}

	// 测试链式调用
	s.Append(4).Append(5)
	if s.Len() != 5 {
		t.Errorf("After chained Append, Len() = %d, want 5", s.Len())
	}

	// 测试空 slice 上追加
	empty := safety.NewSlice[int](nil)
	empty.Append(1)
	if empty.Len() != 1 {
		t.Errorf("After Append on empty slice, Len() = %d, want 1", empty.Len())
	}
}

// TestSlice_AppendSlice 测试 AppendSlice 方法
func TestSlice_AppendSlice(t *testing.T) {
	s := safety.NewSlice([]int{1, 2})

	// 测试追加单个 slice
	s.AppendSlice([]int{3, 4})
	if s.Len() != 4 {
		t.Errorf("After AppendSlice, Len() = %d, want 4", s.Len())
	}
	if s.Get(2) != 3 || s.Get(3) != 4 {
		t.Error("AppendSlice did not append correctly")
	}

	// 测试追加多个 slice
	s.AppendSlice([]int{5}, []int{6, 7})
	if s.Len() != 7 {
		t.Errorf("After multiple AppendSlice, Len() = %d, want 7", s.Len())
	}

	// 测试追加空 slice
	s.AppendSlice([]int{})
	if s.Len() != 7 {
		t.Errorf("After AppendSlice empty, Len() = %d, want 7", s.Len())
	}

	// 测试空 slice 上追加
	empty := safety.NewSlice[int](nil)
	empty.AppendSlice([]int{1, 2})
	if empty.Len() != 2 {
		t.Errorf("After AppendSlice on empty slice, Len() = %d, want 2", empty.Len())
	}
}

// TestSlice_Delete 测试 Delete 方法
func TestSlice_Delete(t *testing.T) {
	// 测试删除中间元素（索引 2，值为 3）
	s1 := safety.NewSlice([]int{1, 2, 3, 4, 5})
	s1.Delete(2)
	if s1.Len() != 4 {
		t.Errorf("After Delete(2), Len() = %d, want 4", s1.Len())
	}
	// 删除后应该是 [1, 2, 4, 5]，索引 2 的值应该是 4
	if s1.Get(2) != 4 {
		t.Errorf("After Delete(2), Get(2) = %d, want 4", s1.Get(2))
	}

	// 测试删除第一个元素
	s2 := safety.NewSlice([]int{1, 2, 3})
	s2.Delete(0)
	if s2.Len() != 2 {
		t.Errorf("After Delete(0), Len() = %d, want 2", s2.Len())
	}
	if s2.Get(0) != 2 {
		t.Errorf("After Delete(0), Get(0) = %d, want 2", s2.Get(0))
	}

	// 测试删除最后一个元素
	s3 := safety.NewSlice([]int{1, 2, 3})
	lastIndex := s3.Len() - 1
	s3.Delete(lastIndex)
	if s3.Len() != 2 {
		t.Errorf("After Delete last, Len() = %d, want 2", s3.Len())
	}
	if s3.Get(0) != 1 || s3.Get(1) != 2 {
		t.Error("After Delete last, values are incorrect")
	}

	// 测试链式调用
	s4 := safety.NewSlice([]int{1, 2, 3})
	s4.Delete(0).Delete(0)
	if s4.Len() != 1 {
		t.Errorf("After chained Delete, Len() = %d, want 1", s4.Len())
	}
	if s4.Get(0) != 3 {
		t.Errorf("After chained Delete, Get(0) = %d, want 3", s4.Get(0))
	}
}

// TestSlice_DeleteFunc 测试 DeleteFunc 方法
func TestSlice_DeleteFunc(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})

	// 测试删除偶数
	s.DeleteFunc(func(v int) bool {
		return v%2 == 0
	})
	if s.Len() != 3 {
		t.Errorf("After DeleteFunc, Len() = %d, want 3", s.Len())
	}
	// 验证只剩下奇数
	values := s.List()
	for _, v := range values {
		if v%2 == 0 {
			t.Errorf("DeleteFunc should have removed even numbers, but found %d", v)
		}
	}

	// 测试删除所有
	s.DeleteFunc(func(v int) bool {
		return true
	})
	if s.Len() != 0 {
		t.Errorf("After DeleteFunc(all), Len() = %d, want 0", s.Len())
	}
}

// TestSlice_Range 测试 Range 方法
func TestSlice_Range(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})

	// 测试遍历所有元素
	visited := make([]int, 0)
	s.Range(func(v int) bool {
		visited = append(visited, v)
		return true
	})
	if len(visited) != 3 {
		t.Errorf("Range visited %d elements, want 3", len(visited))
	}
	if visited[0] != 1 || visited[1] != 2 || visited[2] != 3 {
		t.Error("Range did not visit all elements correctly")
	}

	// 测试提前停止
	count := 0
	s.Range(func(v int) bool {
		count++
		return false // 第一次就停止
	})
	if count != 1 {
		t.Errorf("Range should stop early, visited %d elements, want 1", count)
	}
}

// TestSlice_Len 测试 Len 方法
func TestSlice_Len(t *testing.T) {
	s := safety.NewSlice[int](nil)
	if s.Len() != 0 {
		t.Errorf("Empty slice Len() = %d, want 0", s.Len())
	}

	s.Append(1)
	if s.Len() != 1 {
		t.Errorf("After Append, Len() = %d, want 1", s.Len())
	}

	s.Append(2).Append(3)
	if s.Len() != 3 {
		t.Errorf("After multiple Append, Len() = %d, want 3", s.Len())
	}

	s.Delete(0)
	if s.Len() != 2 {
		t.Errorf("After Delete, Len() = %d, want 2", s.Len())
	}
}

// TestSlice_Clone 测试 Clone 方法
func TestSlice_Clone(t *testing.T) {
	original := safety.NewSlice([]int{1, 2, 3})
	cloned := original.Clone()

	// 验证克隆的 slice 有相同的内容
	if cloned.Len() != original.Len() {
		t.Errorf("Cloned slice Len() = %d, want %d", cloned.Len(), original.Len())
	}
	for i := 0; i < original.Len(); i++ {
		if cloned.Get(i) != original.Get(i) {
			t.Errorf("Cloned slice Get(%d) = %d, want %d", i, cloned.Get(i), original.Get(i))
		}
	}

	// 验证是独立的副本
	original.Append(4)
	if cloned.Len() != 3 {
		t.Errorf("After modifying original, cloned Len() = %d, want 3", cloned.Len())
	}

	cloned.Append(5)
	if original.Len() != 4 {
		t.Errorf("After modifying cloned, original Len() = %d, want 4", original.Len())
	}
}

// TestSlice_Clear 测试 Clear 方法
func TestSlice_Clear(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})
	s.Clear()
	if s.Len() != 0 {
		t.Errorf("After Clear, Len() = %d, want 0", s.Len())
	}

	// 测试清空后可以继续使用
	s.Append(10)
	if s.Len() != 1 {
		t.Errorf("After Clear and Append, Len() = %d, want 1", s.Len())
	}
}

// TestSlice_List 测试 List 方法
func TestSlice_List(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})
	list := s.List()

	// 验证返回的 slice 有相同的内容
	if len(list) != 3 {
		t.Errorf("List() returned slice with %d elements, want 3", len(list))
	}
	if list[0] != 1 || list[1] != 2 || list[2] != 3 {
		t.Error("List() returned incorrect values")
	}

	// 验证是独立的副本
	list[0] = 10
	if s.Get(0) != 1 {
		t.Errorf("After modifying returned list, original Get(0) = %d, want 1", s.Get(0))
	}

	s.Set(0, 20)
	if list[0] != 10 {
		t.Errorf("After modifying original, returned list[0] = %d, want 10", list[0])
	}
}

// TestSlice_MarshalBinary 测试 MarshalBinary 方法
func TestSlice_MarshalBinary(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})
	data, err := s.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalBinary() returned empty data")
	}

	// 验证可以反序列化
	var unmarshaled []int
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if len(unmarshaled) != 3 || unmarshaled[0] != 1 || unmarshaled[1] != 2 || unmarshaled[2] != 3 {
		t.Error("Unmarshaled data does not match original")
	}
}

// TestSlice_UnmarshalBinary 测试 UnmarshalBinary 方法
func TestSlice_UnmarshalBinary(t *testing.T) {
	original := []int{1, 2, 3}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	s := safety.NewSlice([]int{})
	if err := s.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}

	if s.Len() != 3 {
		t.Errorf("After UnmarshalBinary, Len() = %d, want 3", s.Len())
	}
	if s.Get(0) != 1 || s.Get(1) != 2 || s.Get(2) != 3 {
		t.Error("After UnmarshalBinary, values do not match")
	}
}

// TestSlice_BinaryMarshaler 测试 encoding.BinaryMarshaler 接口
func TestSlice_BinaryMarshaler(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})
	var marshaler encoding.BinaryMarshaler = s
	data, err := marshaler.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalBinary() returned empty data")
	}
}

// TestSlice_BinaryUnmarshaler 测试 encoding.BinaryUnmarshaler 接口
func TestSlice_BinaryUnmarshaler(t *testing.T) {
	original := []int{1, 2, 3}
	data, _ := json.Marshal(original)

	s := safety.NewSlice([]int{})
	var unmarshaler encoding.BinaryUnmarshaler = s
	if err := unmarshaler.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}
	if s.Len() != 3 {
		t.Errorf("After UnmarshalBinary, Len() = %d, want 3", s.Len())
	}
}

// TestSlice_String 测试 String 方法
func TestSlice_String(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})
	str := s.String()
	if len(str) == 0 {
		t.Error("String() returned empty string")
	}

	// 验证是有效的 JSON
	var unmarshaled []int
	if err := json.Unmarshal([]byte(str), &unmarshaled); err != nil {
		t.Fatalf("String() returned invalid JSON: %v", err)
	}
	if len(unmarshaled) != 3 || unmarshaled[0] != 1 || unmarshaled[1] != 2 || unmarshaled[2] != 3 {
		t.Error("String() returned incorrect JSON")
	}
}

// TestSlice_Concurrent 测试并发安全性
func TestSlice_Concurrent(t *testing.T) {
	s := safety.NewSlice[int](nil)
	const numGoroutines = 100
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发追加
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				s.Append(id*numOps + j)
			}
		}(i)
	}

	wg.Wait()

	// 验证没有 panic 并且 slice 有内容
	if s.Len() == 0 {
		t.Error("After concurrent operations, slice should not be empty")
	}

	// 并发读取
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOps && j < s.Len(); j++ {
				_ = s.Get(j)
				_ = s.Len()
				_ = s.List()
			}
		}()
	}

	wg.Wait()
}

// TestSlice_ConcurrentDelete 测试并发删除
func TestSlice_ConcurrentDelete(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发删除（注意：删除操作会改变索引，所以需要小心）
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			// 只删除第一个元素，避免索引问题
			if s.Len() > 0 {
				s.Delete(0)
			}
		}()
	}

	wg.Wait()

	// 验证没有 panic
	if s.Len() < 0 {
		t.Error("After concurrent delete, Len() should be >= 0")
	}
}

// TestSlice_ConcurrentRange 测试并发遍历
func TestSlice_ConcurrentRange(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发遍历
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			count := 0
			s.Range(func(v int) bool {
				count++
				return true
			})
			if count < 0 || count > 5 {
				t.Errorf("Range visited %d elements, expected 0-5", count)
			}
		}()
	}

	wg.Wait()
}

// TestSlice_DifferentTypes 测试不同的类型
func TestSlice_DifferentTypes(t *testing.T) {
	// 测试 string 类型
	s1 := safety.NewSlice([]string{"a", "b", "c"})
	if s1.Len() != 3 {
		t.Errorf("Slice[string] Len() = %d, want 3", s1.Len())
	}

	// 测试 float 类型
	s2 := safety.NewSlice([]float64{1.5, 2.5, 3.5})
	if s2.Len() != 3 {
		t.Errorf("Slice[float64] Len() = %d, want 3", s2.Len())
	}

	// 测试 bool 类型
	s3 := safety.NewSlice([]bool{true, false, true})
	if s3.Len() != 3 {
		t.Errorf("Slice[bool] Len() = %d, want 3", s3.Len())
	}
}

// TestSlice_EmptySlice 测试空 slice 的各种操作
func TestSlice_EmptySlice(t *testing.T) {
	s := safety.NewSlice[int](nil)

	// 测试空 slice 的各种操作
	if s.Len() != 0 {
		t.Errorf("Empty slice Len() = %d, want 0", s.Len())
	}

	if len(s.List()) != 0 {
		t.Error("Empty slice List() should return empty slice")
	}

	// 测试遍历空 slice
	count := 0
	s.Range(func(v int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Empty slice Range visited %d elements, want 0", count)
	}

	// 测试在空 slice 上追加
	s.Append(1)
	if s.Len() != 1 {
		t.Errorf("After Append on empty slice, Len() = %d, want 1", s.Len())
	}
}

// TestSlice_AppendSliceEmpty 测试追加空 slice
func TestSlice_AppendSliceEmpty(t *testing.T) {
	s := safety.NewSlice([]int{1, 2})
	s.AppendSlice([]int{})
	if s.Len() != 2 {
		t.Errorf("After AppendSlice empty, Len() = %d, want 2", s.Len())
	}
}

// TestSlice_CloneEmpty 测试克隆空 slice
func TestSlice_CloneEmpty(t *testing.T) {
	s := safety.NewSlice[int](nil)
	cloned := s.Clone()
	if cloned.Len() != 0 {
		t.Errorf("Cloned empty slice Len() = %d, want 0", cloned.Len())
	}
}

// TestSlice_MarshalEmpty 测试序列化空 slice
func TestSlice_MarshalEmpty(t *testing.T) {
	s := safety.NewSlice[int](nil)
	data, err := s.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() on empty slice error = %v", err)
	}

	// 空 slice 应该序列化为 "[]"
	var unmarshaled []int
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if len(unmarshaled) != 0 {
		t.Error("Unmarshaled empty slice should be empty")
	}
}

// TestSlice_UnmarshalEmpty 测试反序列化空 slice
func TestSlice_UnmarshalEmpty(t *testing.T) {
	data := []byte("[]")
	s := safety.NewSlice([]int{})
	if err := s.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() empty slice error = %v", err)
	}
	if s.Len() != 0 {
		t.Errorf("After UnmarshalBinary empty slice, Len() = %d, want 0", s.Len())
	}
}

// TestSlice_StringEmpty 测试空 slice 的字符串表示
func TestSlice_StringEmpty(t *testing.T) {
	s := safety.NewSlice[int](nil)
	str := s.String()
	// 空 slice 序列化为 JSON 时可能是 "[]" 或 "null"
	if str != "[]" && str != "null" {
		t.Errorf("Empty slice String() = %q, want \"[]\" or \"null\"", str)
	}
}

// TestSlice_DeleteOutOfBounds 测试删除越界索引
func TestSlice_DeleteOutOfBounds(t *testing.T) {
	s := safety.NewSlice([]int{1, 2, 3})

	// 测试删除越界索引（应该 panic）
	defer func() {
		if r := recover(); r == nil {
			t.Error("Delete with out-of-bounds index should panic")
		}
	}()
	s.Delete(10)
}

// TestSlice_ConcurrentAppend 测试并发追加
func TestSlice_ConcurrentAppend(t *testing.T) {
	s := safety.NewSlice[int](nil)

	const numGoroutines = 50
	const numOps = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发追加
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				s.Append(id*numOps + j)
			}
		}(i)
	}

	wg.Wait()

	// 验证没有 panic 并且 slice 有内容
	if s.Len() != numGoroutines*numOps {
		t.Errorf("After concurrent Append, Len() = %d, want %d", s.Len(), numGoroutines*numOps)
	}
}

// BenchmarkSlice_Get 基准测试 Get 方法
func BenchmarkSlice_Get(b *testing.B) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Get(2)
	}
}

// BenchmarkSlice_Set 基准测试 Set 方法
func BenchmarkSlice_Set(b *testing.B) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set(2, i)
	}
}

// BenchmarkSlice_Append 基准测试 Append 方法
func BenchmarkSlice_Append(b *testing.B) {
	s := safety.NewSlice[int](nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Append(i)
	}
}

// BenchmarkSlice_Delete 基准测试 Delete 方法
func BenchmarkSlice_Delete(b *testing.B) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Delete(0)
		s.Append(i) // 重新添加以便下次删除
	}
}

// BenchmarkSlice_Len 基准测试 Len 方法
func BenchmarkSlice_Len(b *testing.B) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Len()
	}
}

// BenchmarkSlice_MarshalBinary 基准测试 MarshalBinary 方法
func BenchmarkSlice_MarshalBinary(b *testing.B) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.MarshalBinary()
	}
}

// BenchmarkSlice_UnmarshalBinary 基准测试 UnmarshalBinary 方法
func BenchmarkSlice_UnmarshalBinary(b *testing.B) {
	data, _ := json.Marshal([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	s := safety.NewSlice([]int{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.UnmarshalBinary(data)
	}
}

// BenchmarkSlice_ConcurrentGet 基准测试并发 Get
func BenchmarkSlice_ConcurrentGet(b *testing.B) {
	s := safety.NewSlice([]int{1, 2, 3, 4, 5})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = s.Get(2)
		}
	})
}

// BenchmarkSlice_ConcurrentAppend 基准测试并发 Append
func BenchmarkSlice_ConcurrentAppend(b *testing.B) {
	s := safety.NewSlice[int](nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.Append(i)
			i++
		}
	})
}
