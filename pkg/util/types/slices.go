package types

import (
	"sort"
)

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

// MergeSlice 合并切片
func MergeSlice[T any](slices ...[]T) []T {
	if len(slices) == 0 {
		return nil
	}
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	list := make([]T, 0, totalLen)
	for _, s := range slices {
		list = append(list, s...)
	}
	return list
}

// MergeSliceWithUnique 合并切片，并去重
func MergeSliceWithUnique[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return nil
	}
	var totalLen int
	for _, s := range slices {
		totalLen += len(s)
	}
	list := make([]T, 0, totalLen)
	for _, s := range slices {
		list = append(list, s...)
	}
	return SliceUnique(list)
}

// SliceUnique 切片去重
func SliceUnique[T comparable](s []T) []T {
	if IsNil(s) || len(s) == 0 {
		return nil
	}
	m := make(map[T]struct{}, len(s))
	for _, v := range s {
		m[v] = struct{}{}
	}
	list := make([]T, 0, len(m))
	for k := range m {
		list = append(list, k)
	}
	return list
}

type compare interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

// SlicesIntersection 求交集
func SlicesIntersection[T compare](slice1, slice2 []T) []T {
	// 排序切片
	sort.SliceIsSorted(slice1, func(i, j int) bool {
		return slice1[i] < slice1[j]
	})
	sort.Slice(slice2, func(i, j int) bool {
		return slice2[i] < slice2[j]
	})

	var intersection []T
	i, j := 0, 0

	// 双指针遍历
	for i < len(slice1) && j < len(slice2) {
		if slice1[i] == slice2[j] {
			// 如果元素相同，则加入结果切片，并且两个指针都向前移动
			intersection = append(intersection, slice1[i])
			i++
			j++
		} else if slice1[i] < slice2[j] {
			// 如果slice1[i]小于slice2[j]，则i指针向前
			i++
		} else {
			// 否则j指针向前
			j++
		}
	}

	return intersection
}

// To 转换为其他类型
func To[T, R any](list []T, f func(T) R) []R {
	rs := make([]R, 0, len(list))
	for _, v := range list {
		rs = append(rs, f(v))
	}
	return rs
}

// ToFilter 过滤
func ToFilter[T, R any](list []T, f func(T) (R, bool)) []R {
	rs := make([]R, 0, len(list))
	for _, v := range list {
		okVal, ok := f(v)
		if ok {
			rs = append(rs, okVal)
		}
	}
	return rs
}

// Index 查找元素的索引
func Index[T comparable](list []T, v T) int {
	for i, item := range list {
		if item == v {
			return i
		}
	}
	return -1
}

// IndexOf 查找元素的索引
func IndexOf[T any](list []T, f func(T) bool) int {
	for i, item := range list {
		if f(item) {
			return i
		}
	}
	return -1
}

// Contains 是否包含
func Contains[T comparable](list []T, v T) bool {
	return Index(list, v) != -1
}

// ContainsOf 是否包含
func ContainsOf[T any](list []T, f func(T) bool) bool {
	return IndexOf(list, f) != -1
}

// Filter 过滤
func Filter[T any](list []T, f func(T) bool) []T {
	rs := make([]T, 0, len(list))
	for _, v := range list {
		if f(v) {
			rs = append(rs, v)
		}
	}
	return rs
}
