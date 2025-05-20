package crypto_test

import (
	"testing"

	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/util/crypto"
)

func Test_Default(t *testing.T) {
	// Create a new AES instance with default key and IV
	aes, err := crypto.NewAes()
	if err != nil {
		t.Error(err)
	}

	plaintext := []byte("Hello, AES!")

	// Encrypt
	ciphertext, err := aes.Encrypt(plaintext)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Ciphertext: %x\n", ciphertext)
	// Decrypt
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Decrypted: %s\n", decrypted)
}

func Test_ECB(t *testing.T) {
	// Create a new AES instance with ECB mode
	aes, err := crypto.NewAes(crypto.WithMod(config.Crypto_AesConfig_ECB))
	if err != nil {
		t.Error(err)
	}

	plaintext := []byte("Hello, AES ECB Mode!")

	// Encrypt
	ciphertext, err := aes.Encrypt(plaintext)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Ciphertext (ECB): %x\n", ciphertext)

	// Decrypt
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Decrypted (ECB): %s\n", decrypted)
}

func Test_GCM(t *testing.T) {
	// Create a new AES instance with GCM mode
	aes, err := crypto.NewAes(crypto.WithMod(config.Crypto_AesConfig_GCM))
	if err != nil {
		t.Error(err)
	}

	plaintext := []byte("Hello, AES GCM Mode!")

	// Encrypt
	ciphertext, err := aes.Encrypt(plaintext)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Ciphertext (GCM): %x\n", ciphertext)

	// Decrypt
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		t.Error(err)
	}
	t.Logf("Decrypted (GCM): %s\n", decrypted)
}
