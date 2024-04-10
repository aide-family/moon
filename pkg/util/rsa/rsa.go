package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"k8s.io/client-go/util/keyutil"
)

const (
	DefaultLength = 2048
)

// RSAEncryptByPublicKey 加密
func RSAEncryptByPublicKey(plainText []byte, key *rsa.PublicKey) ([]byte, error) {
	partLen := key.N.BitLen()/8 - 11
	chunks := split(plainText, partLen)

	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		buf, err := rsa.EncryptPKCS1v15(rand.Reader, key, chunk)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(buf)
	}

	return []byte(base64.RawStdEncoding.EncodeToString(buffer.Bytes())), nil
}

// RSADecryptByPrivateKey 解密
func RSADecryptByPrivateKey(cipherText []byte, key *rsa.PrivateKey) ([]byte, error) {
	partLen := key.N.BitLen() / 8
	raw, err := base64.RawStdEncoding.DecodeString(string(cipherText))
	chunks := split(raw, partLen)

	buffer := bytes.NewBufferString("")
	for _, chunk := range chunks {
		decrypted, err := rsa.DecryptPKCS1v15(rand.Reader, key, chunk)
		if err != nil {
			return []byte{}, err
		}
		buffer.Write(decrypted)
	}

	return buffer.Bytes(), err
}
func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}

func NewPrivateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, DefaultLength)
}
func EncodePrivateKeyPEM(key *rsa.PrivateKey) []byte {
	block := pem.Block{
		Type:  keyutil.RSAPrivateKeyBlockType,
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.EncodeToMemory(&block)
}
