package data

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/aide-family/Alioth/pkg/merr"
)

type secretBox struct {
	key []byte
}

func newSecretBox(raw string) (*secretBox, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, merr.Server("encrypt key is empty")
	}
	sum := sha256.Sum256([]byte(raw))
	return &secretBox{key: sum[:]}, nil
}

func (b *secretBox) encrypt(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	block, err := aes.NewCipher(b.key)
	if err != nil {
		return "", merr.Server("request failed")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", merr.Server("request failed")
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", merr.Server("request failed")
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (b *secretBox) decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", merr.Server("request failed")
	}
	block, err := aes.NewCipher(b.key)
	if err != nil {
		return "", merr.Server("request failed")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", merr.Server("request failed")
	}
	if len(raw) < gcm.NonceSize() {
		return "", merr.Server("request failed")
	}
	nonce, body := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return "", merr.Server("request failed")
	}
	return string(plain), nil
}
