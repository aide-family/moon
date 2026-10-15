package data

import (
	"crypto/rand"
	"encoding/hex"
)

func newID() string {
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	return hex.EncodeToString(raw[:])
}
