package slices

import (
	"encoding"
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/validate"
)

// FindByValue find slice by value, return value and ok
func FindByValue[T any, R comparable](s []T, val R, f func(v T) R) (r T, ok bool) {
	for _, v := range s {
		if f(v) == val {
			return v, true
		}
	}
	return
}

// Map map slice
func Map[T any, R any](s []T, f func(v T) R) []R {
	r := make([]R, 0, len(s))
	for _, v := range s {
		r = append(r, f(v))
	}
	return r
}

// MapFilter map slice and filter
func MapFilter[T any, R any](s []T, f func(v T) (R, bool)) []R {
	r := make([]R, 0, len(s))
	for _, v := range s {
		if v, ok := f(v); ok {
			r = append(r, v)
		}
	}
	return r
}

// Unique unique slice
func Unique[T comparable](s []T) []T {
	m := make(map[T]struct{}, len(s))
	r := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := m[v]; !ok {
			m[v] = struct{}{}
			r = append(r, v)
		}
	}
	return r
}

// UniqueWithFunc unique slice with func
func UniqueWithFunc[T any, K comparable](s []T, f func(v T) K) []T {
	m := make(map[K]struct{}, len(s))
	r := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := m[f(v)]; !ok {
			m[f(v)] = struct{}{}
			r = append(r, v)
		}
	}
	return r
}

func ToMap[T any, K comparable](s []T, f func(v T) K) map[K]T {
	m := make(map[K]T)
	for _, v := range s {
		m[f(v)] = v
	}
	return m
}

func UnmarshalBinary[T any](data []any, src *[]*T) error {
	if validate.IsNil(src) {
		return nil
	}
	list := make([][]byte, 0, len(data))
	for _, v := range data {
		switch val := v.(type) {
		case []byte:
			list = append(list, val)
		case string:
			list = append(list, []byte(val))
		}
	}
	for _, v := range list {
		var item T
		switch item := any(item).(type) {
		case encoding.BinaryUnmarshaler:
			if err := item.UnmarshalBinary(v); err != nil {
				return err
			}
		default:
			if err := json.Unmarshal(v, &item); err != nil {
				return err
			}
		}
		*src = append(*src, &item)
	}
	return nil
}

func GroupBy[T any, K comparable](s []T, f func(v T) K) map[K][]T {
	m := make(map[K][]T)
	for _, v := range s {
		if _, ok := m[f(v)]; !ok {
			m[f(v)] = make([]T, 0, len(s))
		}
		m[f(v)] = append(m[f(v)], v)
	}
	return m
}

func Contains[T comparable](s []T, v T) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}
