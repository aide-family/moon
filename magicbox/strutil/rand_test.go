package strutil_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aide-family/magicbox/strutil"
)

// TestRandomString_Length 测试 RandomString 生成字符串的长度
func TestRandomString_Length(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "Length 0",
			length: 0,
		},
		{
			name:   "Length 1",
			length: 1,
		},
		{
			name:   "Length 10",
			length: 10,
		},
		{
			name:   "Length 100",
			length: 100,
		},
		{
			name:   "Length 1000",
			length: 1000,
		},
		{
			name:   "Length 16",
			length: 16,
		},
		{
			name:   "Length 32",
			length: 32,
		},
		{
			name:   "Length 64",
			length: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.RandomString(tt.length)
			if len(result) != tt.length {
				t.Errorf("RandomString(%d) length = %d, want %d", tt.length, len(result), tt.length)
			}
		})
	}
}

// TestRandomString_Charset 测试 RandomString 生成的字符都在默认字符集中
func TestRandomString_Charset(t *testing.T) {
	// 默认字符集
	defaultCharset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"

	// 生成多个随机字符串并验证字符集
	for i := 0; i < 100; i++ {
		result := strutil.RandomString(100)
		for _, char := range result {
			if !strings.ContainsRune(defaultCharset, char) {
				t.Errorf("RandomString() generated character %c not in charset", char)
			}
		}
	}
}

// TestRandomString_Randomness 测试 RandomString 的随机性
// 多次调用应该产生不同的结果（至少大部分情况下）
func TestRandomString_Randomness(t *testing.T) {
	results := make(map[string]bool)
	count := 100

	// 生成多个随机字符串
	for i := 0; i < count; i++ {
		result := strutil.RandomString(20)
		if results[result] {
			t.Logf("Duplicate string found: %s (this is possible but unlikely)", result)
		}
		results[result] = true
	}

	// 验证产生了多个不同的字符串（至少应该有一些不同）
	// 由于是随机生成，理论上应该有很多不同的字符串
	if len(results) < count/10 {
		t.Errorf("RandomString() generated only %d unique strings out of %d attempts, randomness may be insufficient", len(results), count)
	}

	t.Logf("Generated %d unique strings out of %d attempts", len(results), count)
}

// TestRandomStringWithCharset_Length 测试 RandomStringWithCharset 生成字符串的长度
func TestRandomStringWithCharset_Length(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		charset  string
		expected int
	}{
		{
			name:     "Length 0",
			length:   0,
			charset:  "abc",
			expected: 0,
		},
		{
			name:     "Length 1",
			length:   1,
			charset:  "abc",
			expected: 1,
		},
		{
			name:     "Length 10",
			length:   10,
			charset:  "abc",
			expected: 10,
		},
		{
			name:     "Length 100",
			length:   100,
			charset:  "0123456789",
			expected: 100,
		},
		{
			name:     "Length 1000",
			length:   1000,
			charset:  "abcdefghijklmnopqrstuvwxyz",
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.RandomStringWithCharset(tt.length, tt.charset)
			if len(result) != tt.expected {
				t.Errorf("RandomStringWithCharset(%d, %q) length = %d, want %d", tt.length, tt.charset, len(result), tt.expected)
			}
		})
	}
}

// TestRandomStringWithCharset_Charset 测试 RandomStringWithCharset 生成的字符都在指定字符集中
func TestRandomStringWithCharset_Charset(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
	}{
		{
			name:    "Lowercase letters",
			length:  100,
			charset: "abcdefghijklmnopqrstuvwxyz",
		},
		{
			name:    "Uppercase letters",
			length:  100,
			charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
		{
			name:    "Digits",
			length:  100,
			charset: "0123456789",
		},
		{
			name:    "Alphanumeric",
			length:  100,
			charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		{
			name:    "Special characters",
			length:  100,
			charset: "!@#$%^&*()",
		},
		{
			name:    "Single character charset",
			length:  100,
			charset: "a",
		},
		{
			name:    "Two character charset",
			length:  100,
			charset: "ab",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 生成多个随机字符串并验证字符集
			for i := 0; i < 10; i++ {
				result := strutil.RandomStringWithCharset(tt.length, tt.charset)
				for _, char := range result {
					if !strings.ContainsRune(tt.charset, char) {
						t.Errorf("RandomStringWithCharset(%d, %q) generated character %c not in charset", tt.length, tt.charset, char)
					}
				}
			}
		})
	}
}

// TestRandomStringWithCharset_Randomness 测试 RandomStringWithCharset 的随机性
func TestRandomStringWithCharset_Randomness(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
	}{
		{
			name:    "Small charset",
			length:  20,
			charset: "abc",
		},
		{
			name:    "Large charset",
			length:  20,
			charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
		{
			name:    "Digits only",
			length:  20,
			charset: "0123456789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := make(map[string]bool)
			count := 100

			// 生成多个随机字符串
			for i := 0; i < count; i++ {
				result := strutil.RandomStringWithCharset(tt.length, tt.charset)
				results[result] = true
			}

			// 对于较大的字符集，应该产生更多不同的字符串
			// 对于较小的字符集，可能会有一些重复，但仍然应该有足够的多样性
			minUnique := count / 10
			if len(tt.charset) > 10 {
				minUnique = count / 2
			}

			if len(results) < minUnique {
				t.Logf("Warning: RandomStringWithCharset() generated only %d unique strings out of %d attempts with charset %q", len(results), count, tt.charset)
			}

			t.Logf("Generated %d unique strings out of %d attempts with charset %q", len(results), count, tt.charset)
		})
	}
}

// TestRandomStringWithCharset_EdgeCases 测试边界情况
func TestRandomStringWithCharset_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		charset  string
		validate func(string) bool
	}{
		{
			name:    "Empty charset with length 0",
			length:  0,
			charset: "",
			validate: func(s string) bool {
				return len(s) == 0
			},
		},
		{
			name:    "Single character charset",
			length:  10,
			charset: "a",
			validate: func(s string) bool {
				return len(s) == 10 && strings.ReplaceAll(s, "a", "") == ""
			},
		},
		{
			name:    "Two character charset",
			length:  100,
			charset: "ab",
			validate: func(s string) bool {
				if len(s) != 100 {
					return false
				}
				for _, char := range s {
					if char != 'a' && char != 'b' {
						return false
					}
				}
				return true
			},
		},
		{
			name:    "Very long charset",
			length:  10,
			charset: strings.Repeat("abcdefghijklmnopqrstuvwxyz", 10),
			validate: func(s string) bool {
				return len(s) == 10
			},
		},
		{
			name:    "Single byte charset",
			length:  20,
			charset: "abcdefghijklmnopqrstuvwxyz",
			validate: func(s string) bool {
				return len(s) == 20
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 对于空字符集，函数可能会 panic（取决于实现）
			// 让我们先测试一下
			if tt.charset == "" && tt.length > 0 {
				defer func() {
					if r := recover(); r != nil {
						t.Logf("RandomStringWithCharset panicked with empty charset (expected): %v", r)
					}
				}()
			}

			result := strutil.RandomStringWithCharset(tt.length, tt.charset)
			if !tt.validate(result) {
				t.Errorf("RandomStringWithCharset(%d, %q) = %q, validation failed", tt.length, tt.charset, result)
			}
		})
	}
}

// TestRandomStringWithCharset_SingleCharacterCharset 测试单字符字符集
func TestRandomStringWithCharset_SingleCharacterCharset(t *testing.T) {
	charset := "a"
	length := 10

	result := strutil.RandomStringWithCharset(length, charset)

	// 验证长度
	if len(result) != length {
		t.Errorf("RandomStringWithCharset(%d, %q) length = %d, want %d", length, charset, len(result), length)
	}

	// 验证所有字符都是 'a'
	for i, char := range result {
		if char != 'a' {
			t.Errorf("RandomStringWithCharset(%d, %q) character at index %d = %c, want 'a'", length, charset, i, char)
		}
	}
}

// TestRandomID_Length 测试 RandomID 生成字符串的长度
func TestRandomID_Length(t *testing.T) {
	result := strutil.RandomID()
	expectedLength := 10

	if len(result) != expectedLength {
		t.Errorf("RandomID() length = %d, want %d", len(result), expectedLength)
	}
}

// TestRandomID_Charset 测试 RandomID 生成的字符都在正确的字符集中
func TestRandomID_Charset(t *testing.T) {
	// RandomID 使用的字符集
	expectedCharset := "abcdefghijklmnopqrstuvwxyz0123456789"

	// 生成多个随机 ID 并验证字符集
	for i := 0; i < 100; i++ {
		result := strutil.RandomID()
		for _, char := range result {
			if !strings.ContainsRune(expectedCharset, char) {
				t.Errorf("RandomID() generated character %c not in charset %q", char, expectedCharset)
			}
		}
	}
}

// TestRandomID_Randomness 测试 RandomID 的随机性
func TestRandomID_Randomness(t *testing.T) {
	results := make(map[string]bool)
	count := 1000

	// 生成多个随机 ID
	for i := 0; i < count; i++ {
		result := strutil.RandomID()
		results[result] = true
	}

	// 验证产生了多个不同的 ID
	// 对于 10 个字符，36 个字符的字符集，理论上有 36^10 种可能
	// 1000 次调用应该产生接近 1000 个不同的 ID
	if len(results) < count*9/10 {
		t.Errorf("RandomID() generated only %d unique IDs out of %d attempts, randomness may be insufficient", len(results), count)
	}

	t.Logf("Generated %d unique IDs out of %d attempts", len(results), count)
}

// TestRandomID_Format 测试 RandomID 的格式（应该只包含小写字母和数字）
func TestRandomID_Format(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := strutil.RandomID()

		// 验证长度
		if len(result) != 10 {
			t.Errorf("RandomID() length = %d, want 10", len(result))
		}

		// 验证字符集（只包含小写字母和数字）
		for _, char := range result {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				t.Errorf("RandomID() generated invalid character %c, want lowercase letter or digit", char)
			}
		}
	}
}

// TestRandomString_Consistency 测试 RandomString 在不同调用之间的行为
func TestRandomString_Consistency(t *testing.T) {
	// 多次调用应该产生不同长度的字符串（如果长度相同）
	// 但长度应该一致
	length := 20
	results := make([]string, 10)

	for i := 0; i < 10; i++ {
		results[i] = strutil.RandomString(length)
	}

	// 验证所有结果长度都相同
	for i, result := range results {
		if len(result) != length {
			t.Errorf("RandomString(%d) result[%d] length = %d, want %d", length, i, len(result), length)
		}
	}

	// 验证至少有一些不同的字符串
	unique := make(map[string]bool)
	for _, result := range results {
		unique[result] = true
	}

	if len(unique) == 1 {
		t.Logf("Warning: All RandomString() calls produced the same result (unlikely but possible)")
	}
}

// TestRandomStringWithCharset_VariousCharsets 测试各种不同的字符集
func TestRandomStringWithCharset_VariousCharsets(t *testing.T) {
	tests := []struct {
		name    string
		charset string
		length  int
	}{
		{
			name:    "Binary",
			charset: "01",
			length:  20,
		},
		{
			name:    "Hex",
			charset: "0123456789abcdef",
			length:  20,
		},
		{
			name:    "Base64-like",
			charset: "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/",
			length:  20,
		},
		{
			name:    "URL-safe",
			charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_",
			length:  20,
		},
		{
			name:    "Password-like",
			charset: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*",
			length:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.RandomStringWithCharset(tt.length, tt.charset)

			// 验证长度
			if len(result) != tt.length {
				t.Errorf("RandomStringWithCharset(%d, %q) length = %d, want %d", tt.length, tt.charset, len(result), tt.length)
			}

			// 验证字符集
			for _, char := range result {
				if !strings.ContainsRune(tt.charset, char) {
					t.Errorf("RandomStringWithCharset(%d, %q) generated character %c not in charset", tt.length, tt.charset, char)
				}
			}
		})
	}
}

// TestRandomString_NoPanic 测试 RandomString 不会 panic
func TestRandomString_NoPanic(t *testing.T) {
	// 多次调用确保不会 panic
	for i := 0; i < 100; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RandomString(%d) panicked: %v", i, r)
				}
			}()
			_ = strutil.RandomString(i)
		}()
	}
}

// TestRandomStringWithCharset_NoPanic 测试 RandomStringWithCharset 不会 panic
func TestRandomStringWithCharset_NoPanic(t *testing.T) {
	testCharsets := []string{
		"abc",
		"0123456789",
		"abcdefghijklmnopqrstuvwxyz",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"!@#$%^&*()",
	}

	// 多次调用确保不会 panic
	for _, charset := range testCharsets {
		for length := 0; length < 100; length++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("RandomStringWithCharset(%d, %q) panicked: %v", length, charset, r)
					}
				}()
				_ = strutil.RandomStringWithCharset(length, charset)
			}()
		}
	}
}

// TestRandomID_NoPanic 测试 RandomID 不会 panic
func TestRandomID_NoPanic(t *testing.T) {
	// 多次调用确保不会 panic
	for i := 0; i < 100; i++ {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("RandomID() panicked: %v", r)
				}
			}()
			_ = strutil.RandomID()
		}()
	}
}

// BenchmarkRandomString 基准测试 RandomString 函数
func BenchmarkRandomString(b *testing.B) {
	length := 20
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.RandomString(length)
	}
}

// BenchmarkRandomStringWithCharset 基准测试 RandomStringWithCharset 函数
func BenchmarkRandomStringWithCharset(b *testing.B) {
	length := 20
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.RandomStringWithCharset(length, charset)
	}
}

// BenchmarkRandomID 基准测试 RandomID 函数
func BenchmarkRandomID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.RandomID()
	}
}

// BenchmarkRandomString_VariousLengths 基准测试不同长度的 RandomString
func BenchmarkRandomString_VariousLengths(b *testing.B) {
	lengths := []int{10, 50, 100, 500, 1000}
	for _, length := range lengths {
		b.Run(fmt.Sprintf("Length%d", length), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = strutil.RandomString(length)
			}
		})
	}
}
