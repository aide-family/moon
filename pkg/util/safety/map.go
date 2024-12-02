package safety

import (
	"sync"
)

// NewMap 创建一个安全的map
func NewMap[K comparable, T any]() *Map[K, T] {
	return &Map[K, T]{
		m: new(sync.Map),
	}
}

// Map 安全的map
type Map[K comparable, T any] struct {
	m *sync.Map
}

// Get 获取map中的值
func (m *Map[K, T]) Get(key K) (T, bool) {
	v, ok := m.m.Load(key)
	if !ok {
		var zero T
		return zero, false
	}
	return v.(T), true
}

// Set 设置map中的值
func (m *Map[K, T]) Set(key K, value T) {
	m.m.Store(key, value)
}

// Delete 删除map中的值
func (m *Map[K, T]) Delete(key K) {
	m.m.Delete(key)
}

// List 获取map中的所有值
func (m *Map[K, T]) List() map[K]T {
	values := make(map[K]T)
	m.m.Range(func(key, value any) bool {
		values[key.(K)] = value.(T)
		return true
	})
	return values
}

// Clear 清空map
func (m *Map[K, T]) Clear() {
	m.m.Clear()
}
