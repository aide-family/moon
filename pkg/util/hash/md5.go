package hash

import (
	"crypto/md5"
	"fmt"
)

func MD5(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}
