package kv

import (
	"sort"
	"strings"
)

func NewStringMap(ms ...map[string]string) StringMap {
	return New[string, string](ms...)
}

type StringMap = Map[string, string]

func SortString(m map[string]string) string {
	keys := StringMap(m).Keys()
	sort.Strings(keys)
	var buf strings.Builder
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(m[k])
	}
	return buf.String()
}
