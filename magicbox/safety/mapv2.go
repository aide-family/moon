package safety

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"sync"
)

// SyncMap is a thread-safe map that wraps sync.Map.
// The mu mutex is only used to protect the m pointer during replacement (e.g., Clear, Unmarshal).
// All other operations rely on sync.Map's internal concurrency safety.
type SyncMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  *sync.Map
}

var _ encoding.BinaryMarshaler = (*SyncMap[string, any])(nil)
var _ encoding.BinaryUnmarshaler = (*SyncMap[string, any])(nil)
var _ sql.Scanner = (*SyncMap[string, any])(nil)
var _ driver.Valuer = (*SyncMap[string, any])(nil)

func NewSyncMap[K comparable, V any](m map[K]V) *SyncMap[K, V] {
	sm := &sync.Map{}
	for k, v := range m {
		sm.Store(k, v)
	}
	return &SyncMap[K, V]{m: sm}
}

func (m *SyncMap[K, V]) Get(k K) (V, bool) {
	v, ok := m.m.Load(k)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}

func (m *SyncMap[K, V]) Set(k K, v V) *SyncMap[K, V] {
	m.m.Store(k, v)
	return m
}

func (m *SyncMap[K, V]) Append(ms ...map[K]V) *SyncMap[K, V] {
	for _, mm := range ms {
		for k, v := range mm {
			m.m.Store(k, v)
		}
	}
	return m
}

func (m *SyncMap[K, V]) Delete(k K) *SyncMap[K, V] {
	m.m.Delete(k)
	return m
}

func (m *SyncMap[K, V]) DeleteFunc(f func(k K, v V) bool) *SyncMap[K, V] {
	m.m.Range(func(k, v any) bool {
		if f(k.(K), v.(V)) {
			m.m.Delete(k)
		}
		return true
	})
	return m
}

func (m *SyncMap[K, V]) Range(f func(k K, v V) bool) {
	m.m.Range(func(k, v any) bool {
		return f(k.(K), v.(V))
	})
}

func (m *SyncMap[K, V]) Len() int {
	var count int
	m.m.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

func (m *SyncMap[K, V]) Keys() []K {
	keys := make([]K, 0, m.Len())
	m.m.Range(func(k, v any) bool {
		keys = append(keys, k.(K))
		return true
	})
	return keys
}

func (m *SyncMap[K, V]) Values() []V {
	values := make([]V, 0, m.Len())
	m.m.Range(func(k, v any) bool {
		values = append(values, v.(V))
		return true
	})
	return values
}

func (m *SyncMap[K, V]) Clear() *SyncMap[K, V] {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m = &sync.Map{}
	return m
}

func (m *SyncMap[K, V]) Clone() *SyncMap[K, V] {
	newMap := &sync.Map{}
	m.m.Range(func(k, v any) bool {
		newMap.Store(k, v)
		return true
	})
	return &SyncMap[K, V]{m: newMap}
}

func (m *SyncMap[K, V]) Map() map[K]V {
	newMap := make(map[K]V, m.Len())
	m.m.Range(func(k, v any) bool {
		newMap[k.(K)] = v.(V)
		return true
	})
	return newMap
}

func (m *SyncMap[K, V]) UnmarshalBinary(data []byte) error {
	var newMap map[K]V
	if err := json.Unmarshal(data, &newMap); err != nil {
		return err
	}
	newSyncMap := &sync.Map{}
	for k, v := range newMap {
		newSyncMap.Store(k, v)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m = newSyncMap
	return nil
}

func (m *SyncMap[K, V]) MarshalBinary() ([]byte, error) {
	return json.Marshal(m.Map())
}

func (m *SyncMap[K, V]) String() string {
	bs, _ := json.Marshal(m.Map())
	return string(bs)
}

func (m *SyncMap[K, V]) Value() (driver.Value, error) {
	return json.Marshal(m.Map())
}

func (m *SyncMap[K, V]) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		newMap := m.Map()
		if err := json.Unmarshal(src, &newMap); err != nil {
			return err
		}
		m.m = &sync.Map{}
		for k, v := range newMap {
			m.m.Store(k, v)
		}
		return nil
	case string:
		newMap := m.Map()
		if err := json.Unmarshal([]byte(src), &newMap); err != nil {
			return err
		}
		m.m = &sync.Map{}
		for k, v := range newMap {
			m.m.Store(k, v)
		}
		return nil
	case nil:
		m.m = &sync.Map{}
		return nil
	default:
		return fmt.Errorf("unsupported type: %T, expected []byte or string", src)
	}
}
