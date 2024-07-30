package types

// ToMap 切片转map
func ToMap[T any, R comparable](list []T, f func(T) R) map[R]T {
	m := make(map[R]T, len(list))
	for _, v := range list {
		m[f(v)] = v
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
