package strategy

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/aide-cloud/universal/cipher"
	"prometheus-manager/pkg/util/password"
)

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (b *BasicAuth) Scan(value interface{}) error {
	if value == nil || b == nil {
		return nil
	}
	switch expr := value.(type) {
	case []byte:
		basicAuth := NewBasicAuthWithString(string(expr))
		if basicAuth == nil {
			return nil
		}
		*b = *basicAuth
	case string:
		basicAuth := NewBasicAuthWithString(expr)
		if basicAuth == nil {
			return nil
		}
		*b = *basicAuth
	}
	return nil
}

func (b *BasicAuth) Value() (driver.Value, error) {
	if b == nil {
		return nil, nil
	}
	return b.String(), nil
}

func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

func NewBasicAuthWithString(str string) *BasicAuth {
	if str == "" {
		return nil
	}
	aes, err := cipher.NewAes(password.DefaultKey, password.DefaultIv)
	if err != nil {
		return nil
	}
	bs, err := aes.DecryptBase64(str)
	if err != nil {
		return nil
	}
	var b BasicAuth
	err = json.Unmarshal(bs, &b)
	if err != nil {
		return nil
	}
	return &b
}

func (b *BasicAuth) Bytes() []byte {
	if b == nil {
		return nil
	}
	bs, _ := json.Marshal(b)
	return bs
}

func (b *BasicAuth) String() string {
	if b == nil {
		return ""
	}
	aes, err := cipher.NewAes(password.DefaultKey, password.DefaultIv)
	if err != nil {
		return ""
	}
	newPass, err := aes.EncryptBase64(b.Bytes())
	if err != nil {
		return ""
	}
	return newPass
}
