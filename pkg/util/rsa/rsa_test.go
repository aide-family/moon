package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestRSAEncryption(t *testing.T) {
	plainText := []byte("Hello, World!")
	key, err := rsa.GenerateKey(rand.Reader, DefaultLength)
	if err != nil {
		t.Fatal("Failed to generate RSA key:", err)
	}

	// encrypt
	cipherText, err := RSAEncryptByPublicKey(plainText, &key.PublicKey)
	if err != nil {
		t.Fatal("Failed to encrypt:", err)
	}

	// decrypt
	decryptedText, err := RSADecryptByPrivateKey(cipherText, key)
	if err != nil {
		t.Fatal("Failed to decrypt:", err)
	}

	if string(decryptedText) != string(plainText) {
		t.Error("Decrypted text does not match original plain text")
	}
	t.Log("Encryption and decryption successful")
}

func TestEncodePrivateKeyPEM(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, DefaultLength)
	if err != nil {
		t.Fatal("Failed to generate RSA key:", err)
	}

	pemBytes := EncodePrivateKeyPEM(key)
	if len(pemBytes) == 0 {
		t.Error("Failed to encode private key to PEM")
	}

	t.Logf("Private key encoded to PEM:\n%s", string(pemBytes))
}
