package password

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"

	"github.com/aide-family/magicbox/strutil"
)

type Password interface {
	Equal(password string) bool
	Value() string
	Salt() string
}

type password struct {
	value, enValue, salt string
}

func (p *password) Equal(hashedPassword string) bool {
	return checkPassword(hashedPassword, obfuscatePassword(p.value, p.salt))
}

func (p *password) Value() string {
	return p.enValue
}

func (p *password) Salt() string {
	return p.salt
}

func New(value string, salts ...string) (Password, error) {
	var salt string
	if len(salts) > 0 && salts[0] != "" {
		salt = salts[0]
	} else {
		salt = strutil.RandomString(16)
	}

	enValue, err := hashPassword(obfuscatePassword(value, salt))
	if err != nil {
		return nil, err
	}
	return &password{
		value:   value,
		enValue: enValue,
		salt:    salt,
	}, nil
}

func obfuscatePassword(password, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(password))
	hashed := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashed)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func checkPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
