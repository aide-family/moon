package safety_test

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/magicbox/safety"
)

// TestCopyValueCtx_Basic 测试 CopyValueCtx 的基本功能
func TestCopyValueCtx_Basic(t *testing.T) {
	// 创建一个包含值的原始 context
	type key string
	originalKey := key("test-key")
	originalValue := "test-value"
	originalCtx := context.WithValue(context.Background(), originalKey, originalValue)

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证值被保留
	value := copiedCtx.Value(originalKey)
	if value != originalValue {
		t.Errorf("CopyValueCtx() preserved value = %v, want %v", value, originalValue)
	}
}

// TestCopyValueCtx_ValuesPreserved 测试 CopyValueCtx 保留多个值
func TestCopyValueCtx_ValuesPreserved(t *testing.T) {
	type key1 string
	type key2 string
	type key3 string

	key1Value := key1("key1")
	key2Value := key2("key2")
	key3Value := key3("key3")

	value1 := "value1"
	value2 := 42
	value3 := true

	// 创建一个包含多个值的 context
	ctx := context.Background()
	ctx = context.WithValue(ctx, key1Value, value1)
	ctx = context.WithValue(ctx, key2Value, value2)
	ctx = context.WithValue(ctx, key3Value, value3)

	// 复制 context
	copiedCtx := safety.CopyValueCtx(ctx)

	// 验证所有值都被保留
	if v := copiedCtx.Value(key1Value); v != value1 {
		t.Errorf("CopyValueCtx() preserved value1 = %v, want %v", v, value1)
	}
	if v := copiedCtx.Value(key2Value); v != value2 {
		t.Errorf("CopyValueCtx() preserved value2 = %v, want %v", v, value2)
	}
	if v := copiedCtx.Value(key3Value); v != value3 {
		t.Errorf("CopyValueCtx() preserved value3 = %v, want %v", v, value3)
	}
}

// TestCopyValueCtx_Deadline 测试 CopyValueCtx 的 Deadline 方法
func TestCopyValueCtx_Deadline(t *testing.T) {
	// 创建一个带超时的 context
	deadline := time.Now().Add(5 * time.Second)
	originalCtx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证 Deadline 返回零值
	d, ok := copiedCtx.Deadline()
	if !d.IsZero() {
		t.Errorf("CopyValueCtx().Deadline() = %v, want zero time", d)
	}
	if ok {
		t.Errorf("CopyValueCtx().Deadline() ok = %v, want false", ok)
	}
}

// TestCopyValueCtx_Done 测试 CopyValueCtx 的 Done 方法
func TestCopyValueCtx_Done(t *testing.T) {
	// 创建一个可取消的 context
	originalCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证 Done 返回 nil
	done := copiedCtx.Done()
	if done != nil {
		t.Errorf("CopyValueCtx().Done() = %v, want nil", done)
	}
}

// TestCopyValueCtx_Err 测试 CopyValueCtx 的 Err 方法
func TestCopyValueCtx_Err(t *testing.T) {
	// 创建一个可取消的 context 并取消它
	originalCtx, cancel := context.WithCancel(context.Background())
	cancel() // 取消原始 context

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证 Err 返回 nil（即使原始 context 已被取消）
	err := copiedCtx.Err()
	if err != nil {
		t.Errorf("CopyValueCtx().Err() = %v, want nil", err)
	}
}

// TestCopyValueCtx_NilContext 测试 CopyValueCtx 处理 nil context
// 注意：当传入 nil context 时，访问 Value 方法可能会 panic
func TestCopyValueCtx_NilContext(t *testing.T) {
	// 测试 nil context 会 panic（这是预期的行为）
	defer func() {
		if r := recover(); r == nil {
			// 如果访问 Value 方法，可能会 panic
			// 这是正常的，因为 nil context 不能访问 Value
		}
	}()

	copiedCtx := safety.CopyValueCtx(nil)
	if copiedCtx == nil {
		t.Error("CopyValueCtx(nil) returned nil, want non-nil context")
	}

	// 验证可以调用 Deadline, Done, Err 方法而不会 panic
	_, _ = copiedCtx.Deadline()
	_ = copiedCtx.Done()
	_ = copiedCtx.Err()

	// 访问 Value 方法可能会 panic，因为内部 context 是 nil
	// 这是正常的 Go 行为
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("CopyValueCtx(nil).Value() panicked as expected: %v", r)
			}
		}()
		type key string
		_ = copiedCtx.Value(key("key"))
	}()
}

// TestCopyValueCtx_CancelledContext 测试 CopyValueCtx 处理已取消的 context
func TestCopyValueCtx_CancelledContext(t *testing.T) {
	type key string
	testKey := key("test-key")
	testValue := "test-value"

	// 创建一个已取消的 context
	originalCtx, cancel := context.WithCancel(context.Background())
	originalCtx = context.WithValue(originalCtx, testKey, testValue)
	cancel() // 取消 context

	// 等待确保 context 已被取消
	select {
	case <-originalCtx.Done():
		// Context 已被取消
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled")
	}

	// 验证原始 context 已被取消
	if err := originalCtx.Err(); err == nil {
		t.Fatal("Original context should be cancelled")
	}

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证值被保留
	value := copiedCtx.Value(testKey)
	if value != testValue {
		t.Errorf("CopyValueCtx() preserved value = %v, want %v", value, testValue)
	}

	// 验证新 context 不会被取消
	if err := copiedCtx.Err(); err != nil {
		t.Errorf("CopyValueCtx().Err() = %v, want nil", err)
	}

	// 验证 Done 返回 nil
	if done := copiedCtx.Done(); done != nil {
		t.Errorf("CopyValueCtx().Done() = %v, want nil", done)
	}
}

// TestCopyValueCtx_TimeoutContext 测试 CopyValueCtx 处理带超时的 context
func TestCopyValueCtx_TimeoutContext(t *testing.T) {
	type key string
	testKey := key("test-key")
	testValue := "test-value"

	// 创建一个带超时的 context
	originalCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	originalCtx = context.WithValue(originalCtx, testKey, testValue)

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证值被保留
	value := copiedCtx.Value(testKey)
	if value != testValue {
		t.Errorf("CopyValueCtx() preserved value = %v, want %v", value, testValue)
	}

	// 验证 Deadline 返回零值
	d, ok := copiedCtx.Deadline()
	if !d.IsZero() {
		t.Errorf("CopyValueCtx().Deadline() = %v, want zero time", d)
	}
	if ok {
		t.Errorf("CopyValueCtx().Deadline() ok = %v, want false", ok)
	}

	// 等待原始 context 超时
	time.Sleep(150 * time.Millisecond)

	// 验证原始 context 已超时
	select {
	case <-originalCtx.Done():
		// Context 已超时
	default:
		t.Fatal("Original context should be timed out")
	}

	// 验证新 context 不会超时
	if err := copiedCtx.Err(); err != nil {
		t.Errorf("CopyValueCtx().Err() = %v, want nil", err)
	}

	// 验证 Done 返回 nil
	if done := copiedCtx.Done(); done != nil {
		t.Errorf("CopyValueCtx().Done() = %v, want nil", done)
	}
}

// TestCopyValueCtx_BackgroundContext 测试 CopyValueCtx 处理 Background context
func TestCopyValueCtx_BackgroundContext(t *testing.T) {
	// 使用 Background context
	originalCtx := context.Background()

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证不会 panic
	if copiedCtx == nil {
		t.Error("CopyValueCtx(context.Background()) returned nil, want non-nil context")
	}

	// 验证 Deadline 返回零值
	d, ok := copiedCtx.Deadline()
	if !d.IsZero() {
		t.Errorf("CopyValueCtx().Deadline() = %v, want zero time", d)
	}
	if ok {
		t.Errorf("CopyValueCtx().Deadline() ok = %v, want false", ok)
	}

	// 验证 Done 返回 nil
	if done := copiedCtx.Done(); done != nil {
		t.Errorf("CopyValueCtx().Done() = %v, want nil", done)
	}

	// 验证 Err 返回 nil
	if err := copiedCtx.Err(); err != nil {
		t.Errorf("CopyValueCtx().Err() = %v, want nil", err)
	}
}

// TestCopyValueCtx_TODOContext 测试 CopyValueCtx 处理 TODO context
func TestCopyValueCtx_TODOContext(t *testing.T) {
	// 使用 TODO context
	originalCtx := context.TODO()

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 验证不会 panic
	if copiedCtx == nil {
		t.Error("CopyValueCtx(context.TODO()) returned nil, want non-nil context")
	}

	// 验证 Deadline 返回零值
	d, ok := copiedCtx.Deadline()
	if !d.IsZero() {
		t.Errorf("CopyValueCtx().Deadline() = %v, want zero time", d)
	}
	if ok {
		t.Errorf("CopyValueCtx().Deadline() ok = %v, want false", ok)
	}

	// 验证 Done 返回 nil
	if done := copiedCtx.Done(); done != nil {
		t.Errorf("CopyValueCtx().Done() = %v, want nil", done)
	}

	// 验证 Err 返回 nil
	if err := copiedCtx.Err(); err != nil {
		t.Errorf("CopyValueCtx().Err() = %v, want nil", err)
	}
}

// TestCopyValueCtx_NestedValues 测试 CopyValueCtx 处理嵌套的值
func TestCopyValueCtx_NestedValues(t *testing.T) {
	type key1 string
	type key2 string
	type key3 string

	// 创建嵌套的 context
	ctx := context.Background()
	ctx = context.WithValue(ctx, key1("key1"), "value1")
	ctx = context.WithValue(ctx, key2("key2"), "value2")
	ctx = context.WithValue(ctx, key3("key3"), "value3")

	// 再创建一层嵌套
	ctx = context.WithValue(ctx, key1("key1-2"), "value1-2")

	// 复制 context
	copiedCtx := safety.CopyValueCtx(ctx)

	// 验证所有值都被保留（包括嵌套的值）
	if v := copiedCtx.Value(key1("key1")); v != "value1" {
		t.Errorf("CopyValueCtx() preserved nested value1 = %v, want %v", v, "value1")
	}
	if v := copiedCtx.Value(key2("key2")); v != "value2" {
		t.Errorf("CopyValueCtx() preserved nested value2 = %v, want %v", v, "value2")
	}
	if v := copiedCtx.Value(key3("key3")); v != "value3" {
		t.Errorf("CopyValueCtx() preserved nested value3 = %v, want %v", v, "value3")
	}
	if v := copiedCtx.Value(key1("key1-2")); v != "value1-2" {
		t.Errorf("CopyValueCtx() preserved nested value1-2 = %v, want %v", v, "value1-2")
	}
}

// TestCopyValueCtx_DifferentValueTypes 测试 CopyValueCtx 处理不同类型的值
func TestCopyValueCtx_DifferentValueTypes(t *testing.T) {
	type key string

	// 创建包含不同类型值的 context
	ctx := context.Background()
	ctx = context.WithValue(ctx, key("string"), "hello")
	ctx = context.WithValue(ctx, key("int"), 42)
	ctx = context.WithValue(ctx, key("bool"), true)
	ctx = context.WithValue(ctx, key("float"), 3.14)
	ctx = context.WithValue(ctx, key("slice"), []int{1, 2, 3})
	ctx = context.WithValue(ctx, key("map"), map[string]int{"a": 1, "b": 2})

	// 复制 context
	copiedCtx := safety.CopyValueCtx(ctx)

	// 验证所有值都被保留
	if v := copiedCtx.Value(key("string")); v != "hello" {
		t.Errorf("CopyValueCtx() preserved string value = %v, want %v", v, "hello")
	}
	if v := copiedCtx.Value(key("int")); v != 42 {
		t.Errorf("CopyValueCtx() preserved int value = %v, want %v", v, 42)
	}
	if v := copiedCtx.Value(key("bool")); v != true {
		t.Errorf("CopyValueCtx() preserved bool value = %v, want %v", v, true)
	}
	if v := copiedCtx.Value(key("float")); v != 3.14 {
		t.Errorf("CopyValueCtx() preserved float value = %v, want %v", v, 3.14)
	}
	if v := copiedCtx.Value(key("slice")); v == nil {
		t.Errorf("CopyValueCtx() preserved slice value = %v, want non-nil", v)
	}
	if v := copiedCtx.Value(key("map")); v == nil {
		t.Errorf("CopyValueCtx() preserved map value = %v, want non-nil", v)
	}
}

// TestCopyValueCtx_NoPanic 测试 CopyValueCtx 不会 panic
func TestCopyValueCtx_NoPanic(t *testing.T) {
	// 测试各种情况不会 panic
	testCases := []struct {
		name        string
		ctx         context.Context
		shouldPanic bool
	}{
		{
			name:        "nil context",
			ctx:         nil,
			shouldPanic: true, // nil context 访问 Value 会 panic
		},
		{
			name:        "Background context",
			ctx:         context.Background(),
			shouldPanic: false,
		},
		{
			name:        "TODO context",
			ctx:         context.TODO(),
			shouldPanic: false,
		},
		{
			name: "Cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			shouldPanic: false,
		},
		{
			name: "Timeout context",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				return ctx
			}(),
			shouldPanic: false,
		},
		{
			name: "Deadline context",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(100*time.Millisecond))
				defer cancel()
				return ctx
			}(),
			shouldPanic: false,
		},
		{
			name: "Context with values",
			ctx: func() context.Context {
				type key string
				return context.WithValue(context.Background(), key("key"), "value")
			}(),
			shouldPanic: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			copiedCtx := safety.CopyValueCtx(tt.ctx)
			if copiedCtx == nil {
				t.Error("CopyValueCtx() returned nil")
				return
			}

			// 验证可以调用 Deadline, Done, Err 方法而不会 panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("CopyValueCtx().Deadline/Done/Err() panicked: %v", r)
					}
				}()
				_, _ = copiedCtx.Deadline()
				_ = copiedCtx.Done()
				_ = copiedCtx.Err()
			}()

			// 访问 Value 方法可能会 panic（如果 ctx 是 nil）
			func() {
				defer func() {
					if r := recover(); r != nil {
						if tt.shouldPanic {
							t.Logf("CopyValueCtx().Value() panicked as expected: %v", r)
						} else {
							t.Errorf("CopyValueCtx().Value() panicked unexpectedly: %v", r)
						}
					} else if tt.shouldPanic {
						t.Error("CopyValueCtx().Value() should have panicked but didn't")
					}
				}()
				type key string
				_ = copiedCtx.Value(key("key"))
			}()
		})
	}
}

// TestCopyValueCtx_Isolation 测试 CopyValueCtx 创建的 context 与原 context 隔离
func TestCopyValueCtx_Isolation(t *testing.T) {
	type key string
	testKey := key("test-key")

	// 创建一个可取消的 context
	originalCtx, cancel := context.WithCancel(context.Background())
	originalCtx = context.WithValue(originalCtx, testKey, "original-value")

	// 复制 context
	copiedCtx := safety.CopyValueCtx(originalCtx)

	// 取消原始 context
	cancel()

	// 等待确保原始 context 已被取消
	select {
	case <-originalCtx.Done():
		// Context 已被取消
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Original context should be cancelled")
	}

	// 验证原始 context 已被取消
	if err := originalCtx.Err(); err == nil {
		t.Fatal("Original context should be cancelled")
	}

	// 验证新 context 不会被取消
	if err := copiedCtx.Err(); err != nil {
		t.Errorf("CopyValueCtx().Err() = %v, want nil (context should be isolated)", err)
	}

	// 验证 Done 返回 nil
	if done := copiedCtx.Done(); done != nil {
		t.Errorf("CopyValueCtx().Done() = %v, want nil (context should be isolated)", done)
	}

	// 验证值仍然可以访问
	if v := copiedCtx.Value(testKey); v != "original-value" {
		t.Errorf("CopyValueCtx() preserved value = %v, want %v", v, "original-value")
	}
}

// TestCopyValueCtx_ValueNotFound 测试 CopyValueCtx 处理不存在的值
func TestCopyValueCtx_ValueNotFound(t *testing.T) {
	type key string

	// 创建一个不包含值的 context
	ctx := context.Background()

	// 复制 context
	copiedCtx := safety.CopyValueCtx(ctx)

	// 验证不存在的值返回 nil
	value := copiedCtx.Value(key("non-existent-key"))
	if value != nil {
		t.Errorf("CopyValueCtx().Value(non-existent-key) = %v, want nil", value)
	}
}

// TestCopyValueCtx_ConcurrentAccess 测试 CopyValueCtx 的并发访问
func TestCopyValueCtx_ConcurrentAccess(t *testing.T) {
	type key string
	testKey := key("test-key")
	testValue := "test-value"

	// 创建一个包含值的 context
	ctx := context.WithValue(context.Background(), testKey, testValue)

	// 复制 context
	copiedCtx := safety.CopyValueCtx(ctx)

	// 并发访问复制的 context
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Concurrent access panicked: %v", r)
				}
			}()

			// 访问值
			_ = copiedCtx.Value(testKey)

			// 访问 Deadline
			_, _ = copiedCtx.Deadline()

			// 访问 Done
			_ = copiedCtx.Done()

			// 访问 Err
			_ = copiedCtx.Err()

			done <- true
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestCopyValueCtx_ChainedCopy 测试链式复制 context
func TestCopyValueCtx_ChainedCopy(t *testing.T) {
	type key string
	testKey := key("test-key")
	testValue := "test-value"

	// 创建原始 context
	ctx := context.WithValue(context.Background(), testKey, testValue)

	// 第一次复制
	copiedCtx1 := safety.CopyValueCtx(ctx)

	// 第二次复制（从第一次复制的结果）
	copiedCtx2 := safety.CopyValueCtx(copiedCtx1)

	// 验证值仍然可以访问
	if v := copiedCtx2.Value(testKey); v != testValue {
		t.Errorf("ChainedCopyValueCtx() preserved value = %v, want %v", v, testValue)
	}

	// 验证 Deadline 返回零值
	d, ok := copiedCtx2.Deadline()
	if !d.IsZero() {
		t.Errorf("ChainedCopyValueCtx().Deadline() = %v, want zero time", d)
	}
	if ok {
		t.Errorf("ChainedCopyValueCtx().Deadline() ok = %v, want false", ok)
	}

	// 验证 Done 返回 nil
	if done := copiedCtx2.Done(); done != nil {
		t.Errorf("ChainedCopyValueCtx().Done() = %v, want nil", done)
	}

	// 验证 Err 返回 nil
	if err := copiedCtx2.Err(); err != nil {
		t.Errorf("ChainedCopyValueCtx().Err() = %v, want nil", err)
	}
}

// BenchmarkCopyValueCtx 基准测试 CopyValueCtx 函数
func BenchmarkCopyValueCtx(b *testing.B) {
	type key string
	ctx := context.WithValue(context.Background(), key("key"), "value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = safety.CopyValueCtx(ctx)
	}
}

// BenchmarkCopyValueCtx_WithMultipleValues 基准测试 CopyValueCtx 函数（多个值）
func BenchmarkCopyValueCtx_WithMultipleValues(b *testing.B) {
	type key1 string
	type key2 string
	type key3 string
	type key4 string
	type key5 string
	ctx := context.Background()
	ctx = context.WithValue(ctx, key1("key1"), "value1")
	ctx = context.WithValue(ctx, key2("key2"), "value2")
	ctx = context.WithValue(ctx, key3("key3"), "value3")
	ctx = context.WithValue(ctx, key4("key4"), "value4")
	ctx = context.WithValue(ctx, key5("key5"), "value5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = safety.CopyValueCtx(ctx)
	}
}

// BenchmarkCopyValueCtx_ValueAccess 基准测试访问复制的 context 的值
func BenchmarkCopyValueCtx_ValueAccess(b *testing.B) {
	type key string
	ctx := context.WithValue(context.Background(), key("key"), "value")
	copiedCtx := safety.CopyValueCtx(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = copiedCtx.Value(key("key"))
	}
}

// BenchmarkCopyValueCtx_Deadline 基准测试访问复制的 context 的 Deadline
func BenchmarkCopyValueCtx_Deadline(b *testing.B) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Hour))
	defer cancel()
	copiedCtx := safety.CopyValueCtx(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = copiedCtx.Deadline()
	}
}

// BenchmarkCopyValueCtx_Done 基准测试访问复制的 context 的 Done
func BenchmarkCopyValueCtx_Done(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	copiedCtx := safety.CopyValueCtx(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = copiedCtx.Done()
	}
}

// BenchmarkCopyValueCtx_Err 基准测试访问复制的 context 的 Err
func BenchmarkCopyValueCtx_Err(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	copiedCtx := safety.CopyValueCtx(ctx)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = copiedCtx.Err()
	}
}
