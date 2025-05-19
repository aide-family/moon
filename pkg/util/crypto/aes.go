package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"io"
	"sync"

	"github.com/go-kratos/kratos/v2/errors"

	"github.com/aide-family/moon/pkg/config"
)

var (
	aesInstance AES
	aesOnce     sync.Once
)

// WithAes returns a new AES instance with the provided options.
func WithAes(opts ...AesOption) (AES, error) {
	var err error
	aesOnce.Do(func() {
		aesInstance, err = NewAes(opts...)
	})
	return aesInstance, err
}

const (
	defaultAesKey = "palace-secretKey"
	defaultAesIv  = "palace-secret-iv"
)

// NewAes creates a new AES instance with the provided options.
func NewAes(opts ...AesOption) (AES, error) {
	a := &aesImpl{
		key:  []byte(defaultAesKey),
		iv:   []byte(defaultAesIv),
		mode: config.Crypto_AesConfig_CBC,
	}

	for _, opt := range opts {
		opt(a)
	}

	if a.mode == config.Crypto_AesConfig_ECB {
		key, err := aesSha1Padding(a.key, 128)
		if err != nil {
			return nil, err
		}
		a.key = generateKey(key)
	}

	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}
	a.block = block

	return a, nil
}

func Sha1(data []byte) []byte {
	h := sha1.New()
	h.Write(data)
	return h.Sum(nil)
}

func generateKey(key []byte) (genKey []byte) {
	genKey = make([]byte, 16)
	copy(genKey, key)
	for i := 16; i < len(key); {
		for j := 0; j < 16 && i < len(key); j, i = j+1, i+1 {
			genKey[j] ^= key[i]
		}
	}
	return genKey
}

func aesSha1Padding(keyBytes []byte, encryptLength int) ([]byte, error) {
	hashes := Sha1(Sha1(keyBytes))
	maxLen := len(hashes)
	realLen := encryptLength / 8
	if realLen > maxLen {
		return nil, errors.New(400, "INVALID_LENGTH", "invalid length")
	}
	return hashes[0:realLen], nil
}

type aesImpl struct {
	key, iv []byte
	mode    config.Crypto_AesConfig_MODE

	block cipher.Block
}

type AesOption func(a *aesImpl)

type AES interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// Encrypt encrypts the plaintext using AES.
func (a *aesImpl) Encrypt(plaintext []byte) ([]byte, error) {
	switch a.mode {
	case config.Crypto_AesConfig_CBC:
		return a.encryptCBC(plaintext)
	case config.Crypto_AesConfig_GCM:
		return a.encryptGCM(plaintext)
	case config.Crypto_AesConfig_ECB:
		return a.encryptECB(plaintext)
	default:
		return nil, errors.Newf(400, "UNSUPPORTED_AES_MODE", "unsupported AES mode %v", a.mode)
	}
}

// Decrypt decrypts the ciphertext using AES.
func (a *aesImpl) Decrypt(ciphertext []byte) ([]byte, error) {
	switch a.mode {
	case config.Crypto_AesConfig_CBC:
		return a.decryptCBC(ciphertext)
	case config.Crypto_AesConfig_GCM:
		return a.decryptGCM(ciphertext)
	case config.Crypto_AesConfig_ECB:
		return a.decryptECB(ciphertext)
	default:
		return nil, errors.Newf(400, "UNSUPPORTED_AES_MODE", "unsupported AES mode %v", a.mode)
	}
}

// encryptCBC encrypts the plaintext using AES in CBC mode.
func (a *aesImpl) encryptCBC(plaintext []byte) ([]byte, error) {
	blockSize := a.block.BlockSize()
	plaintext = pkcs7Pad(plaintext, blockSize)

	ciphertext := make([]byte, blockSize+len(plaintext))
	iv := ciphertext[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	mode := cipher.NewCBCEncrypter(a.block, iv)
	mode.CryptBlocks(ciphertext[blockSize:], plaintext)

	return ciphertext, nil
}

// decryptCBC decrypts the ciphertext using AES in CBC mode.
func (a *aesImpl) decryptCBC(ciphertext []byte) ([]byte, error) {
	blockSize := a.block.BlockSize()

	if len(ciphertext) < blockSize {
		return nil, errors.New(400, "CIPHERTEXT_TOO_SHORT", "ciphertext too short")
	}

	iv := ciphertext[:blockSize]
	ciphertext = ciphertext[blockSize:]

	if len(ciphertext)%blockSize != 0 {
		return nil, errors.New(400, "CIPHERTEXT_NOT_A_MULTIPLE_OF_THE_BLOCK_SIZE", "ciphertext is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(a.block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	plaintext := pkcs7Unpad(ciphertext)
	return plaintext, nil
}

// encryptGCM encrypts the plaintext using AES in GCM mode.
func (a *aesImpl) encryptGCM(plaintext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(a.block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decryptGCM decrypts the ciphertext using AES in GCM mode.
func (a *aesImpl) decryptGCM(ciphertext []byte) ([]byte, error) {
	gcm, err := cipher.NewGCM(a.block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New(400, "CIPHERTEXT_TOO_SHORT", "ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// encryptECB encrypts the plaintext using AES in ECB mode.
func (a *aesImpl) encryptECB(plaintext []byte) ([]byte, error) {
	length := (len(plaintext) + aes.BlockSize) / aes.BlockSize
	plain := make([]byte, length*aes.BlockSize)
	blockSize := a.block.BlockSize()

	copy(plain, plaintext)
	pad := byte(len(plain) - len(plaintext))
	for i := len(plaintext); i < len(plain); i++ {
		plain[i] = pad
	}
	encrypted := make([]byte, len(plain))
	// 分组分块加密
	for bs, be := 0, blockSize; bs <= len(plaintext); bs, be = bs+blockSize, be+blockSize {
		a.block.Encrypt(encrypted[bs:be], plain[bs:be])
	}

	return encrypted, nil
}

// decryptECB decrypts the ciphertext using AES in ECB mode.
func (a *aesImpl) decryptECB(ciphertext []byte) ([]byte, error) {
	blockSize := a.block.BlockSize()

	decrypted := make([]byte, len(ciphertext))

	for bs, be := 0, blockSize; bs < len(ciphertext); bs, be = bs+blockSize, be+blockSize {
		a.block.Decrypt(decrypted[bs:be], ciphertext[bs:be])
	}

	trim := 0
	if len(decrypted) > 0 {
		trim = len(decrypted) - int(decrypted[len(decrypted)-1])
	}

	return decrypted[:trim], nil
}

// pkcs7Pad pads the data to the specified block size using PKCS7 padding.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// pkcs7Unpad removes the PKCS7 padding from the data.
func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	padding := int(data[len(data)-1]) // 获取填充字节
	if padding > len(data) {
		return data // 如果填充字节无效，返回原始数据
	}
	return data[:len(data)-padding] // 移除填充字节
}
