package validate

import (
	"regexp"
	"strings"

	"github.com/moon-monitor/moon/pkg/merr"
)

// TextIsNull 判断字符串是否为空
func TextIsNull(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}

func TextIsNotNull(text string) bool {
	return !TextIsNull(text)
}

func CheckEmail(email string) error {
	match := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if TextIsNull(email) || !match.MatchString(email) {
		return merr.ErrorParamsError("The email format is incorrect.")
	}
	return nil
}

func CheckURL(url string) error {
	match := regexp.MustCompile(`^(http|https)://`)
	if TextIsNull(url) || !match.MatchString(url) {
		return merr.ErrorParamsError("The url format is incorrect.")
	}
	return nil
}
