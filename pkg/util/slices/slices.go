package slices

// To 转换为其他类型
func To[T, R any](list []T, f func(T) R) []R {
	rs := make([]R, 0, len(list))
	for _, v := range list {
		rs = append(rs, f(v))
	}
	return rs
}
