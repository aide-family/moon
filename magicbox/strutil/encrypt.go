package strutil

import (
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/aide-family/magicbox/pointer"
)

type EncryptInterface interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

var encrypt EncryptInterface = &defaultBase64Encrypt{}
var once sync.Once

func SetEncrypt(e EncryptInterface) {
	once.Do(func() {
		encrypt = e
	})
}

func GetEncrypt() EncryptInterface {
	return encrypt
}

var _ sql.Scanner = (*EncryptString)(nil)
var _ driver.Valuer = (*EncryptString)(nil)

type EncryptString string

// Value implements driver.Valuer.
func (e EncryptString) Value() (driver.Value, error) {
	return encrypt.Encrypt(string(e))
}

// Scan implements sql.Scanner.
func (e *EncryptString) Scan(src any) error {
	if pointer.IsNil(src) {
		*e = ""
		return nil
	}
	value := ""
	switch v := src.(type) {
	case string:
		value = v
	case []byte:
		value = string(v)
	default:
		return fmt.Errorf("invalid type %T", src)
	}
	encrypted, err := encrypt.Decrypt(value)
	if err != nil {
		return err
	}
	*e = EncryptString(encrypted)
	return nil
}

type defaultBase64Encrypt struct{}

func (e *defaultBase64Encrypt) Encrypt(s string) (string, error) {
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}

func (e *defaultBase64Encrypt) Decrypt(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
