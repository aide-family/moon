package types

import (
	"strconv"
	"strings"
)

// TextIsNull 判断字符串是否为空
func TextIsNull(text string) bool {
	return len(text) == 0
}

// StrToUint32 将字符串转换为uint32类型
func StrToUint32(s string) (uint32, error) {
	value, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	// Cast to uint32 after parsing
	return uint32(value), nil
}

// StrToUint64 将字符串转换为uint64类型
func StrToUint64(s string) (uint64, error) {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	// Cast to uint64 after parsing
	return uint64(value), nil
}

// StrToInt64 将字符串转换为int64类型
func StrToInt64(s string) (int64, error) {
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// TextTrim 移除字符串首尾空格
func TextTrim(text string) string {
	return strings.TrimSpace(text)
}

// TextTrimIsNull 判断字符串首尾空格后是否为空
func TextTrimIsNull(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}
