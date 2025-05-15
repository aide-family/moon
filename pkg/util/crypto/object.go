package crypto

import (
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

var _ sql.Scanner = (*Object[any])(nil)
var _ driver.Valuer = (*Object[any])(nil)

func NewObject[T any](data T) *Object[T] {
	return &Object[T]{
		Data: data,
	}
}

type Object[T any] struct {
	Data T `json:"data"`
}

func (o *Object[T]) Get() T {
	return o.Data
}

func (o *Object[T]) Scan(value interface{}) error {
	aes, err := WithAes()
	if err != nil {
		return err
	}
	var origin string
	switch val := value.(type) {
	case []byte:
		origin = string(val)
	case string:
		origin = val
	default:
		return merr.ErrorInternalServerError("invalid value type of crypto.Object")
	}
	if len(origin) == 0 {
		return nil
	}
	decodedString, err := base64.StdEncoding.DecodeString(origin)
	if err != nil {
		return err
	}
	decrypt, err := aes.Decrypt(decodedString)
	if err != nil {
		return err
	}
	return json.Unmarshal(decrypt, o)
}

func (o Object[T]) Value() (driver.Value, error) {
	if validate.IsNil(o) || validate.IsNil(o.Data) {
		return "", nil
	}
	aes, err := WithAes()
	if err != nil {
		return "", err
	}
	bs, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	encrypt, err := aes.Encrypt(bs)
	if err != nil {
		return "", err
	}
	encodeToString := base64.StdEncoding.EncodeToString(encrypt)
	return encodeToString, nil
}
