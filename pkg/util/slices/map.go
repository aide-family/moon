package slices

func ToMap[T any, R comparable](list []T, f func(T) R) map[R]T {
	m := make(map[R]T)
	for _, v := range list {
		m[f(v)] = v
	}
	return m
}
