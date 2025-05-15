package password

import (
	"crypto/rand"
	"math/big"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()" // Optional, includes special characters

// GenerateRandomPassword generates a random password of the specified length.
func GenerateRandomPassword(length int) string {
	if length < 0 {
		return ""
	}
	p := make([]byte, length)
	for i := range p {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		p[i] = charset[index.Int64()]
	}

	return string(p)
}
