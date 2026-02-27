package strutil_test

import (
	"strings"
	"testing"

	"github.com/aide-family/magicbox/strutil"
)

// TestTitle 测试 Title 函数
func TestTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{
			name:     "No arguments",
			input:    []string{},
			expected: "",
		},
		{
			name:     "Single word",
			input:    []string{"hello"},
			expected: "Hello",
		},
		{
			name:     "Single word already capitalized",
			input:    []string{"Hello"},
			expected: "Hello",
		},
		{
			name:     "Multiple words",
			input:    []string{"hello", "world"},
			expected: "Hello World",
		},
		{
			name:     "Multiple words with different cases",
			input:    []string{"hello", "WORLD", "test"},
			expected: "Hello World Test",
		},
		{
			name:     "Single word uppercase",
			input:    []string{"HELLO"},
			expected: "Hello",
		},
		{
			name:     "Empty string",
			input:    []string{""},
			expected: "",
		},
		{
			name:     "Multiple empty strings",
			input:    []string{"", "", ""},
			expected: "  ",
		},
		{
			name:     "String with spaces",
			input:    []string{"hello world"},
			expected: "Hello World",
		},
		{
			name:     "String with multiple spaces",
			input:    []string{"hello  world"},
			expected: "Hello  World",
		},
		{
			name:     "String with leading spaces",
			input:    []string{"  hello"},
			expected: "  Hello",
		},
		{
			name:     "String with trailing spaces",
			input:    []string{"hello  "},
			expected: "Hello  ",
		},
		{
			name:     "Three words",
			input:    []string{"hello", "world", "test"},
			expected: "Hello World Test",
		},
		{
			name:     "Single character",
			input:    []string{"a"},
			expected: "A",
		},
		{
			name:     "Multiple single characters",
			input:    []string{"a", "b", "c"},
			expected: "A B C",
		},
		{
			name:     "Words with numbers",
			input:    []string{"hello", "123", "world"},
			expected: "Hello 123 World",
		},
		{
			name:     "Words with special characters",
			input:    []string{"hello", "world!", "test"},
			expected: "Hello World! Test",
		},
		{
			name:     "Mixed case words",
			input:    []string{"hELLo", "WoRLd"},
			expected: "Hello World",
		},
		{
			name:     "All uppercase",
			input:    []string{"HELLO", "WORLD"},
			expected: "Hello World",
		},
		{
			name:     "All lowercase",
			input:    []string{"hello", "world"},
			expected: "Hello World",
		},
		{
			name:     "Single long word",
			input:    []string{"hello world test"},
			expected: "Hello World Test",
		},
		{
			name:     "Words with tabs",
			input:    []string{"hello\tworld"},
			expected: "Hello\tWorld",
		},
		{
			name:     "Words with newlines",
			input:    []string{"hello\nworld"},
			expected: "Hello\nWorld",
		},
		{
			name:     "One empty string in middle",
			input:    []string{"hello", "", "world"},
			expected: "Hello  World",
		},
		{
			name:     "Multiple words with empty strings",
			input:    []string{"hello", "", "world", "", "test"},
			expected: "Hello  World  Test",
		},
		{
			name:     "String with only spaces",
			input:    []string{"   "},
			expected: "   ",
		},
		{
			name:     "String with only tabs",
			input:    []string{"\t\t\t"},
			expected: "\t\t\t",
		},
		{
			name:     "Very long string",
			input:    []string{"this is a very long string that should be converted to title case"},
			expected: "This Is A Very Long String That Should Be Converted To Title Case",
		},
		{
			name:     "Many words",
			input:    []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"},
			expected: "One Two Three Four Five Six Seven Eight Nine Ten",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.Title(tt.input...)
			if result != tt.expected {
				t.Errorf("Title(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestTitle_VariadicArgs 测试 Title 函数的可变参数
func TestTitle_VariadicArgs(t *testing.T) {
	// 测试无参数
	result := strutil.Title()
	if result != "" {
		t.Errorf("Title() = %q, want %q", result, "")
	}

	// 测试单个参数
	result = strutil.Title("hello")
	if result != "Hello" {
		t.Errorf("Title(\"hello\") = %q, want %q", result, "Hello")
	}

	// 测试多个参数
	result = strutil.Title("hello", "world", "test")
	if result != "Hello World Test" {
		t.Errorf("Title(\"hello\", \"world\", \"test\") = %q, want %q", result, "Hello World Test")
	}

	// 测试很多参数
	result = strutil.Title("a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
	expected := "A B C D E F G H I J"
	if result != expected {
		t.Errorf("Title(10 args) = %q, want %q", result, expected)
	}
}

// TestIsEmpty 测试 IsEmpty 函数
func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "String with only spaces",
			input:    "   ",
			expected: true,
		},
		{
			name:     "String with only tabs",
			input:    "\t\t\t",
			expected: true,
		},
		{
			name:     "String with only newlines",
			input:    "\n\n\n",
			expected: true,
		},
		{
			name:     "String with mixed whitespace",
			input:    " \t\n \t\n ",
			expected: true,
		},
		{
			name:     "String with leading spaces",
			input:    "  hello",
			expected: false,
		},
		{
			name:     "String with trailing spaces",
			input:    "hello  ",
			expected: false,
		},
		{
			name:     "String with both leading and trailing spaces",
			input:    "  hello  ",
			expected: false,
		},
		{
			name:     "Non-empty string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "String with content and spaces",
			input:    "hello world",
			expected: false,
		},
		{
			name:     "String with numbers",
			input:    "123",
			expected: false,
		},
		{
			name:     "String with special characters",
			input:    "!@#$",
			expected: false,
		},
		{
			name:     "Single character",
			input:    "a",
			expected: false,
		},
		{
			name:     "Single space",
			input:    " ",
			expected: true,
		},
		{
			name:     "Single tab",
			input:    "\t",
			expected: true,
		},
		{
			name:     "Single newline",
			input:    "\n",
			expected: true,
		},
		{
			name:     "String with zero-width space",
			input:    "\u200B",
			expected: false, // Zero-width space is not considered whitespace by TrimSpace
		},
		{
			name:     "String with unicode spaces",
			input:    "\u00A0\u2000\u2001", // Non-breaking space, en quad, em quad
			expected: true,
		},
		{
			name:     "Very long string",
			input:    "this is a very long string",
			expected: false,
		},
		{
			name:     "String with Chinese characters",
			input:    "你好",
			expected: false,
		},
		{
			name:     "String with mixed content",
			input:    "hello 123 world!",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.IsEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmpty(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsNotEmpty 测试 IsNotEmpty 函数
func TestIsNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "String with only spaces",
			input:    "   ",
			expected: false,
		},
		{
			name:     "String with only tabs",
			input:    "\t\t\t",
			expected: false,
		},
		{
			name:     "String with only newlines",
			input:    "\n\n\n",
			expected: false,
		},
		{
			name:     "String with mixed whitespace",
			input:    " \t\n \t\n ",
			expected: false,
		},
		{
			name:     "String with leading spaces",
			input:    "  hello",
			expected: true,
		},
		{
			name:     "String with trailing spaces",
			input:    "hello  ",
			expected: true,
		},
		{
			name:     "String with both leading and trailing spaces",
			input:    "  hello  ",
			expected: true,
		},
		{
			name:     "Non-empty string",
			input:    "hello",
			expected: true,
		},
		{
			name:     "String with content and spaces",
			input:    "hello world",
			expected: true,
		},
		{
			name:     "String with numbers",
			input:    "123",
			expected: true,
		},
		{
			name:     "String with special characters",
			input:    "!@#$",
			expected: true,
		},
		{
			name:     "Single character",
			input:    "a",
			expected: true,
		},
		{
			name:     "Single space",
			input:    " ",
			expected: false,
		},
		{
			name:     "Single tab",
			input:    "\t",
			expected: false,
		},
		{
			name:     "Single newline",
			input:    "\n",
			expected: false,
		},
		{
			name:     "String with zero-width space",
			input:    "\u200B",
			expected: true, // Zero-width space is not considered whitespace by TrimSpace
		},
		{
			name:     "String with unicode spaces",
			input:    "\u00A0\u2000\u2001", // Non-breaking space, en quad, em quad
			expected: false,
		},
		{
			name:     "Very long string",
			input:    "this is a very long string",
			expected: true,
		},
		{
			name:     "String with Chinese characters",
			input:    "你好",
			expected: true,
		},
		{
			name:     "String with mixed content",
			input:    "hello 123 world!",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.IsNotEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("IsNotEmpty(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsEmpty_IsNotEmpty_Complementary 测试 IsEmpty 和 IsNotEmpty 是互补的
func TestIsEmpty_IsNotEmpty_Complementary(t *testing.T) {
	testStrings := []string{
		"",
		"   ",
		"\t\t\t",
		"\n\n\n",
		" \t\n \t\n ",
		"hello",
		"  hello",
		"hello  ",
		"  hello  ",
		"hello world",
		"123",
		"!@#$",
		"a",
		" ",
		"\t",
		"\n",
		"this is a very long string",
		"你好",
		"hello 123 world!",
		"a\nb\tc d",
	}

	for _, s := range testStrings {
		t.Run(s, func(t *testing.T) {
			isEmpty := strutil.IsEmpty(s)
			isNotEmpty := strutil.IsNotEmpty(s)

			// 验证它们是互补的
			if isEmpty == isNotEmpty {
				t.Errorf("IsEmpty(%q) = %v, IsNotEmpty(%q) = %v, they should be opposite", s, isEmpty, s, isNotEmpty)
			}

			// 验证逻辑关系
			if isEmpty && isNotEmpty {
				t.Errorf("IsEmpty(%q) and IsNotEmpty(%q) cannot both be true", s, s)
			}

			if !isEmpty && !isNotEmpty {
				t.Errorf("IsEmpty(%q) and IsNotEmpty(%q) cannot both be false", s, s)
			}
		})
	}
}

// TestTitle_EdgeCases 测试 Title 函数的边界情况
func TestTitle_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		validate func(string) bool
	}{
		{
			name:  "Nil slice (no args)",
			input: nil,
			validate: func(s string) bool {
				return s == ""
			},
		},
		{
			name:  "Single character",
			input: []string{"a"},
			validate: func(s string) bool {
				return s == "A"
			},
		},
		{
			name:  "Single uppercase character",
			input: []string{"A"},
			validate: func(s string) bool {
				return s == "A"
			},
		},
		{
			name:  "Single number",
			input: []string{"1"},
			validate: func(s string) bool {
				return s == "1"
			},
		},
		{
			name:  "Single special character",
			input: []string{"!"},
			validate: func(s string) bool {
				return s == "!"
			},
		},
		{
			name:  "Very long single word",
			input: []string{strings.Repeat("a", 1000)},
			validate: func(s string) bool {
				return len(s) == 1000 && s[0] == 'A' && s[1:] == strings.Repeat("a", 999)
			},
		},
		{
			name:  "Many empty strings",
			input: []string{"", "", "", "", ""},
			validate: func(s string) bool {
				return s == "    "
			},
		},
		{
			name:  "Mixed empty and non-empty",
			input: []string{"hello", "", "world", "", "test"},
			validate: func(s string) bool {
				return s == "Hello  World  Test"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.input == nil {
				result = strutil.Title()
			} else {
				result = strutil.Title(tt.input...)
			}

			if !tt.validate(result) {
				t.Errorf("Title(%v) = %q, validation failed", tt.input, result)
			}
		})
	}
}

// TestIsEmpty_Consistency 测试 IsEmpty 的一致性
func TestIsEmpty_Consistency(t *testing.T) {
	testCases := []string{
		"",
		"   ",
		"hello",
		"  hello  ",
	}

	for _, s := range testCases {
		t.Run(s, func(t *testing.T) {
			// 多次调用应该返回相同结果
			results := make([]bool, 10)
			for i := 0; i < 10; i++ {
				results[i] = strutil.IsEmpty(s)
			}

			first := results[0]
			for i, result := range results {
				if result != first {
					t.Errorf("IsEmpty(%q) returned inconsistent value: result[0] = %v, result[%d] = %v", s, first, i, result)
				}
			}
		})
	}
}

// TestIsNotEmpty_Consistency 测试 IsNotEmpty 的一致性
func TestIsNotEmpty_Consistency(t *testing.T) {
	testCases := []string{
		"",
		"   ",
		"hello",
		"  hello  ",
	}

	for _, s := range testCases {
		t.Run(s, func(t *testing.T) {
			// 多次调用应该返回相同结果
			results := make([]bool, 10)
			for i := 0; i < 10; i++ {
				results[i] = strutil.IsNotEmpty(s)
			}

			first := results[0]
			for i, result := range results {
				if result != first {
					t.Errorf("IsNotEmpty(%q) returned inconsistent value: result[0] = %v, result[%d] = %v", s, first, i, result)
				}
			}
		})
	}
}

// TestTitle_Unicode 测试 Title 函数处理 Unicode 字符
func TestTitle_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		validate func(string) bool
	}{
		{
			name:  "Chinese characters",
			input: []string{"你好", "世界"},
			validate: func(s string) bool {
				return strings.Contains(s, "你好") && strings.Contains(s, "世界")
			},
		},
		{
			name:  "Mixed English and Chinese",
			input: []string{"hello", "世界"},
			validate: func(s string) bool {
				return strings.Contains(s, "Hello") && strings.Contains(s, "世界")
			},
		},
		{
			name:  "Emoji characters",
			input: []string{"hello", "😀", "world"},
			validate: func(s string) bool {
				return strings.Contains(s, "Hello") && strings.Contains(s, "😀") && strings.Contains(s, "World")
			},
		},
		{
			name:  "Russian characters",
			input: []string{"привет", "мир"},
			validate: func(s string) bool {
				return strings.Contains(s, "Привет") && strings.Contains(s, "Мир")
			},
		},
		{
			name:  "Greek characters",
			input: []string{"γεια", "σου"},
			validate: func(s string) bool {
				return strings.Contains(s, "Γεια") && strings.Contains(s, "Σου")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.Title(tt.input...)
			if !tt.validate(result) {
				t.Errorf("Title(%v) = %q, validation failed", tt.input, result)
			}
		})
	}
}

// BenchmarkTitle 基准测试 Title 函数
func BenchmarkTitle(b *testing.B) {
	testStrings := []string{"hello", "world", "test"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.Title(testStrings...)
	}
}

// BenchmarkTitle_SingleWord 基准测试 Title 函数（单个词）
func BenchmarkTitle_SingleWord(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.Title("hello")
	}
}

// BenchmarkTitle_ManyWords 基准测试 Title 函数（多个词）
func BenchmarkTitle_ManyWords(b *testing.B) {
	testStrings := []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.Title(testStrings...)
	}
}

// BenchmarkTitle_LongString 基准测试 Title 函数（长字符串）
func BenchmarkTitle_LongString(b *testing.B) {
	longString := "this is a very long string that should be converted to title case"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.Title(longString)
	}
}

// BenchmarkIsEmpty 基准测试 IsEmpty 函数
func BenchmarkIsEmpty(b *testing.B) {
	testStrings := []string{"", "   ", "hello", "  hello  "}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			_ = strutil.IsEmpty(s)
		}
	}
}

// BenchmarkIsNotEmpty 基准测试 IsNotEmpty 函数
func BenchmarkIsNotEmpty(b *testing.B) {
	testStrings := []string{"", "   ", "hello", "  hello  "}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			_ = strutil.IsNotEmpty(s)
		}
	}
}

// BenchmarkIsEmpty_EmptyString 基准测试 IsEmpty 函数（空字符串）
func BenchmarkIsEmpty_EmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.IsEmpty("")
	}
}

// BenchmarkIsEmpty_WhitespaceOnly 基准测试 IsEmpty 函数（仅空白字符）
func BenchmarkIsEmpty_WhitespaceOnly(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.IsEmpty("   ")
	}
}

// BenchmarkIsEmpty_NonEmptyString 基准测试 IsEmpty 函数（非空字符串）
func BenchmarkIsEmpty_NonEmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.IsEmpty("hello")
	}
}

// BenchmarkIsNotEmpty_EmptyString 基准测试 IsNotEmpty 函数（空字符串）
func BenchmarkIsNotEmpty_EmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.IsNotEmpty("")
	}
}

// BenchmarkIsNotEmpty_WhitespaceOnly 基准测试 IsNotEmpty 函数（仅空白字符）
func BenchmarkIsNotEmpty_WhitespaceOnly(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.IsNotEmpty("   ")
	}
}

// BenchmarkIsNotEmpty_NonEmptyString 基准测试 IsNotEmpty 函数（非空字符串）
func BenchmarkIsNotEmpty_NonEmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.IsNotEmpty("hello")
	}
}

// TestSplitSkipEmpty 测试 SplitSkipEmpty 函数
func TestSplitSkipEmpty(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		expected []string
	}{
		{
			name:     "Empty string",
			s:        "",
			sep:      ",",
			expected: nil,
		},
		{
			name:     "Simple split",
			s:        "a,b,c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Split with empty strings in middle",
			s:        "a,,b,,c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Split with empty string at start",
			s:        ",a,b,c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Split with empty string at end",
			s:        "a,b,c,",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Split with empty strings at both ends",
			s:        ",a,b,c,",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Only empty strings",
			s:        ",,,",
			sep:      ",",
			expected: []string{},
		},
		{
			name:     "Single character separator",
			s:        "a|b|c",
			sep:      "|",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Multi-character separator",
			s:        "a||b||c",
			sep:      "||",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Space separator",
			s:        "a b c",
			sep:      " ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Space separator with multiple spaces",
			s:        "a  b  c",
			sep:      " ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Tab separator",
			s:        "a\tb\tc",
			sep:      "\t",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Newline separator",
			s:        "a\nb\nc",
			sep:      "\n",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Single element",
			s:        "a",
			sep:      ",",
			expected: []string{"a"},
		},
		{
			name:     "No separator in string",
			s:        "abc",
			sep:      ",",
			expected: []string{"abc"},
		},
		{
			name:     "Separator not found",
			s:        "a,b,c",
			sep:      "|",
			expected: []string{"a,b,c"},
		},
		{
			name:     "Empty separator",
			s:        "abc",
			sep:      "",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "String with numbers",
			s:        "1,2,3,4,5",
			sep:      ",",
			expected: []string{"1", "2", "3", "4", "5"},
		},
		{
			name:     "String with special characters",
			s:        "a!b!c",
			sep:      "!",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "String with unicode characters",
			s:        "你好,世界,测试",
			sep:      ",",
			expected: []string{"你好", "世界", "测试"},
		},
		{
			name:     "String with mixed content and empty strings",
			s:        "a,,b, ,c,,",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Long string",
			s:        "a,b,c,d,e,f,g,h,i,j",
			sep:      ",",
			expected: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		},
		{
			name:     "String with only separator",
			s:        ",",
			sep:      ",",
			expected: []string{},
		},
		{
			name:     "String with multiple consecutive separators",
			s:        "a,,,b,,,c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "String with leading and trailing spaces in elements",
			s:        " a , b , c ",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "String with spaces around separators",
			s:        "a , b , c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "String with tabs and spaces",
			s:        "a\t,\tb\t,\tc",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "String with only whitespace segments",
			s:        " ,  ,   ",
			sep:      ",",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.SplitSkipEmpty(tt.s, tt.sep)
			if !equalStringSlice(result, tt.expected) {
				t.Errorf("SplitSkipEmpty(%q, %q) = %v, want %v", tt.s, tt.sep, result, tt.expected)
			}
		})
	}
}

// equalStringSlice compares two string slices for equality
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSplitSkipEmpty_EdgeCases 测试 SplitSkipEmpty 函数的边界情况
func TestSplitSkipEmpty_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		sep      string
		validate func([]string) bool
	}{
		{
			name: "Very long string",
			s:    strings.Repeat("a,", 1000) + "b",
			sep:  ",",
			validate: func(result []string) bool {
				return len(result) == 1001 && result[0] == "a" && result[1000] == "b"
			},
		},
		{
			name: "String with whitespace-only segments",
			s:    "a, ,b,  ,c",
			sep:  ",",
			validate: func(result []string) bool {
				// Whitespace-only strings are trimmed and filtered out
				return len(result) == 3 && result[0] == "a" && result[1] == "b" && result[2] == "c"
			},
		},
		{
			name: "String with tab and newline",
			s:    "a\tb\nc",
			sep:  "\t",
			validate: func(result []string) bool {
				return len(result) == 2 && result[0] == "a" && result[1] == "b\nc"
			},
		},
		{
			name: "Unicode separator",
			s:    "a你好b你好c",
			sep:  "你好",
			validate: func(result []string) bool {
				return len(result) == 3 && result[0] == "a" && result[1] == "b" && result[2] == "c"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strutil.SplitSkipEmpty(tt.s, tt.sep)
			if !tt.validate(result) {
				t.Errorf("SplitSkipEmpty(%q, %q) = %v, validation failed", tt.s, tt.sep, result)
			}
		})
	}
}

// TestSplitSkipEmpty_Consistency 测试 SplitSkipEmpty 的一致性
func TestSplitSkipEmpty_Consistency(t *testing.T) {
	testCases := []struct {
		s   string
		sep string
	}{
		{"a,b,c", ","},
		{"a,,b,,c", ","},
		{"", ","},
		{"a", ","},
		{"a,b", ","},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			// 多次调用应该返回相同结果
			results := make([][]string, 10)
			for i := 0; i < 10; i++ {
				results[i] = strutil.SplitSkipEmpty(tc.s, tc.sep)
			}

			first := results[0]
			for i, result := range results {
				if !equalStringSlice(result, first) {
					t.Errorf("SplitSkipEmpty(%q, %q) returned inconsistent value: result[0] = %v, result[%d] = %v", tc.s, tc.sep, first, i, result)
				}
			}
		})
	}
}

// BenchmarkSplitSkipEmpty 基准测试 SplitSkipEmpty 函数
func BenchmarkSplitSkipEmpty(b *testing.B) {
	testString := "a,b,c,d,e,f,g,h,i,j"
	sep := ","
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.SplitSkipEmpty(testString, sep)
	}
}

// BenchmarkSplitSkipEmpty_WithEmptyStrings 基准测试 SplitSkipEmpty 函数（包含空字符串）
func BenchmarkSplitSkipEmpty_WithEmptyStrings(b *testing.B) {
	testString := "a,,b,,c,,d,,e"
	sep := ","
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.SplitSkipEmpty(testString, sep)
	}
}

// BenchmarkSplitSkipEmpty_LongString 基准测试 SplitSkipEmpty 函数（长字符串）
func BenchmarkSplitSkipEmpty_LongString(b *testing.B) {
	testString := strings.Repeat("a,", 1000) + "b"
	sep := ","
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.SplitSkipEmpty(testString, sep)
	}
}

// BenchmarkSplitSkipEmpty_EmptyString 基准测试 SplitSkipEmpty 函数（空字符串）
func BenchmarkSplitSkipEmpty_EmptyString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strutil.SplitSkipEmpty("", ",")
	}
}
