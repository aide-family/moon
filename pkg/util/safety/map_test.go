package safety

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []map[string]int
		expected map[string]int
	}{
		{
			name:     "empty map",
			input:    []map[string]int{},
			expected: map[string]int{},
		},
		{
			name: "single map",
			input: []map[string]int{
				{"a": 1, "b": 2},
			},
			expected: map[string]int{"a": 1, "b": 2},
		},
		{
			name: "multiple maps",
			input: []map[string]int{
				{"a": 1, "b": 2},
				{"c": 3, "d": 4},
			},
			expected: map[string]int{"a": 1, "b": 2, "c": 3, "d": 4},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMap(tt.input...)
			assert.Equal(t, len(tt.expected), m.Len())

			for k, v := range tt.expected {
				val, ok := m.Get(k)
				assert.True(t, ok)
				assert.Equal(t, v, val)
			}
		})
	}
}

func TestMap_Len(t *testing.T) {
	m := NewMap[string, int]()
	assert.Equal(t, 0, m.Len())

	m.Set("a", 1)
	assert.Equal(t, 1, m.Len())

	m.Set("b", 2)
	assert.Equal(t, 2, m.Len())

	m.Delete("a")
	assert.Equal(t, 1, m.Len())

	m.Clear()
	assert.Equal(t, 0, m.Len())
}

func TestMap_Get(t *testing.T) {
	m := NewMap[string, int]()

	// Test getting from empty map
	val, ok := m.Get("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, 0, val)

	// Test getting existing value
	m.Set("key", 42)
	val, ok = m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 42, val)

	// Test getting non-existent key
	val, ok = m.Get("other")
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestMap_Set(t *testing.T) {
	m := NewMap[string, int]()

	// Test setting new value
	m.Set("key", 42)
	val, ok := m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 42, val)
	assert.Equal(t, 1, m.Len())

	// Test overwriting existing value
	m.Set("key", 100)
	val, ok = m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 100, val)
	assert.Equal(t, 1, m.Len()) // Length should still be 1
}

func TestMap_Delete(t *testing.T) {
	m := NewMap[string, int]()

	// Test deleting from empty map
	m.Delete("nonexistent")
	assert.Equal(t, 0, m.Len())

	// Test deleting existing value
	m.Set("key", 42)
	assert.Equal(t, 1, m.Len())

	m.Delete("key")
	assert.Equal(t, 0, m.Len())

	val, ok := m.Get("key")
	assert.False(t, ok)
	assert.Equal(t, 0, val)

	// Test deleting same key multiple times
	m.Delete("key")
	assert.Equal(t, 0, m.Len())
}

func TestMap_Append(t *testing.T) {
	m := NewMap[string, int]()

	// Test appending to empty map
	m.Append(map[string]int{"a": 1, "b": 2})
	assert.Equal(t, 2, m.Len())

	val, ok := m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	val, ok = m.Get("b")
	assert.True(t, ok)
	assert.Equal(t, 2, val)

	// Test appending multiple maps
	m.Append(
		map[string]int{"c": 3},
		map[string]int{"d": 4, "e": 5},
	)
	assert.Equal(t, 5, m.Len())

	val, ok = m.Get("c")
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	val, ok = m.Get("d")
	assert.True(t, ok)
	assert.Equal(t, 4, val)

	val, ok = m.Get("e")
	assert.True(t, ok)
	assert.Equal(t, 5, val)

	// Test overwriting existing keys
	m.Append(map[string]int{"a": 100})
	assert.Equal(t, 5, m.Len()) // Length should still be 5

	val, ok = m.Get("a")
	assert.True(t, ok)
	assert.Equal(t, 100, val) // Value should be updated
}

func TestMap_List(t *testing.T) {
	m := NewMap[string, int]()

	// Test empty map
	result := m.List()
	assert.Empty(t, result)

	// Test with values
	expected := map[string]int{"a": 1, "b": 2, "c": 3}
	for k, v := range expected {
		m.Set(k, v)
	}

	result = m.List()
	assert.Equal(t, expected, result)
	assert.Equal(t, len(expected), len(result))
}

func TestMap_Clear(t *testing.T) {
	m := NewMap[string, int]()

	// Test clearing empty map
	m.Clear()
	assert.Equal(t, 0, m.Len())

	// Test clearing map with values
	m.Set("a", 1)
	m.Set("b", 2)
	assert.Equal(t, 2, m.Len())

	m.Clear()
	assert.Equal(t, 0, m.Len())

	// Verify all values are gone
	val, ok := m.Get("a")
	assert.False(t, ok)
	assert.Equal(t, 0, val)

	val, ok = m.Get("b")
	assert.False(t, ok)
	assert.Equal(t, 0, val)
}

func TestMap_First(t *testing.T) {
	m := NewMap[string, int]()

	// Test empty map
	val, ok := m.First()
	assert.False(t, ok)
	assert.Equal(t, 0, val)

	// Test with single value
	m.Set("key", 42)
	val, ok = m.First()
	assert.True(t, ok)
	assert.Equal(t, 42, val)

	// Test with multiple values (should return first one found)
	m.Set("another", 100)
	val, ok = m.First()
	assert.True(t, ok)
	// Note: sync.Map iteration order is not guaranteed, so we just check that we got a value
	assert.True(t, val == 42 || val == 100)
}

func TestMap_Concurrency(t *testing.T) {
	m := NewMap[int, string]()
	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				m.Set(key, "value")
			}
		}(i)
	}

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				m.Get(key)
			}
		}(i)
	}

	// Test concurrent deletes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := id*numOperations + j
				m.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// Verify the map is still functional after concurrent operations
	m.Set(999, "test")
	val, ok := m.Get(999)
	assert.True(t, ok)
	assert.Equal(t, "test", val)
}

func TestMap_ComplexTypes(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	m := NewMap[string, Person]()

	person1 := Person{Name: "Alice", Age: 30}
	person2 := Person{Name: "Bob", Age: 25}

	m.Set("alice", person1)
	m.Set("bob", person2)

	assert.Equal(t, 2, m.Len())

	val, ok := m.Get("alice")
	assert.True(t, ok)
	assert.Equal(t, person1, val)

	val, ok = m.Get("bob")
	assert.True(t, ok)
	assert.Equal(t, person2, val)

	// Test List with complex types
	result := m.List()
	expected := map[string]Person{
		"alice": person1,
		"bob":   person2,
	}
	assert.Equal(t, expected, result)
}

func TestMap_PointerTypes(t *testing.T) {
	m := NewMap[string, *int]()

	val1 := 42
	val2 := 100

	m.Set("ptr1", &val1)
	m.Set("ptr2", &val2)

	assert.Equal(t, 2, m.Len())

	result, ok := m.Get("ptr1")
	assert.True(t, ok)
	assert.Equal(t, &val1, result)
	assert.Equal(t, 42, *result)

	// Test First with pointer types
	first, ok := m.First()
	assert.True(t, ok)
	assert.NotNil(t, first)
}

func TestMap_LengthTracking(t *testing.T) {
	m := NewMap[string, int]()

	// Test length tracking with Set
	assert.Equal(t, 0, m.Len())

	m.Set("a", 1)
	assert.Equal(t, 1, m.Len())

	m.Set("b", 2)
	assert.Equal(t, 2, m.Len())

	// Test length tracking with Delete
	m.Delete("a")
	assert.Equal(t, 1, m.Len())

	m.Delete("b")
	assert.Equal(t, 0, m.Len())

	// Test length tracking with Clear
	m.Set("c", 3)
	m.Set("d", 4)
	assert.Equal(t, 2, m.Len())

	m.Clear()
	assert.Equal(t, 0, m.Len())

	// Test length tracking with Append
	m.Append(map[string]int{"e": 5, "f": 6})
	assert.Equal(t, 2, m.Len())

	m.Append(map[string]int{"g": 7})
	assert.Equal(t, 3, m.Len())
}

func TestMap_StressTest(t *testing.T) {
	m := NewMap[int, string]()
	const numOperations = 10000

	// Stress test with many operations
	for i := 0; i < numOperations; i++ {
		m.Set(i, "value")
	}

	assert.Equal(t, numOperations, m.Len())

	// Verify all values are present
	for i := 0; i < numOperations; i++ {
		val, ok := m.Get(i)
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	}

	// Delete all values
	for i := 0; i < numOperations; i++ {
		m.Delete(i)
	}

	assert.Equal(t, 0, m.Len())

	// Verify all values are gone
	for i := 0; i < numOperations; i++ {
		val, ok := m.Get(i)
		assert.False(t, ok)
		assert.Equal(t, "", val)
	}
}

func TestMap_RaceCondition(t *testing.T) {
	m := NewMap[int, int]()
	const numGoroutines = 10
	const duration = 100 * time.Millisecond

	done := make(chan bool)

	// Start multiple goroutines that continuously read and write
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					m.Set(id, id*2)
					m.Get(id)
					m.Delete(id)
				}
			}
		}(i)
	}

	// Let them run for a while
	time.Sleep(duration)
	close(done)

	// Verify the map is still functional
	m.Set(999, 999)
	val, ok := m.Get(999)
	assert.True(t, ok)
	assert.Equal(t, 999, val)
}

func BenchmarkMap_Set(b *testing.B) {
	m := NewMap[int, string]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Set(i, "value")
	}
}

func BenchmarkMap_Get(b *testing.B) {
	m := NewMap[int, string]()
	for i := 0; i < 1000; i++ {
		m.Set(i, "value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(i % 1000)
	}
}

func BenchmarkMap_Concurrent(b *testing.B) {
	m := NewMap[int, string]()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(i, "value")
			m.Get(i)
			i++
		}
	})
}

// TestMap_RaceConditions 测试 Map 的竞态条件
func TestMap_RaceConditions(t *testing.T) {
	m := NewMap[int, string]()
	const numGoroutines = 100
	const operationsPerGoroutine = 1000

	var wg sync.WaitGroup

	// 并发读写测试
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := id*operationsPerGoroutine + j
				m.Set(key, "value")
				m.Get(key)
				m.Len()
			}
		}(i)
	}

	// 并发删除测试
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				key := id*operationsPerGoroutine + j
				m.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}

// TestMap_MemoryLeak 测试 Map 内存泄漏
func TestMap_MemoryLeak(t *testing.T) {
	// 记录初始内存使用
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	m := NewMap[int, string]()

	// 大量插入和删除操作
	for i := 0; i < 10000; i++ {
		m.Set(i, "value")
	}

	for i := 0; i < 10000; i++ {
		m.Delete(i)
	}

	// 强制垃圾回收
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// 检查内存增长是否合理（允许一定的增长）
	memoryGrowth := int64(m2.Alloc) - int64(m1.Alloc)
	if memoryGrowth > 1024*1024 { // 1MB
		t.Errorf("Potential memory leak detected: memory growth %d bytes", memoryGrowth)
	}
}

// TestMap_StressTestRace 压力测试
func TestMap_StressTestRace(t *testing.T) {
	m := NewMap[int, string]()
	const numOperations = 100000

	// 并发压力测试
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations/10; j++ {
				key := id*(numOperations/10) + j
				m.Set(key, "value")
				m.Get(key)
				if j%10 == 0 {
					m.Delete(key)
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	expectedLen := numOperations - (numOperations / 10)
	if m.Len() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, m.Len())
	}
}

// TestMap_ConcurrentModification 测试并发修改
func TestMap_ConcurrentModification(t *testing.T) {
	m := NewMap[int, string]()
	const numGoroutines = 20
	const duration = 100 * time.Millisecond

	done := make(chan bool)

	// 启动多个 goroutine 进行并发修改
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					key := id*100 + int(time.Now().UnixNano()%1000)
					m.Set(key, "value")
					m.Get(key)
					m.Delete(key)
				}
			}
		}(i)
	}

	// 运行一段时间
	time.Sleep(duration)
	close(done)

	// 验证 map 仍然可用
	m.Set(999, "test")
	val, ok := m.Get(999)
	if !ok || val != "test" {
		t.Error("Map became unusable after concurrent modification")
	}
}

// TestGoroutineLeak 测试 goroutine 泄漏
func TestGoroutineLeak(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	// 执行大量并发操作
	m := NewMap[int, string]()
	s := NewSlice[int](0)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				key := id*100 + j
				m.Set(key, "value")
				s.Append(key)
				m.Get(key)
				if s.Len() > 0 {
					s.Pop()
				}
			}
		}(i)
	}

	wg.Wait()

	// 等待一段时间让所有 goroutine 完成
	time.Sleep(100 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	goroutineGrowth := finalGoroutines - initialGoroutines

	// 允许少量 goroutine 增长（测试框架等）
	if goroutineGrowth > 10 {
		t.Errorf("Potential goroutine leak detected: goroutine growth %d", goroutineGrowth)
	}
}

// BenchmarkMap_ConcurrentReadWrite 并发读写基准测试
func BenchmarkMap_ConcurrentReadWrite(b *testing.B) {
	m := NewMap[int, string]()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m.Set(i, "value")
			m.Get(i)
			m.Len()
			i++
		}
	})
}
