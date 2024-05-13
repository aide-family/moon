package cipher

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"
)

// MD5 returns the MD5 checksum of the data.
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateRandomString(letterLength, numberLength int) string {
	// 设置随机数种子
	rand.Seed(time.Now().UnixNano())

	// 生成字母字符串
	var letters strings.Builder
	for i := 0; i < letterLength; i++ {
		letter := byte(rand.Intn(26) + 65) // 生成随机字母 ASCII 码值范围是 65-90
		letters.WriteByte(letter)
	}

	// 生成数字字符串
	var numbers strings.Builder
	for i := 0; i < numberLength; i++ {
		number := byte(rand.Intn(10) + 48) // 生成随机数字 ASCII 码值范围是 48-57
		numbers.WriteByte(number)
	}

	// 返回组合的字符串
	return strings.ToUpper(letters.String()) + numbers.String()
}
