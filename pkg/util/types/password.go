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
		value = values[0]
		salt = GenerateSalt()
	case 2:
		value = values[0]
		salt = values[1]
	default:
		salt = GenerateSalt()
		value = random.GenerateRandomString(8, 0)
	}
	return &password{
		salt:  salt,
		value: value,
	}
}

type (
	// Password 密码
	Password interface {
		GetEncryptValue() (string, error)
		GetValue() string
		GetSalt() string
		fmt.Stringer
	}
	password struct {
		value, salt string
	}
)

// GetSalt 获取盐
func (p *password) GetSalt() string {
	return p.salt
}

// String 获取加密值
func (p *password) String() string {
	v, _ := p.GetEncryptValue()
	return v
}

// GetValue 获取密码值
func (p *password) GetValue() string {
	return p.value
}

// GetEncryptValue 获取加密值
func (p *password) GetEncryptValue() (string, error) {
	return GeneratePassword(p.value, p.salt)
}

// GeneratePassword 生成密码
func GeneratePassword(password, salt string) (string, error) {
	newPass := MD5(password + salt)
	return newPass, nil
}

// GenerateSalt 生成盐
func GenerateSalt() string {
	return MD5(strconv.FormatInt(time.Now().Unix(), 10))[0:16]
}

// ValidatePassword 校验密码
func ValidatePassword(password, checkPass, salt string) (err error) {
	if password == MD5(checkPass+salt) {
		return nil
	}
	return ErrValidatePassword
}
