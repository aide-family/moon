package types

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// TextIsNull 判断字符串是否为空
func TextIsNull(text string) bool {
	return len(text) == 0
}

// TextNotNull 判断字符串是否不为空
func TextNotNull(text string) bool {
	return !TextIsNull(text)
}

// TextJoinToBytes 拼接字符串
func TextJoinToBytes(ss ...string) []byte {
	if len(ss) == 0 {
		return nil
	}
	if len(ss) == 1 {
		return []byte(ss[0])
	}
	length := 0
	for _, s := range ss {
		length += len(s)
	}
	buf := make([]byte, 0, length)
	for _, s := range ss {
		buf = append(buf, s...)
	}
	return buf
}

// TextJoin 拼接字符串
func TextJoin(s ...string) string {
	return string(TextJoinToBytes(s...))
}

// TextJoinByStringerToBytes 拼接字符串
func TextJoinByStringerToBytes(ss ...fmt.Stringer) []byte {
	if len(ss) == 0 {
		return nil
	}
	if len(ss) == 1 {
		return []byte(ss[0].String())
	}
	length := 0
	for _, s := range ss {
		length += len(s.String())
	}
	buf := make([]byte, 0, length)
	for _, s := range ss {
		buf = append(buf, s.String()...)
	}
	return buf
}

// TextJoinByStringer 拼接字符串
func TextJoinByStringer(s ...fmt.Stringer) string {
	return string(TextJoinByStringerToBytes(s...))
}

// TextJoinByBytesToBytes 拼接字符串
func TextJoinByBytesToBytes(ss ...[]byte) []byte {
	if len(ss) == 0 {
		return nil
	}
	if len(ss) == 1 {
		return ss[0]
	}
	length := 0
	for _, s := range ss {
		length += len(s)
	}
	buf := make([]byte, 0, length)
	for _, s := range ss {
		buf = append(buf, s...)
	}
	return buf
}

// TextJoinByBytes 拼接字符串
func TextJoinByBytes(s ...[]byte) string {
	return string(TextJoinByBytesToBytes(s...))
}

// GetAPI 从url中获取api
func GetAPI(path string) string {
	addr := strings.TrimPrefix(path, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	// 按照/分割
	parts := strings.Split(addr, "/")
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	u, err := url.JoinPath("/", parts[1:]...)
	if err != nil {
		return ""
	}
	return u
}

// UUID 生成uuid
func UUID() string {
	return uuid.New().String()
}

// ConvertString convert any type to string
func ConvertString(v interface{}) string {
	var key string
	if v == nil {
		return key
	}
	switch v := v.(type) {
	case float64:
		key = strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		key = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		key = strconv.Itoa(v)
	case uint:
		key = strconv.FormatUint(uint64(v), 10)
	case int8:
		key = strconv.Itoa(int(v))
	case uint8:
		key = strconv.FormatUint(uint64(v), 10)
	case int16:
		key = strconv.Itoa(int(v))
	case uint16:
		key = strconv.FormatUint(uint64(v), 10)
	case int32:
		key = strconv.Itoa(int(v))
	case uint32:
		key = strconv.FormatUint(uint64(v), 10)
	case int64:
		key = strconv.FormatInt(v, 10)
	case uint64:
		key = strconv.FormatUint(v, 10)
	case string:
		key = v
	case bool:
		key = strconv.FormatBool(v)
	case []byte:
		key = string(v)
	case fmt.Stringer:
		key = v.String()
	default:
		newValue, _ := Marshal(v)
		key = string(newValue)
	}
	return key
}
