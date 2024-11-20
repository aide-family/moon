package runtime

import "context"

// Reader 读取器
type Reader interface {
	// Get 获取数据
	Get(ctx context.Context, key string, out any) error
	// List 获取数据列表
	List(ctx context.Context, out any) error
}

// Writer 写入器
type Writer interface {
	// Create 创建数据
	Create(ctx context.Context, object any) error
	// Update 更新数据
	Update(ctx context.Context, object any) error
	// Delete 删除数据
	Delete(ctx context.Context, object any) error
}
