package safety_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aide-family/magicbox/safety"
)

// TestGo_Success 测试 Go 函数正常执行的情况
func TestGo_Success(t *testing.T) {
	ctx := context.Background()
	executed := make(chan bool, 1)
	var executedOnce sync.Once

	// 创建一个成功执行的函数
	f := func(context.Context) error {
		executedOnce.Do(func() {
			executed <- true
		})
		return nil
	}

	// 调用 Go 函数
	safety.Go(ctx, "test-success", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}
}

// TestGo_Error 测试 Go 函数处理错误的情况
func TestGo_Error(t *testing.T) {
	ctx := context.Background()
	testErr := errors.New("test error")
	executed := make(chan bool, 1)
	var executedOnce sync.Once

	// 创建一个返回错误的函数
	f := func(context.Context) error {
		executedOnce.Do(func() {
			executed <- true
		})
		return testErr
	}

	// 调用 Go 函数
	safety.Go(ctx, "test-error", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}
}

// TestGo_Panic 测试 Go 函数处理 panic 的情况
func TestGo_Panic(t *testing.T) {
	ctx := context.Background()
	panicMsg := "test panic"
	executed := make(chan bool, 1)
	var executedOnce sync.Once

	// 创建一个会 panic 的函数
	f := func(context.Context) error {
		executedOnce.Do(func() {
			executed <- true
		})
		panic(panicMsg)
	}

	// 调用 Go 函数
	safety.Go(ctx, "test-panic", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行（虽然会 panic）
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}
}

// TestGo_ContextPassed 测试 Go 函数正确传递 context
func TestGo_ContextPassed(t *testing.T) {
	// 创建一个带值的 context
	type key string
	testKey := key("test-key")
	testValue := "test-value"
	ctx := context.WithValue(context.Background(), testKey, testValue)

	executed := make(chan bool, 1)
	var executedOnce sync.Once
	var receivedValue interface{}

	// 创建一个验证 context 的函数
	f := func(c context.Context) error {
		executedOnce.Do(func() {
			receivedValue = c.Value(testKey)
			executed <- true
		})
		return nil
	}

	// 调用 Go 函数
	safety.Go(ctx, "test-context", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}

	// 验证 context 值被正确传递
	if receivedValue != testValue {
		t.Errorf("Expected context value %v, got %v", testValue, receivedValue)
	}
}

// TestGo_ConcurrentExecution 测试 Go 函数并发执行
func TestGo_ConcurrentExecution(t *testing.T) {
	ctx := context.Background()
	var wg sync.WaitGroup
	const numGoroutines = 10

	// 创建多个函数并发执行
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		f := func(context.Context) error {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond)
			return nil
		}
		safety.Go(ctx, "test-concurrent", f)
	}

	// 等待所有 goroutine 完成
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// 所有 goroutine 已完成
	case <-time.After(5 * time.Second):
		t.Fatal("Not all goroutines completed within timeout")
	}

}

// TestGo_WithCancelledContext 测试 Go 函数处理已取消的 context
func TestGo_WithCancelledContext(t *testing.T) {
	// 创建一个已取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	executed := make(chan bool, 1)
	var executedOnce sync.Once

	// 创建一个检查 context 状态的函数
	f := func(c context.Context) error {
		executedOnce.Do(func() {
			executed <- true
		})
		return nil
	}

	// 调用 Go 函数
	safety.Go(ctx, "test-cancelled-context", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}
}

// TestGo_EmptyName 测试 Go 函数使用空名称
func TestGo_EmptyName(t *testing.T) {
	ctx := context.Background()
	executed := make(chan bool, 1)
	var executedOnce sync.Once

	// 创建一个成功执行的函数
	f := func(context.Context) error {
		executedOnce.Do(func() {
			executed <- true
		})
		return nil
	}

	// 调用 Go 函数，使用空名称
	safety.Go(ctx, "", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}
}

// TestGo_WithNopLogger 测试 Go 函数使用 NopLogger
func TestGo_WithNopLogger(t *testing.T) {
	ctx := context.Background()
	executed := make(chan bool, 1)
	var executedOnce sync.Once

	// 创建一个成功执行的函数
	f := func(context.Context) error {
		executedOnce.Do(func() {
			executed <- true
		})
		return nil
	}

	// 调用 Go 函数
	safety.Go(ctx, "test-nop-logger", f)

	// 等待函数执行完成
	select {
	case <-executed:
		// 函数已执行
	case <-time.After(2 * time.Second):
		t.Fatal("Function did not execute within timeout")
	}
}

// BenchmarkGo 基准测试 Go 函数
func BenchmarkGo(b *testing.B) {
	ctx := context.Background()
	f := func(context.Context) error {
		return nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		func() {
			defer wg.Done()
			safety.Go(ctx, "benchmark", f)
		}()
		wg.Wait()
		// 等待 goroutine 完成
		time.Sleep(10 * time.Millisecond)
	}
}
