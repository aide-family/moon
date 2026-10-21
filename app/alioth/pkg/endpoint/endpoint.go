package endpoint

import (
	"net/url"
	"strings"
)

// Normalize 去掉地址首尾空白和末尾斜杠，并把协议、主机名转成小写。
// 数据源连接是否重复按这个结果比较，避免 `http://host/` 和 `http://HOST`
// 被当成两个连接。解析失败时只做去空白处理，由调用方继续报地址错误。
func Normalize(raw string) string {
	trimmed := strings.TrimSpace(raw)
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Host == "" {
		return strings.TrimRight(trimmed, "/")
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	return strings.TrimRight(parsed.String(), "/")
}
