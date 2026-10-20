// Package notify 提供通知渠道之间共用的纯函数。
package notify

import "strings"

// SplitReceivers 拆分逗号、分号或换行分隔的接收方，去掉空项和首尾空白。
// alioth 用它做保存前校验，merak 用它做投递前兜底，两处必须保持一致。
func SplitReceivers(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r'
	})
	var out []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
