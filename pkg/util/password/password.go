package password

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var ErrPasswordSizeLTZero = errors.New("password size must be greater than 0")

// HashPassword hashing is done using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// CheckPassword Verify that the password matches
func CheckPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateSalt Generate random salt
func GenerateSalt(size int) ([]byte, error) {
	if size < 0 {
		return nil, ErrPasswordSizeLTZero
	}
	salt := make([]byte, size)
	_, err := rand.Read(salt)
	return salt, err
}

// ObfuscatePassword Confuse Password and Salt
func ObfuscatePassword(password, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(password))
	hashed := h.Sum(nil)

	return base64.StdEncoding.EncodeToString(hashed)
}
