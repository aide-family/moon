package safety

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSlice(t *testing.T) {
	s := NewSlice[int](10)
	assert.NotNil(t, s)
	assert.Equal(t, 0, s.Len())
}

func TestSlice_AppendAndLen(t *testing.T) {
	s := NewSlice[string](2)
	s.Append("a")
	s.Append("b")
	assert.Equal(t, 2, s.Len())
}

func TestSlice_Get(t *testing.T) {
	s := NewSlice[int](2)
	s.Append(10)
	s.Append(20)
	v, ok := s.Get(0)
	assert.True(t, ok)
	assert.Equal(t, 10, v)
	v, ok = s.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 20, v)
	v, ok = s.Get(2)
	assert.False(t, ok)
	assert.Equal(t, 0, v)
	v, ok = s.Get(-1)
	assert.False(t, ok)
	assert.Equal(t, 0, v)
}

func TestSlice_Set(t *testing.T) {
	s := NewSlice[int](2)
	s.Append(1)
	s.Append(2)
	ok := s.Set(1, 100)
	assert.True(t, ok)
	v, _ := s.Get(1)
	assert.Equal(t, 100, v)
	ok = s.Set(2, 200)
	assert.False(t, ok)
	ok = s.Set(-1, 300)
	assert.False(t, ok)
}

func TestSlice_DeleteFirst(t *testing.T) {
	s := NewSlice[int](3)
	s.Append(1)
	s.Append(2)
	s.Append(3)
	s.DeleteFirst()
	assert.Equal(t, 2, s.Len())
	v, _ := s.Get(0)
	assert.Equal(t, 2, v)
	s.DeleteFirst()
	s.DeleteFirst()
	assert.Equal(t, 0, s.Len())
	// 删除空 slice 不应 panic
	s.DeleteFirst()
	assert.Equal(t, 0, s.Len())
}

func TestSlice_DeleteLast(t *testing.T) {
	s := NewSlice[int](3)
	s.Append(1)
	s.Append(2)
	s.Append(3)
	s.DeleteLast()
	assert.Equal(t, 2, s.Len())
	v, _ := s.Get(1)
	assert.Equal(t, 2, v)
	s.DeleteLast()
	s.DeleteLast()
	assert.Equal(t, 0, s.Len())
	// 删除空 slice 不应 panic
	s.DeleteLast()
	assert.Equal(t, 0, s.Len())
}

func TestSlice_Delete(t *testing.T) {
	s := NewSlice[int](5)
	s.Append(1)
	s.Append(2)
	s.Append(3)
	s.Append(4)
	s.Delete(1)
	assert.Equal(t, 3, s.Len())
	v, _ := s.Get(1)
	assert.Equal(t, 3, v)
	s.Delete(0)
	assert.Equal(t, 2, s.Len())
	v, _ = s.Get(0)
	assert.Equal(t, 3, v)
	s.Delete(10) // 越界
	assert.Equal(t, 2, s.Len())
	s.Delete(-1) // 越界
	assert.Equal(t, 2, s.Len())
}

func TestSlice_Pop(t *testing.T) {
	s := NewSlice[string](2)
	v, ok := s.Pop()
	assert.False(t, ok)
	assert.Equal(t, "", v)
	s.Append("a")
	s.Append("b")
	v, ok = s.Pop()
	assert.True(t, ok)
	assert.Equal(t, "a", v)
	assert.Equal(t, 1, s.Len())
	v, ok = s.Pop()
	assert.True(t, ok)
	assert.Equal(t, "b", v)
	assert.Equal(t, 0, s.Len())
	v, ok = s.Pop()
	assert.False(t, ok)
}

func TestSlice_PopLast(t *testing.T) {
	s := NewSlice[string](2)
	v, ok := s.PopLast()
	assert.False(t, ok)
	assert.Equal(t, "", v)
	s.Append("a")
	s.Append("b")
	v, ok = s.PopLast()
	assert.True(t, ok)
	assert.Equal(t, "b", v)
	assert.Equal(t, 1, s.Len())
	v, ok = s.PopLast()
	assert.True(t, ok)
	assert.Equal(t, "a", v)
	assert.Equal(t, 0, s.Len())
	v, ok = s.PopLast()
	assert.False(t, ok)
}

func TestSlice_Concurrency(t *testing.T) {
	s := NewSlice[int](0)
	var wg sync.WaitGroup
	n := 1000
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(v int) {
			s.Append(v)
			wg.Done()
		}(i)
	}
	wg.Wait()
	assert.Equal(t, n, s.Len())
}

// TestSlice_RaceConditions 测试 Slice 的竞态条件
func TestSlice_RaceConditions(t *testing.T) {
	s := NewSlice[int](0)
	const numGoroutines = 50
	const operationsPerGoroutine = 500

	var wg sync.WaitGroup

	// 并发追加测试
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				s.Append(id*operationsPerGoroutine + j)
			}
		}(i)
	}

	// 并发读取测试
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				s.Len()
				if s.Len() > 0 {
					s.Get(0)
				}
			}
		}()
	}

	// 并发删除测试
	for i := 0; i < numGoroutines/4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				if s.Len() > 0 {
					s.Pop()
				}
			}
		}()
	}

	wg.Wait()
}

// TestSlice_MemoryLeak 测试 Slice 内存泄漏
func TestSlice_MemoryLeak(t *testing.T) {
	// 记录初始内存使用
	runtime.GC()
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	s := NewSlice[int](0)

	// 大量追加和弹出操作
	for i := 0; i < 10000; i++ {
		s.Append(i)
	}

	for i := 0; i < 10000; i++ {
		s.Pop()
	}

	// 强制垃圾回收
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// 检查内存增长是否合理
	memoryGrowth := int64(m2.Alloc) - int64(m1.Alloc)
	if memoryGrowth > 1024*1024 { // 1MB
		t.Errorf("Potential memory leak detected: memory growth %d bytes", memoryGrowth)
	}
}

// TestSlice_StressTestRace 压力测试
func TestSlice_StressTestRace(t *testing.T) {
	s := NewSlice[int](0)
	const numOperations = 50000

	// 并发压力测试
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations/5; j++ {
				value := id*(numOperations/5) + j
				s.Append(value)
				if j%5 == 0 && s.Len() > 0 {
					s.Pop()
				}
			}
		}(i)
	}

	wg.Wait()

	// 验证最终状态
	expectedLen := numOperations - (numOperations / 5)
	if s.Len() != expectedLen {
		t.Errorf("Expected length %d, got %d", expectedLen, s.Len())
	}
}

// TestSlice_ConcurrentModification 测试并发修改
func TestSlice_ConcurrentModification(t *testing.T) {
	s := NewSlice[int](0)
	const numGoroutines = 10
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
					s.Append(id)
					if s.Len() > 0 {
						s.Pop()
					}
				}
			}
		}(i)
	}

	// 运行一段时间
	time.Sleep(duration)
	close(done)

	// 验证 slice 仍然可用
	s.Append(999)
	val, ok := s.Pop()
	if !ok || val != 999 {
		t.Error("Slice became unusable after concurrent modification")
	}
}

// BenchmarkSlice_ConcurrentAppendPop 并发追加弹出基准测试
func BenchmarkSlice_ConcurrentAppendPop(b *testing.B) {
	s := NewSlice[int](0)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			s.Append(i)
			if s.Len() > 0 {
				s.Pop()
			}
			i++
		}
	})
}
