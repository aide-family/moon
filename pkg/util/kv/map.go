package kv

import (
	"database/sql/driver"
	"encoding/json"
)

func New[K comparable, V any](ms ...map[K]V) Map[K, V] {
	m := make(Map[K, V], len(ms))
	for _, v1 := range ms {
		for k, v2 := range v1 {
			m[k] = v2
		}
	}
	return m
}

type Map[K comparable, V any] map[K]V

func (m Map[K, V]) String() string {
	bs, _ := m.MarshalBinary()
	return string(bs)
}

func (m Map[K, V]) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m Map[K, V]) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &m)
}

func (m Map[K, V]) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *Map[K, V]) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

func (m Map[K, V]) Get(key K) (V, bool) {
	v, ok := m[key]
	return v, ok
}

func (m Map[K, V]) GetX(key K) V {
	v, ok := m[key]
	if !ok {
		var zero V
		return zero
	}
	return v
}

func (m Map[K, V]) Set(key K, value V) {
	m[key] = value
}

func (m Map[K, V]) Del(key K) {
	delete(m, key)
}

func (m Map[K, V]) Keys() []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (m Map[K, V]) Values() []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func (m Map[K, V]) Len() int {
	return len(m)
}

func (m Map[K, V]) ToMap() map[K]V {
	return m
}

func (m Map[K, V]) Copy() Map[K, V] {
	return New(m)
}
