package validate

import (
	"regexp"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
)

// TextIsNull determines whether the string is empty
func TextIsNull(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}

func TextIsNotNull(text string) bool {
	return !TextIsNull(text)
}

func CheckEmail(email string) error {
	match := regexp.MustCompile(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`)
	if TextIsNull(email) || !match.MatchString(email) {
		return errors.New(400, "INVALID_EMAIL", "The email format is incorrect.")
	}
	return nil
}

func CheckURL(url string) error {
	match := regexp.MustCompile(`^(http|https)://`)
	if TextIsNull(url) || !match.MatchString(url) {
		return errors.New(400, "INVALID_URL", "The url format is incorrect.")
	}
	return nil
}
