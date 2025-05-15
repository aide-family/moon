package crypto_test

import (
	"testing"

	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/util/crypto"
)

func Test_Default(t *testing.T) {
	// 创建一个新的 AES 实例，使用默认的密钥和 IV
	aes, err := crypto.NewAes()
	if err != nil {
		t.Fatal(err)
	}

	plaintext := []byte("Hello, AES!")

	// 加密
	ciphertext, err := aes.Encrypt(plaintext)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Ciphertext: %x\n", ciphertext)
	// 解密
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		panic(err)
	}
	t.Logf("Decrypted: %s\n", decrypted)
}

func Test_ECB(t *testing.T) {
	// 创建一个新的 AES 实例，使用 ECB 模式
	aes, err := crypto.NewAes(crypto.WithMod(config.Crypto_AesConfig_ECB))
	if err != nil {
		panic(err)
	}

	plaintext := []byte("Hello, AES ECB Mode!")

	// 加密
	ciphertext, err := aes.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}
	t.Logf("Ciphertext (ECB): %x\n", ciphertext)

	// 解密
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		panic(err)
	}
	t.Logf("Decrypted (ECB): %s\n", decrypted)
}

func Test_GCM(t *testing.T) {
	// 创建一个新的 AES 实例，使用 GCM 模式
	aes, err := crypto.NewAes(crypto.WithMod(config.Crypto_AesConfig_GCM))
	if err != nil {
		panic(err)
	}

	plaintext := []byte("Hello, AES GCM Mode!")

	// 加密
	ciphertext, err := aes.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}
	t.Logf("Ciphertext (GCM): %x\n", ciphertext)

	// 解密
	decrypted, err := aes.Decrypt(ciphertext)
	if err != nil {
		panic(err)
	}
	t.Logf("Decrypted (GCM): %s\n", decrypted)
}
