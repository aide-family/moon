package strutil

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()" // Optional, includes special characters

func RandomString(length int) string {
	return RandomStringWithCharset(length, charset)
}

func RandomStringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		b[i] = charset[index.Int64()]
	}
	return string(b)
}

func RandomID() string {
	return RandomStringWithCharset(10, "abcdefghijklmnopqrstuvwxyz0123456789")
}
