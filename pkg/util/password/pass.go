package password

import (
	"strconv"
	"time"

	"github.com/aide-cloud/universal/cipher"

	"prometheus-manager/api/perrors"
)

const (
	key        = "1234567890123456"
	DefaultKey = key
	DefaultIv  = "1234567890123456"
)

var ErrValidatePassword = perrors.ErrorInvalidParams("密码错误")

// GeneratePassword 生成密码
func GeneratePassword(password, salt string) (string, error) {
	aes, err := cipher.NewAes(key, salt)
	if err != nil {
		return "", err
	}
	newPass, err := aes.EncryptBase64([]byte(password))
	if err != nil {
		return "", err
	}
	return newPass, nil
}

// GenerateSalt 生成盐
func GenerateSalt() string {
	return cipher.MD5(strconv.FormatInt(time.Now().Unix(), 10))[0:16]
}

// DecryptPassword 解密密码
func DecryptPassword(password, salt string) (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrValidatePassword
		}
	}()
	aes, err := cipher.NewAes(key, salt)
	if err != nil {
		return "", err
	}
	newPass, err := aes.DecryptBase64(password)
	if err != nil {
		return "", err
	}
	return string(newPass), nil
}

// EncryptPassword 加密密码
func EncryptPassword(password, salt string) (str string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrValidatePassword
		}
	}()
	aes, err := cipher.NewAes(key, salt)
	if err != nil {
		return "", err
	}
	newPass, err := aes.EncryptBase64([]byte(password))
	if err != nil {
		return "", err
	}
	return newPass, nil
}

// ValidatePassword 验证密码
func ValidatePassword(p1, dePass, salt string) bool {
	newPass, err := DecryptPassword(dePass, salt)
	if err != nil {
		return false
	}

	return p1 == newPass
}

// ValidatePasswordErr 验证密码
func ValidatePasswordErr(p1, dePass, salt string) error {
	if ValidatePassword(p1, dePass, salt) {
		return nil
	}
	return ErrValidatePassword
}
