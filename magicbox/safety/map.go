// Package safety is a package that provides a map with safety.
package safety

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"maps"
	"sync"
)

var (
	_ encoding.BinaryMarshaler   = (*Map[string, any])(nil)
	_ encoding.BinaryUnmarshaler = (*Map[string, any])(nil)
	_ sql.Scanner                = (*Map[string, any])(nil)
	_ driver.Valuer              = (*Map[string, any])(nil)
)

type Map[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// Value implements driver.Valuer.
func (m *Map[K, V]) Value() (driver.Value, error) {
	return json.Marshal(m.m)
}

// Scan implements sql.Scanner.
func (m *Map[K, V]) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, &m.m)
	case string:
		return json.Unmarshal([]byte(src), &m.m)
	case nil:
		m.m = make(map[K]V)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T, expected []byte or string", src)
	}
}

func NewMap[K comparable, V any](m map[K]V) *Map[K, V] {
	return &Map[K, V]{m: maps.Clone(m)}
}

func (m *Map[K, V]) Get(k K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[k]
	return v, ok
}

func (m *Map[K, V]) Set(k K, v V) *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[k] = v
	return m
}

func (m *Map[K, V]) Append(ms ...map[K]V) *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, mm := range ms {
		maps.Insert(m.m, maps.All(mm))
	}
	return m
}

func (m *Map[K, V]) Delete(k K) *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, k)
	return m
}

func (m *Map[K, V]) DeleteFunc(f func(k K, v V) bool) *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	maps.DeleteFunc(m.m, f)
	return m
}

func (m *Map[K, V]) Range(f func(k K, v V) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.m {
		if !f(k, v) {
			break
		}
	}
}

func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}

func (m *Map[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	keys := make([]K, 0, len(m.m))
	for k := range m.m {
		keys = append(keys, k)
	}
	return keys
}

func (m *Map[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]V, 0, len(m.m))
	for _, v := range m.m {
		values = append(values, v)
	}
	return values
}

func (m *Map[K, V]) Clear() *Map[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m = make(map[K]V)
	return m
}

func (m *Map[K, V]) Clone() *Map[K, V] {
	newMap := make(map[K]V)
	m.mu.RLock()
	defer m.mu.RUnlock()
	maps.Copy(newMap, m.m)
	return NewMap(newMap)
}

func (m *Map[K, V]) Map() map[K]V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return maps.Clone(m.m)
}

func (m *Map[K, V]) UnmarshalBinary(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return json.Unmarshal(data, &m.m)
}

func (m *Map[K, V]) MarshalBinary() (data []byte, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return json.Marshal(m.m)
}

func (m *Map[K, V]) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bs, _ := json.Marshal(m.m)
	return string(bs)
}
