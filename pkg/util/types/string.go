package types

import (
	"regexp"

	"github.com/aide-family/moon/api/merr"
)

// TextIsNull 判断字符串是否为空
func TextIsNull(text string) bool {
	return len(text) == 0
}

// CheckEmail 检查邮箱格式
func CheckEmail(email string) error {
	match := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if TextIsNull(email) || !match.MatchString(email) {
		return merr.ErrorAlert("邮箱格式不正确").WithMetadata(map[string]string{
			"email": email,
		})
	}
	return nil
}
