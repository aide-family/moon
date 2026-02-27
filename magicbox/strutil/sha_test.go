package strutil

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

// TestSHA256 tests the SHA256 function with various inputs
func TestSHA256(t *testing.T) {
	// 测试用例定义
	tests := []struct {
		name     string // 测试用例名称
		input    string // 输入字符串
		expected string // 期望的SHA256哈希值
	}{
		{
			name:     "Empty string",                                                     // 空字符串测试
			input:    "",                                                                 // 输入为空字符串
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // 标准SHA256空字符串哈希值
		},
		{
			name:     "Simple string",                                                    // 简单字符串测试
			input:    "hello",                                                            // 输入为"hello"
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", // 修正后的SHA256哈希值
		},
		{
			name:     "String with spaces",                                               // 包含空格的字符串测试
			input:    "hello world",                                                      // 输入为"hello world"
			expected: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", // 标准SHA256哈希值
		},
		{
			name:     "Special characters",                                               // 特殊字符测试
			input:    "!@#$%^&*()",                                                       // 输入为特殊字符
			expected: "95ce789c5c9d18490972709838ca3a9719094bca3ac16332cfec0652b0236141", // 修正后的SHA256哈希值
		},
		{
			name:     "Chinese characters",                                               // 中文字符测试
			input:    "你好世界",                                                             // 输入为中文字符串
			expected: "beca6335b20ff57ccc47403ef4d9e0b8fccb4442b3151c2e7d50050673d43172", // 修正后的SHA256哈希值
		},
		{
			name:     "Long string",                                                                                           // 长字符串测试
			input:    "This is a very long string used to test the SHA256 function implementation in Go programming language", // 长输入字符串
			expected: "a14f40455e20de553bc4ad75a3ac7c91b928ca45cc8134c00f5e0e0f10019754",                                      // 修正后的SHA256哈希值
		},
		{
			name:     "Numeric string",                                                   // 数字字符串测试
			input:    "123456789",                                                        // 输入为数字字符串
			expected: "15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225", // 标准SHA256哈希值
		},
	}

	// 执行测试用例
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 调用被测函数
			result := SHA256(tt.input)

			// 验证结果是否符合预期
			if result != tt.expected {
				t.Errorf("SHA256(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSHA256Consistency 测试SHA256函数的一致性 - 相同输入应该产生相同输出
func TestSHA256Consistency(t *testing.T) {
	input := "consistency test string" // 测试字符串

	// 多次调用函数验证一致性
	result1 := SHA256(input)
	result2 := SHA256(input)
	result3 := SHA256(input)

	// 验证所有结果都相同
	if result1 != result2 || result2 != result3 {
		t.Errorf("SHA256 function is not consistent: got %q, %q, %q", result1, result2, result3)
	}
}

// TestSHA256AgainstStandardImplementation 测试与标准库实现的结果一致性
func TestSHA256AgainstStandardImplementation(t *testing.T) {
	testString := "test against standard implementation" // 测试字符串

	// 使用我们实现的函数计算哈希值
	ourResult := SHA256(testString)

	// 使用标准库直接计算哈希值作为对比
	h := sha256.New()
	h.Write([]byte(testString))
	standardResult := hex.EncodeToString(h.Sum(nil))

	// 验证两者结果一致
	if ourResult != standardResult {
		t.Errorf("Our implementation differs from standard library: got %q, want %q", ourResult, standardResult)
	}
}

// BenchmarkSHA256 基准测试 - 测试SHA256函数性能
func BenchmarkSHA256(b *testing.B) {
	testString := "benchmark test string" // 基准测试使用的字符串

	// 重置计时器
	b.ResetTimer()

	// 运行基准测试
	for i := 0; i < b.N; i++ {
		SHA256(testString)
	}
}

// ExampleSHA256 示例函数 - 展示如何使用SHA256函数
func ExampleSHA256() {
	result := SHA256("hello")
	fmt.Println(result)
	// Output: 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
}
