package types

import "encoding/json"

// ToMap 切片转map
func ToMap[T any, R comparable](list []T, f func(T) R) map[R]T {
	m := make(map[R]T, len(list))
	for _, v := range list {
		m[f(v)] = v
	}
	return m
}

// ToMapSlice 切片转map value为切片
func ToMapSlice[T any, R comparable](list []T, f func(T) R) map[R][]T {
	m := make(map[R][]T)
	for _, v := range list {
		key := f(v)
		m[key] = append(m[key], v)
	}
	return m
}

// MapsMerge 合并多个map
func MapsMerge[K comparable, V any](ms ...map[K]V) map[K]V {
	m := make(map[K]V)
	for _, v := range ms {
		for k, vv := range v {
			m[k] = vv
		}
	}
	return m
}

// JSONToMap json转map
func JSONToMap(jsonStr string) (map[string]interface{}, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
