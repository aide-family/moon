package types

// SliceTo 将slice转换为指定类型
func SliceTo[T, R any](s []T, call func(T) R) []R {
	if IsNil(s) || len(s) == 0 {
		return nil
	}
	list := make([]R, 0, len(s))
	for _, v := range s {
		list = append(list, call(v))
	}
	return list
}

// SliceToWithFilter 将slice转换为指定类型，并过滤掉指定的值
func SliceToWithFilter[T, R any](s []T, call func(T) (R, bool)) []R {
	if IsNil(s) || len(s) == 0 {
		return nil
	}
	list := make([]R, 0, len(s))
	for _, v := range s {
		if r, ok := call(v); ok {
			list = append(list, r)
		}
	}
	return list
}
