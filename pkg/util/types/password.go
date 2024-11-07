package types

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aide-family/moon/pkg/util/random"

	"google.golang.org/grpc/status"
)

// ErrValidatePassword 密码错误
var ErrValidatePassword = status.Error(http.StatusUnauthorized, "密码错误")

// NewPassword 创建密码
func NewPassword(values ...string) Password {
	var value, salt string
	switch len(values) {
	case 1:
		salt = generateSalt()
		value = generatePassword(values[0], salt)
	case 2:
		value = values[0]
		salt = values[1]
	default:
		salt = generateSalt()
		value = generatePassword(random.GenerateRandomString(8, 0), salt)
	}
	return &password{
		salt:  salt,
		value: value,
	}
}

type (
	// Password 密码
	Password interface {
		GetValue() string
		GetSalt() string
		fmt.Stringer
		Validate(checkPass string) error
	}
	password struct {
		value, salt string
	}
)

func (p *password) Validate(checkPass string) error {
	return validatePassword(p.value, checkPass, p.salt)
}

// GetSalt 获取盐
func (p *password) GetSalt() string {
	return p.salt
}

// String 获取加密值
func (p *password) String() string {
	return p.value
}

// GetValue 获取密码值
func (p *password) GetValue() string {
	return p.value
}

// GeneratePassword 生成密码
func generatePassword(password, salt string) string {
	newPass := MD5(TextJoin(password, salt))
	return newPass
}

// GenerateSalt 生成盐
func generateSalt() string {
	return MD5(strconv.FormatInt(time.Now().Unix(), 10))[0:16]
}

func validatePassword(password, checkPass, salt string) (err error) {
	if password == MD5(TextJoin(checkPass, salt)) {
		return nil
	}
	return ErrValidatePassword
}
