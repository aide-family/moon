package helper

import (
	"regexp"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
)

// CheckEmail 检查邮箱格式
func CheckEmail(email string) error {
	match := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if types.TextIsNull(email) || !match.MatchString(email) {
		return merr.ErrorAlert("邮箱格式不正确").WithMetadata(map[string]string{
			"email": email,
		})
	}
	return nil
}
