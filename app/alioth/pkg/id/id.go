package id

import (
	"crypto/rand"
	"encoding/hex"
)

// New 返回 16 字节随机数的十六进制形式，用作 UID。
func New() string {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
}
