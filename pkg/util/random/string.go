package random

import (
	"math/rand"
	"strings"
	"time"
)

// GenerateRandomString 生成随机字符串, 分别指定字母和数字的长度
func GenerateRandomString(letterLength, numberLength int) string {
	// 设置随机数种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 生成字母字符串
	var letters strings.Builder
	for i := 0; i < letterLength; i++ {
		letter := byte(r.Intn(26) + 65) // 生成随机字母 ASCII 码值范围是 65-90
		letters.WriteByte(letter)
	}

	// 生成数字字符串
	var numbers strings.Builder
	for i := 0; i < numberLength; i++ {
		number := byte(r.Intn(10) + 48) // 生成随机数字 ASCII 码值范围是 48-57
		numbers.WriteByte(number)
	}

	// 返回组合的字符串
	return strings.ToUpper(letters.String()) + numbers.String()
}

// 定义字符集
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_-+=<>?."

// GenerateRandomPassword 生成一个指定长度的随机密码
func GenerateRandomPassword(length int) string {
	// 创建新的随机数生成器
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// 创建一个字符串切片
	var passwordBuilder strings.Builder

	charsetLength := len(charset)
	// 随机生成密码
	for i := 0; i < length; i++ {
		// 从字符集中随机选一个字符
		randomIndex := r.Intn(charsetLength)
		passwordBuilder.WriteByte(charset[randomIndex])
	}

	return passwordBuilder.String()
}
