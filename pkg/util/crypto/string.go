package crypto

import (
	"database/sql"
	"database/sql/driver"
	"encoding/base64"

	"github.com/aide-family/moon/pkg/merr"
)

var _ sql.Scanner = (*String)(nil)
var _ driver.Valuer = (*String)(nil)

type String string

func (s String) EQ(a String) bool {
	return string(s) == string(a)
}

func (s *String) Scan(value interface{}) error {
	aes, err := WithAes()
	if err != nil {
		return err
	}
	if value == nil {
		*s = ""
		return nil
	}
	val := ""
	switch v := value.(type) {
	case string:
		val = v
		if len(val) == 0 {
			*s = ""
			return nil
		}
	case []byte:
		val = string(v)
		if len(val) == 0 {
			*s = ""
			return nil
		}
	default:
		return merr.ErrorInternalServerError("invalid value type of crypto.String")
	}
	decodedString, err := base64.StdEncoding.DecodeString(val)
	if err != nil {
		return err
	}
	decrypt, err := aes.Decrypt(decodedString)
	if err != nil {
		return err
	}
	*s = String(decrypt)
	return nil
}

func (s String) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "", nil
	}
	aes, err := WithAes()
	if err != nil {
		return "", err
	}
	encrypt, err := aes.Encrypt([]byte(s))
	if err != nil {
		return "", err
	}
	encodeToString := base64.StdEncoding.EncodeToString(encrypt)
	return encodeToString, nil
}
