package hash

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 returns the MD5 checksum of the data.
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
