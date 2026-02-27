package strutil_test

import (
	"testing"

	"github.com/aide-family/magicbox/strutil"
)

// TestMaskString 测试 MaskString 函数
func TestMaskString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			start:    3,
			end:      2,
			expected: "",
		},
		{
			name:     "Normal case - example from doc",
			input:    "1234567890",
			start:    3,
			end:      2,
			expected: "123******90",
		},
		{
			name:     "Normal case - email example",
			input:    "abc@example.com",
			start:    3,
			end:      4,
			expected: "abc******.com",
		},
		{
			name:     "Short string - too short to mask",
			input:    "123",
			start:    3,
			end:      2,
			expected: "123******",
		},
		{
			name:     "Short string - exactly start+end",
			input:    "12345",
			start:    3,
			end:      2,
			expected: "12345******",
		},
		{
			name:     "Single character",
			input:    "a",
			start:    1,
			end:      1,
			expected: "a******",
		},
		{
			name:     "Two characters",
			input:    "ab",
			start:    1,
			end:      1,
			expected: "ab******",
		},
		{
			name:     "Three characters - can mask",
			input:    "abc",
			start:    1,
			end:      1,
			expected: "a******c",
		},
		{
			name:     "Four characters - can mask",
			input:    "abcd",
			start:    1,
			end:      1,
			expected: "a******d",
		},
		{
			name:     "Long string",
			input:    "abcdefghijklmnopqrstuvwxyz",
			start:    5,
			end:      5,
			expected: "abcde******vwxyz",
		},
		{
			name:     "Start is 0",
			input:    "1234567890",
			start:    0,
			end:      2,
			expected: "******90",
		},
		{
			name:     "End is 0",
			input:    "1234567890",
			start:    3,
			end:      0,
			expected: "123******",
		},
		{
			name:     "Both start and end are 0",
			input:    "1234567890",
			start:    0,
			end:      0,
			expected: "******",
		},
		{
			name:     "Negative start - should be treated as 0",
			input:    "1234567890",
			start:    -1,
			end:      2,
			expected: "******90",
		},
		{
			name:     "Negative end - should be treated as 0",
			input:    "1234567890",
			start:    3,
			end:      -1,
			expected: "123******",
		},
		{
			name:     "Both negative - should be treated as 0",
			input:    "1234567890",
			start:    -1,
			end:      -1,
			expected: "******",
		},
		{
			name:     "Large start and end values",
			input:    "1234567890",
			start:    10,
			end:      10,
			expected: "1234567890******",
		},
		{
			name:     "Chinese characters",
			input:    "你好世界测试",
			start:    2,
			end:      2,
			expected: "你好******测试",
		},
		{
			name:     "Mixed Chinese and English",
			input:    "hello世界",
			start:    3,
			end:      2,
			expected: "hel******世界",
		},
		{
			name:     "Special characters",
			input:    "!@#$%^&*()",
			start:    2,
			end:      2,
			expected: "!@******()",
		},
		{
			name:     "String with spaces",
			input:    "hello world",
			start:    3,
			end:      3,
			expected: "hel******rld",
		},
		{
			name:     "String with newlines",
			input:    "line1\nline2",
			start:    3,
			end:      3,
			expected: "lin******ne2",
		},
		{
			name:     "Emoji characters",
			input:    "hello😀world",
			start:    3,
			end:      3,
			expected: "hel******rld",
		},
		{
			name:     "Unicode mixed",
			input:    "aαβγδε",
			start:    2,
			end:      2,
			expected: "aα******δε",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.MaskString(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("MaskString(%q, %d, %d) = %q, want %q", tt.input, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestMaskEmail 测试 MaskEmail 函数
func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal email - example from doc",
			input:    "user@example.com",
			expected: "u******r@example.com",
		},
		{
			name:     "Short local part - single character",
			input:    "a@example.com",
			expected: "a******@example.com",
		},
		{
			name:     "Short local part - two characters",
			input:    "ab@example.com",
			expected: "ab******@example.com",
		},
		{
			name:     "Short local part - three characters",
			input:    "abc@example.com",
			expected: "a******c@example.com",
		},
		{
			name:     "Four character local part",
			input:    "user@example.com",
			expected: "u******r@example.com",
		},
		{
			name:     "Long local part",
			input:    "verylongusername@example.com",
			expected: "v******e@example.com",
		},
		{
			name:     "Email with subdomain",
			input:    "user@mail.example.com",
			expected: "u******r@mail.example.com",
		},
		{
			name:     "Email with special characters in local part",
			input:    "user.name+tag@example.com",
			expected: "u******g@example.com",
		},
		{
			name:     "Email with numbers",
			input:    "user123@example.com",
			expected: "u******3@example.com",
		},
		{
			name:     "Email with Chinese characters",
			input:    "用户@example.com",
			expected: "用户******@example.com",
		},
		{
			name:     "Email without @ symbol",
			input:    "notanemail",
			expected: "notanemail******",
		},
		{
			name:     "Email starting with @",
			input:    "@example.com",
			expected: "@example.com******",
		},
		{
			name:     "Email with only @",
			input:    "@",
			expected: "@******",
		},
		{
			name:     "Email with multiple @ symbols",
			input:    "user@domain@example.com",
			expected: "u******r@domain@example.com",
		},
		{
			name:     "Email with domain only",
			input:    "example.com",
			expected: "example.com******",
		},
		{
			name:     "Email with space in local part",
			input:    "user name@example.com",
			expected: "u******e@example.com",
		},
		{
			name:     "Email with special domain",
			input:    "user@test.co.uk",
			expected: "u******r@test.co.uk",
		},
		{
			name:     "Email with IP address domain",
			input:    "user@192.168.1.1",
			expected: "u******r@192.168.1.1",
		},
		{
			name:     "Very long email",
			input:    "verylonglocalpartname@verylongdomainname.example.com",
			expected: "v******e@verylongdomainname.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.MaskEmail(tt.input)
			if result != tt.expected {
				t.Errorf("MaskEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMaskPhone 测试 MaskPhone 函数
func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal phone - example from doc",
			input:    "13812345678",
			expected: "138******5678",
		},
		{
			name:     "Short phone - too short",
			input:    "123",
			expected: "123******",
		},
		{
			name:     "Short phone - exactly 7 digits",
			input:    "1234567",
			expected: "1234567******",
		},
		{
			name:     "Phone with 8 digits",
			input:    "12345678",
			expected: "123******5678",
		},
		{
			name:     "Phone with 9 digits",
			input:    "123456789",
			expected: "123******6789",
		},
		{
			name:     "Phone with 10 digits",
			input:    "1234567890",
			expected: "123******7890",
		},
		{
			name:     "Phone with 11 digits - standard",
			input:    "13812345678",
			expected: "138******5678",
		},
		{
			name:     "Phone with 12 digits",
			input:    "138123456789",
			expected: "138******6789",
		},
		{
			name:     "Phone with 13 digits",
			input:    "1381234567890",
			expected: "138******7890",
		},
		{
			name:     "Phone with country code",
			input:    "+8613812345678",
			expected: "+86******5678",
		},
		{
			name:     "Phone with dashes",
			input:    "138-1234-5678",
			expected: "138******5678",
		},
		{
			name:     "Phone with spaces",
			input:    "138 1234 5678",
			expected: "138******5678",
		},
		{
			name:     "Phone with parentheses",
			input:    "(138)12345678",
			expected: "(13******5678",
		},
		{
			name:     "Phone with letters - treated as string",
			input:    "138abc45678",
			expected: "138******5678",
		},
		{
			name:     "International phone",
			input:    "8613812345678",
			expected: "861******5678",
		},
		{
			name:     "Very long phone number",
			input:    "12345678901234567890",
			expected: "123******7890",
		},
		{
			name:     "Single digit",
			input:    "1",
			expected: "1******",
		},
		{
			name:     "Two digits",
			input:    "12",
			expected: "12******",
		},
		{
			name:     "Three digits",
			input:    "123",
			expected: "123******",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.MaskPhone(tt.input)
			if result != tt.expected {
				t.Errorf("MaskPhone(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMaskBankCard 测试 MaskBankCard 函数
func TestMaskBankCard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Normal bank card - example from doc",
			input:    "6225123412341234",
			expected: "6225******1234",
		},
		{
			name:     "Short card - too short",
			input:    "123",
			expected: "123******",
		},
		{
			name:     "Short card - exactly 8 digits",
			input:    "12345678",
			expected: "12345678******",
		},
		{
			name:     "Card with 9 digits",
			input:    "123456789",
			expected: "1234******6789",
		},
		{
			name:     "Card with 10 digits",
			input:    "1234567890",
			expected: "1234******7890",
		},
		{
			name:     "Card with 11 digits",
			input:    "12345678901",
			expected: "1234******8901",
		},
		{
			name:     "Card with 12 digits",
			input:    "123456789012",
			expected: "1234******9012",
		},
		{
			name:     "Card with 13 digits",
			input:    "1234567890123",
			expected: "1234******0123",
		},
		{
			name:     "Card with 14 digits",
			input:    "12345678901234",
			expected: "1234******1234",
		},
		{
			name:     "Card with 15 digits",
			input:    "123456789012345",
			expected: "1234******2345",
		},
		{
			name:     "Card with 16 digits - standard",
			input:    "6225123412341234",
			expected: "6225******1234",
		},
		{
			name:     "Card with 17 digits",
			input:    "62251234123412345",
			expected: "6225******2345",
		},
		{
			name:     "Card with 18 digits",
			input:    "622512341234123456",
			expected: "6225******3456",
		},
		{
			name:     "Card with 19 digits",
			input:    "6225123412341234567",
			expected: "6225******4567",
		},
		{
			name:     "Card with spaces",
			input:    "6225 1234 1234 1234",
			expected: "6225******1234",
		},
		{
			name:     "Card with dashes",
			input:    "6225-1234-1234-1234",
			expected: "6225******1234",
		},
		{
			name:     "Card with mixed format",
			input:    "6225 1234-1234 1234",
			expected: "6225******1234",
		},
		{
			name:     "Card with letters - treated as string",
			input:    "6225abcd12341234",
			expected: "6225******1234",
		},
		{
			name:     "Very long card number",
			input:    "123456789012345678901234567890",
			expected: "1234******7890",
		},
		{
			name:     "Single digit",
			input:    "1",
			expected: "1******",
		},
		{
			name:     "Two digits",
			input:    "12",
			expected: "12******",
		},
		{
			name:     "Three digits",
			input:    "123",
			expected: "123******",
		},
		{
			name:     "Four digits",
			input:    "1234",
			expected: "1234******",
		},
		{
			name:     "American Express format (15 digits)",
			input:    "378282246310005",
			expected: "3782******0005",
		},
		{
			name:     "Visa format (16 digits)",
			input:    "4111111111111111",
			expected: "4111******1111",
		},
		{
			name:     "MasterCard format (16 digits)",
			input:    "5555555555554444",
			expected: "5555******4444",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.MaskBankCard(tt.input)
			if result != tt.expected {
				t.Errorf("MaskBankCard(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestMaskString_UTF8 专门测试 UTF-8 字符处理
func TestMaskString_UTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Chinese characters",
			input:    "一二三四五六七八九十",
			start:    2,
			end:      2,
			expected: "一二******九十",
		},
		{
			name:     "Japanese characters",
			input:    "こんにちは世界",
			start:    2,
			end:      2,
			expected: "こん******世界",
		},
		{
			name:     "Korean characters",
			input:    "안녕하세요세계",
			start:    2,
			end:      2,
			expected: "안녕******세계",
		},
		{
			name:     "Arabic characters",
			input:    "مرحبا بالعالم",
			start:    3,
			end:      3,
			expected: "مرح******الم",
		},
		{
			name:     "Emoji characters",
			input:    "😀😃😄😁😆",
			start:    1,
			end:      1,
			expected: "😀******😆",
		},
		{
			name:     "Mixed UTF-8 characters",
			input:    "Hello世界😀",
			start:    3,
			end:      2,
			expected: "Hel******界😀",
		},
		{
			name:     "Russian characters",
			input:    "ПриветМир",
			start:    3,
			end:      3,
			expected: "При******Мир",
		},
		{
			name:     "Greek characters",
			input:    "ΓειασουΚοσμε",
			start:    4,
			end:      4,
			expected: "Γεια******οσμε",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.MaskString(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("MaskString(%q, %d, %d) = %q, want %q", tt.input, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestMaskString_EdgeCases 测试边界情况
func TestMaskString_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{
			name:     "Very large start value",
			input:    "1234567890",
			start:    100,
			end:      2,
			expected: "1234567890******",
		},
		{
			name:     "Very large end value",
			input:    "1234567890",
			start:    3,
			end:      100,
			expected: "1234567890******",
		},
		{
			name:     "Zero start and zero end",
			input:    "1234567890",
			start:    0,
			end:      0,
			expected: "******",
		},
		{
			name:     "Start equals length",
			input:    "12345",
			start:    5,
			end:      0,
			expected: "12345******",
		},
		{
			name:     "End equals length",
			input:    "12345",
			start:    0,
			end:      5,
			expected: "12345******",
		},
		{
			name:     "Start plus end equals length",
			input:    "12345",
			start:    3,
			end:      2,
			expected: "12345******",
		},
		{
			name:     "Start plus end exceeds length",
			input:    "12345",
			start:    4,
			end:      2,
			expected: "12345******",
		},
		{
			name:     "Unicode string with varying byte lengths",
			input:    "aαβγδε",
			start:    1,
			end:      1,
			expected: "a******ε",
		},
		{
			name:     "String with only spaces",
			input:    "     ",
			start:    2,
			end:      2,
			expected: "  ******  ",
		},
		{
			name:     "String with only asterisks",
			input:    "******",
			start:    2,
			end:      2,
			expected: "**********",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.MaskString(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("MaskString(%q, %d, %d) = %q, want %q", tt.input, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// BenchmarkMaskString 基准测试 MaskString 函数
func BenchmarkMaskString(b *testing.B) {
	testString := "This is a test string for benchmarking"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.MaskString(testString, 5, 5)
	}
}

// BenchmarkMaskEmail 基准测试 MaskEmail 函数
func BenchmarkMaskEmail(b *testing.B) {
	testEmail := "verylongusername@verylongdomainname.example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.MaskEmail(testEmail)
	}
}

// BenchmarkMaskPhone 基准测试 MaskPhone 函数
func BenchmarkMaskPhone(b *testing.B) {
	testPhone := "13812345678901"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.MaskPhone(testPhone)
	}
}

// BenchmarkMaskBankCard 基准测试 MaskBankCard 函数
func BenchmarkMaskBankCard(b *testing.B) {
	testCard := "622512341234123456"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.MaskBankCard(testCard)
	}
}

// BenchmarkMaskString_UTF8 基准测试 MaskString 函数处理 UTF-8 字符
func BenchmarkMaskString_UTF8(b *testing.B) {
	testString := "你好世界测试字符串用于性能测试"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.MaskString(testString, 3, 3)
	}
}
