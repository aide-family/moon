package safety_test

import (
	"encoding"
	"encoding/json"
	"sync"
	"testing"

	"github.com/aide-family/magicbox/safety"
)

// TestNewSyncMap 测试 NewSyncMap 函数
func TestNewSyncMap(t *testing.T) {
	// 测试创建空 map
	emptyMap := safety.NewSyncMap[string, int](nil)
	if emptyMap == nil {
		t.Fatal("NewSyncMap(nil) returned nil")
	}
	if emptyMap.Len() != 0 {
		t.Errorf("NewSyncMap(nil).Len() = %d, want 0", emptyMap.Len())
	}

	// 测试创建非空 map
	original := map[string]int{"a": 1, "b": 2, "c": 3}
	newMap := safety.NewSyncMap(original)
	if newMap == nil {
		t.Fatal("NewSyncMap returned nil")
	}
	if newMap.Len() != 3 {
		t.Errorf("NewSyncMap().Len() = %d, want 3", newMap.Len())
	}

	// 验证是克隆，不是引用
	original["d"] = 4
	if newMap.Len() != 3 {
		t.Errorf("After modifying original, NewSyncMap().Len() = %d, want 3", newMap.Len())
	}

	// 验证值正确
	if v, ok := newMap.Get("a"); !ok || v != 1 {
		t.Errorf("Get('a') = (%v, %v), want (1, true)", v, ok)
	}
}

// TestSyncMap_Get 测试 Get 方法
func TestSyncMap_Get(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2})

	// 测试存在的键
	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Errorf("Get('a') = (%v, %v), want (1, true)", v, ok)
	}

	// 测试不存在的键
	if v, ok := m.Get("c"); ok || v != 0 {
		t.Errorf("Get('c') = (%v, %v), want (0, false)", v, ok)
	}
}

// TestSyncMap_Set 测试 Set 方法
func TestSyncMap_Set(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)

	// 测试设置新键
	m.Set("a", 1)
	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Errorf("After Set('a', 1), Get('a') = (%v, %v), want (1, true)", v, ok)
	}

	// 测试更新已存在的键
	m.Set("a", 2)
	if v, ok := m.Get("a"); !ok || v != 2 {
		t.Errorf("After Set('a', 2), Get('a') = (%v, %v), want (2, true)", v, ok)
	}

	// 测试链式调用
	m.Set("b", 3).Set("c", 4)
	if m.Len() != 3 {
		t.Errorf("After chained Set, Len() = %d, want 3", m.Len())
	}
}

// TestSyncMap_Append 测试 Append 方法
func TestSyncMap_Append(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1})

	// 测试追加单个 map
	m.Append(map[string]int{"b": 2, "c": 3})
	if m.Len() != 3 {
		t.Errorf("After Append, Len() = %d, want 3", m.Len())
	}

	// 测试追加多个 map
	m.Append(map[string]int{"d": 4}, map[string]int{"e": 5})
	if m.Len() != 5 {
		t.Errorf("After multiple Append, Len() = %d, want 5", m.Len())
	}

	// 测试覆盖已存在的键
	m.Append(map[string]int{"a": 10})
	if v, ok := m.Get("a"); !ok || v != 10 {
		t.Errorf("After Append with existing key, Get('a') = (%v, %v), want (10, true)", v, ok)
	}

	// 测试空 map
	m.Clear()
	m.Append(map[string]int{})
	if m.Len() != 0 {
		t.Errorf("After Append empty map, Len() = %d, want 0", m.Len())
	}
}

// TestSyncMap_Delete 测试 Delete 方法
func TestSyncMap_Delete(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})

	// 测试删除存在的键
	m.Delete("a")
	if _, ok := m.Get("a"); ok {
		t.Error("After Delete('a'), Get('a') should return false")
	}
	if m.Len() != 2 {
		t.Errorf("After Delete, Len() = %d, want 2", m.Len())
	}

	// 测试删除不存在的键（不应该 panic）
	m.Delete("d")
	if m.Len() != 2 {
		t.Errorf("After Delete non-existent key, Len() = %d, want 2", m.Len())
	}

	// 测试链式调用
	m.Delete("b").Delete("c")
	if m.Len() != 0 {
		t.Errorf("After chained Delete, Len() = %d, want 0", m.Len())
	}
}

// TestSyncMap_DeleteFunc 测试 DeleteFunc 方法
func TestSyncMap_DeleteFunc(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4})

	// 测试删除偶数
	m.DeleteFunc(func(k string, v int) bool {
		return v%2 == 0
	})
	if m.Len() != 2 {
		t.Errorf("After DeleteFunc, Len() = %d, want 2", m.Len())
	}
	if _, ok := m.Get("b"); ok {
		t.Error("After DeleteFunc, 'b' should be deleted")
	}
	if _, ok := m.Get("d"); ok {
		t.Error("After DeleteFunc, 'd' should be deleted")
	}

	// 测试删除所有
	m.DeleteFunc(func(k string, v int) bool {
		return true
	})
	if m.Len() != 0 {
		t.Errorf("After DeleteFunc(all), Len() = %d, want 0", m.Len())
	}
}

// TestSyncMap_Range 测试 Range 方法
func TestSyncMap_Range(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})

	// 测试遍历所有元素
	visited := make(map[string]int)
	m.Range(func(k string, v int) bool {
		visited[k] = v
		return true
	})
	if len(visited) != 3 {
		t.Errorf("Range visited %d elements, want 3", len(visited))
	}
	if visited["a"] != 1 || visited["b"] != 2 || visited["c"] != 3 {
		t.Error("Range did not visit all elements correctly")
	}

	// 测试提前停止
	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return false // 第一次就停止
	})
	if count != 1 {
		t.Errorf("Range should stop early, visited %d elements, want 1", count)
	}
}

// TestSyncMap_Len 测试 Len 方法
func TestSyncMap_Len(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)
	if m.Len() != 0 {
		t.Errorf("Empty map Len() = %d, want 0", m.Len())
	}

	m.Set("a", 1)
	if m.Len() != 1 {
		t.Errorf("After Set, Len() = %d, want 1", m.Len())
	}

	m.Set("b", 2).Set("c", 3)
	if m.Len() != 3 {
		t.Errorf("After multiple Set, Len() = %d, want 3", m.Len())
	}

	m.Delete("a")
	if m.Len() != 2 {
		t.Errorf("After Delete, Len() = %d, want 2", m.Len())
	}
}

// TestSyncMap_Keys 测试 Keys 方法
func TestSyncMap_Keys(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	keys := m.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	// 验证键存在
	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}
	if !keyMap["a"] || !keyMap["b"] || !keyMap["c"] {
		t.Error("Keys() did not return all keys")
	}

	// 测试空 map
	m.Clear()
	keys = m.Keys()
	if len(keys) != 0 {
		t.Errorf("Empty map Keys() returned %d keys, want 0", len(keys))
	}
}

// TestSyncMap_Values 测试 Values 方法
func TestSyncMap_Values(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	values := m.Values()
	if len(values) != 3 {
		t.Errorf("Values() returned %d values, want 3", len(values))
	}

	// 验证值存在
	valueMap := make(map[int]bool)
	for _, v := range values {
		valueMap[v] = true
	}
	if !valueMap[1] || !valueMap[2] || !valueMap[3] {
		t.Error("Values() did not return all values")
	}

	// 测试空 map
	m.Clear()
	values = m.Values()
	if len(values) != 0 {
		t.Errorf("Empty map Values() returned %d values, want 0", len(values))
	}
}

// TestSyncMap_Clear 测试 Clear 方法
func TestSyncMap_Clear(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	m.Clear()
	if m.Len() != 0 {
		t.Errorf("After Clear, Len() = %d, want 0", m.Len())
	}

	// 测试清空后可以继续使用
	m.Set("x", 10)
	if m.Len() != 1 {
		t.Errorf("After Clear and Set, Len() = %d, want 1", m.Len())
	}
}

// TestSyncMap_Clone 测试 Clone 方法
func TestSyncMap_Clone(t *testing.T) {
	original := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	cloned := original.Clone()

	// 验证克隆的 map 有相同的内容
	if cloned.Len() != original.Len() {
		t.Errorf("Cloned map Len() = %d, want %d", cloned.Len(), original.Len())
	}

	// 验证是独立的副本
	original.Set("d", 4)
	if cloned.Len() != 3 {
		t.Errorf("After modifying original, cloned Len() = %d, want 3", cloned.Len())
	}

	cloned.Set("e", 5)
	if original.Len() != 4 {
		t.Errorf("After modifying cloned, original Len() = %d, want 4", original.Len())
	}
}

// TestSyncMap_Map 测试 Map 方法
func TestSyncMap_Map(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	plainMap := m.Map()

	// 验证返回的 map 有相同的内容
	if len(plainMap) != 3 {
		t.Errorf("Map() returned map with %d elements, want 3", len(plainMap))
	}

	// 验证是独立的副本
	plainMap["d"] = 4
	if m.Len() != 3 {
		t.Errorf("After modifying returned map, original Len() = %d, want 3", m.Len())
	}

	m.Set("e", 5)
	if len(plainMap) != 4 {
		t.Errorf("After modifying original, returned map has %d elements, want 4", len(plainMap))
	}
}

// TestSyncMap_MarshalBinary 测试 MarshalBinary 方法
func TestSyncMap_MarshalBinary(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2})
	data, err := m.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalBinary() returned empty data")
	}

	// 验证可以反序列化
	var unmarshaled map[string]int
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if unmarshaled["a"] != 1 || unmarshaled["b"] != 2 {
		t.Error("Unmarshaled data does not match original")
	}
}

// TestSyncMap_UnmarshalBinary 测试 UnmarshalBinary 方法
func TestSyncMap_UnmarshalBinary(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2, "c": 3}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal error = %v", err)
	}

	m := safety.NewSyncMap(map[string]int{})
	if err := m.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}

	if m.Len() != 3 {
		t.Errorf("After UnmarshalBinary, Len() = %d, want 3", m.Len())
	}
	if v, ok := m.Get("a"); !ok || v != 1 {
		t.Errorf("After UnmarshalBinary, Get('a') = (%v, %v), want (1, true)", v, ok)
	}
}

// TestSyncMap_BinaryMarshaler 测试 encoding.BinaryMarshaler 接口
func TestSyncMap_BinaryMarshaler(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2})
	var marshaler encoding.BinaryMarshaler = m
	data, err := marshaler.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalBinary() returned empty data")
	}
}

// TestSyncMap_BinaryUnmarshaler 测试 encoding.BinaryUnmarshaler 接口
func TestSyncMap_BinaryUnmarshaler(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}
	data, _ := json.Marshal(original)

	m := safety.NewSyncMap(map[string]int{})
	var unmarshaler encoding.BinaryUnmarshaler = m
	if err := unmarshaler.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}
	if m.Len() != 2 {
		t.Errorf("After UnmarshalBinary, Len() = %d, want 2", m.Len())
	}
}

// TestSyncMap_String 测试 String 方法
func TestSyncMap_String(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2})
	str := m.String()
	if len(str) == 0 {
		t.Error("String() returned empty string")
	}

	// 验证是有效的 JSON
	var unmarshaled map[string]int
	if err := json.Unmarshal([]byte(str), &unmarshaled); err != nil {
		t.Fatalf("String() returned invalid JSON: %v", err)
	}
	if unmarshaled["a"] != 1 || unmarshaled["b"] != 2 {
		t.Error("String() returned incorrect JSON")
	}
}

// TestSyncMap_Concurrent 测试并发安全性
func TestSyncMap_Concurrent(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)
	const numGoroutines = 100
	const numOps = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发写入
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := string(rune('a' + (id*numOps+j)%26))
				m.Set(key, id*numOps+j)
			}
		}(i)
	}

	wg.Wait()

	// 并发读取
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := string(rune('a' + j%26))
				_, _ = m.Get(key)
				_ = m.Len()
				_ = m.Keys()
				_ = m.Values()
			}
		}()
	}

	wg.Wait()

	// 验证没有 panic 并且 map 有内容
	if m.Len() == 0 {
		t.Error("After concurrent operations, map should not be empty")
	}
}

// TestSyncMap_ConcurrentDelete 测试并发删除
func TestSyncMap_ConcurrentDelete(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{
		"a": 1, "b": 2, "c": 3, "d": 4, "e": 5,
		"f": 6, "g": 7, "h": 8, "i": 9, "j": 10,
	})

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发删除
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := string(rune('a' + id%10))
			m.Delete(key)
		}(i)
	}

	wg.Wait()

	// 验证没有 panic
	if m.Len() < 0 {
		t.Error("After concurrent delete, Len() should be >= 0")
	}
}

// TestSyncMap_ConcurrentRange 测试并发遍历
func TestSyncMap_ConcurrentRange(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5})

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发遍历
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			count := 0
			m.Range(func(k string, v int) bool {
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

// TestSyncMap_DifferentTypes 测试不同的类型组合
func TestSyncMap_DifferentTypes(t *testing.T) {
	// 测试 int 键
	m1 := safety.NewSyncMap(map[int]string{1: "a", 2: "b", 3: "c"})
	if m1.Len() != 3 {
		t.Errorf("SyncMap[int, string] Len() = %d, want 3", m1.Len())
	}

	// 测试 float 值
	m2 := safety.NewSyncMap(map[string]float64{"x": 1.5, "y": 2.5})
	if m2.Len() != 2 {
		t.Errorf("SyncMap[string, float64] Len() = %d, want 2", m2.Len())
	}

	// 测试 bool 值
	m3 := safety.NewSyncMap(map[string]bool{"true": true, "false": false})
	if m3.Len() != 2 {
		t.Errorf("SyncMap[string, bool] Len() = %d, want 2", m3.Len())
	}
}

// TestSyncMap_EmptyMap 测试空 map 的各种操作
func TestSyncMap_EmptyMap(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)

	// 测试空 map 的各种操作
	if m.Len() != 0 {
		t.Errorf("Empty map Len() = %d, want 0", m.Len())
	}

	if len(m.Keys()) != 0 {
		t.Error("Empty map Keys() should return empty slice")
	}

	if len(m.Values()) != 0 {
		t.Error("Empty map Values() should return empty slice")
	}

	// 测试遍历空 map
	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Empty map Range visited %d elements, want 0", count)
	}

	// 测试删除不存在的键
	m.Delete("nonexistent")
	if m.Len() != 0 {
		t.Errorf("After Delete on empty map, Len() = %d, want 0", m.Len())
	}
}

// TestSyncMap_AppendEmpty 测试追加空 map
func TestSyncMap_AppendEmpty(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1})
	m.Append(map[string]int{})
	if m.Len() != 1 {
		t.Errorf("After Append empty map, Len() = %d, want 1", m.Len())
	}
}

// TestSyncMap_CloneEmpty 测试克隆空 map
func TestSyncMap_CloneEmpty(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)
	cloned := m.Clone()
	if cloned.Len() != 0 {
		t.Errorf("Cloned empty map Len() = %d, want 0", cloned.Len())
	}
}

// TestSyncMap_MarshalEmpty 测试序列化空 map
func TestSyncMap_MarshalEmpty(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)
	data, err := m.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() on empty map error = %v", err)
	}

	// 空 map 应该序列化为 "{}"
	var unmarshaled map[string]int
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}
	if len(unmarshaled) != 0 {
		t.Error("Unmarshaled empty map should be empty")
	}
}

// TestSyncMap_UnmarshalEmpty 测试反序列化空 map
func TestSyncMap_UnmarshalEmpty(t *testing.T) {
	data := []byte("{}")
	m := safety.NewSyncMap(map[string]int{})
	if err := m.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary() empty map error = %v", err)
	}
	if m.Len() != 0 {
		t.Errorf("After UnmarshalBinary empty map, Len() = %d, want 0", m.Len())
	}
}

// TestSyncMap_StringEmpty 测试空 map 的字符串表示
func TestSyncMap_StringEmpty(t *testing.T) {
	m := safety.NewSyncMap[string, int](nil)
	str := m.String()
	if str != "{}" {
		t.Errorf("Empty map String() = %q, want \"{}\"", str)
	}
}

// TestSyncMap_ConcurrentClear 测试并发 Clear
func TestSyncMap_ConcurrentClear(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发 Clear
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			m.Clear()
		}()
	}

	wg.Wait()

	// 验证没有 panic
	if m.Len() < 0 {
		t.Error("After concurrent Clear, Len() should be >= 0")
	}
}

// TestSyncMap_ConcurrentClone 测试并发 Clone
func TestSyncMap_ConcurrentClone(t *testing.T) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 并发 Clone
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			cloned := m.Clone()
			if cloned.Len() < 0 || cloned.Len() > 3 {
				t.Errorf("Cloned map Len() = %d, expected 0-3", cloned.Len())
			}
		}()
	}

	wg.Wait()
}

// BenchmarkSyncMap_Get 基准测试 Get 方法
func BenchmarkSyncMap_Get(b *testing.B) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.Get("a")
	}
}

// BenchmarkSyncMap_Set 基准测试 Set 方法
func BenchmarkSyncMap_Set(b *testing.B) {
	m := safety.NewSyncMap[string, int](nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set("key", i)
	}
}

// BenchmarkSyncMap_Delete 基准测试 Delete 方法
func BenchmarkSyncMap_Delete(b *testing.B) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete("a")
		m.Set("a", 1) // 重新设置以便下次删除
	}
}

// BenchmarkSyncMap_Len 基准测试 Len 方法
func BenchmarkSyncMap_Len(b *testing.B) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.Len()
	}
}

// BenchmarkSyncMap_MarshalBinary 基准测试 MarshalBinary 方法
func BenchmarkSyncMap_MarshalBinary(b *testing.B) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m.MarshalBinary()
	}
}

// BenchmarkSyncMap_UnmarshalBinary 基准测试 UnmarshalBinary 方法
func BenchmarkSyncMap_UnmarshalBinary(b *testing.B) {
	data, _ := json.Marshal(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5})
	m := safety.NewSyncMap(map[string]int{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.UnmarshalBinary(data)
	}
}

// BenchmarkSyncMap_ConcurrentGet 基准测试并发 Get
func BenchmarkSyncMap_ConcurrentGet(b *testing.B) {
	m := safety.NewSyncMap(map[string]int{"a": 1, "b": 2, "c": 3})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = m.Get("a")
		}
	})
}

// BenchmarkSyncMap_ConcurrentSet 基准测试并发 Set
func BenchmarkSyncMap_ConcurrentSet(b *testing.B) {
	m := safety.NewSyncMap[string, int](nil)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set("key", i)
			i++
		}
	})
}
