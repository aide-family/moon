package types

import (
	"strings"
)

// TextIsNull 判断字符串是否为空
func TextIsNull(text string) bool {
	return len(text) == 0
}

// TextTrim 移除字符串首尾空格
func TextTrim(text string) string {
	return strings.TrimSpace(text)
}

// TextTrimIsNull 判断字符串首尾空格后是否为空
func TextTrimIsNull(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}
