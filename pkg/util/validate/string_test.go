package validate

import (
	"strings"
	"testing"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/stretchr/testify/assert"
)

func TestTextIsNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: true,
		},
		{
			name:     "tab only",
			input:    "\t",
			expected: true,
		},
		{
			name:     "newline only",
			input:    "\n",
			expected: true,
		},
		{
			name:     "mixed whitespace",
			input:    " \t\n\r ",
			expected: true,
		},
		{
			name:     "valid text",
			input:    "hello",
			expected: false,
		},
		{
			name:     "text with leading spaces",
			input:    "  hello",
			expected: false,
		},
		{
			name:     "text with trailing spaces",
			input:    "hello  ",
			expected: false,
		},
		{
			name:     "text with both spaces",
			input:    "  hello  ",
			expected: false,
		},
		{
			name:     "single character",
			input:    "a",
			expected: false,
		},
		{
			name:     "unicode text",
			input:    "你好世界",
			expected: false,
		},
		{
			name:     "unicode with spaces",
			input:    "  你好世界  ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TextIsNull(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %q", tt.input)
		})
	}
}

func TestTextIsNotNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "tab only",
			input:    "\t",
			expected: false,
		},
		{
			name:     "newline only",
			input:    "\n",
			expected: false,
		},
		{
			name:     "mixed whitespace",
			input:    " \t\n\r ",
			expected: false,
		},
		{
			name:     "valid text",
			input:    "hello",
			expected: true,
		},
		{
			name:     "text with leading spaces",
			input:    "  hello",
			expected: true,
		},
		{
			name:     "text with trailing spaces",
			input:    "hello  ",
			expected: true,
		},
		{
			name:     "text with both spaces",
			input:    "  hello  ",
			expected: true,
		},
		{
			name:     "single character",
			input:    "a",
			expected: true,
		},
		{
			name:     "unicode text",
			input:    "你好世界",
			expected: true,
		},
		{
			name:     "unicode with spaces",
			input:    "  你好世界  ",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TextIsNotNull(tt.input)
			assert.Equal(t, tt.expected, result, "Input: %q", tt.input)
		})
	}
}

func TestTextIsNullAndTextIsNotNullRelationship(t *testing.T) {
	// 测试 TextIsNull 和 TextIsNotNull 的互补关系
	testCases := []string{
		"",
		"   ",
		"\t",
		"\n",
		" \t\n\r ",
		"hello",
		"  hello",
		"hello  ",
		"  hello  ",
		"a",
		"你好世界",
		"  你好世界  ",
	}

	for _, tc := range testCases {
		t.Run("complementary_test", func(t *testing.T) {
			isNull := TextIsNull(tc)
			isNotNull := TextIsNotNull(tc)

			// TextIsNull 和 TextIsNotNull 应该是互补的
			assert.Equal(t, !isNull, isNotNull, "Input: %q, TextIsNull: %v, TextIsNotNull: %v", tc, isNull, isNotNull)
		})
	}
}

func TestCheckEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "valid email with plus",
			email:       "test+tag@example.com",
			expectError: false,
		},
		{
			name:        "valid email with dash",
			email:       "test-tag@example.com",
			expectError: false,
		},
		{
			name:        "valid email with dot",
			email:       "test.tag@example.com",
			expectError: false,
		},
		{
			name:        "valid email with multiple dots",
			email:       "test@example.co.uk",
			expectError: false,
		},
		{
			name:        "valid email with underscore",
			email:       "test_tag@example.com",
			expectError: false,
		},
		{
			name:        "valid email with numbers",
			email:       "test123@example.com",
			expectError: false,
		},
		{
			name:        "valid email with mixed case",
			email:       "Test@Example.com",
			expectError: false,
		},
		{
			name:        "empty email",
			email:       "",
			expectError: true,
		},
		{
			name:        "whitespace email",
			email:       "   ",
			expectError: true,
		},
		{
			name:        "missing @",
			email:       "testexample.com",
			expectError: true,
		},
		{
			name:        "missing domain",
			email:       "test@",
			expectError: true,
		},
		{
			name:        "missing local part",
			email:       "@example.com",
			expectError: true,
		},
		{
			name:        "invalid characters",
			email:       "test@#$%^@example.com",
			expectError: true,
		},
		{
			name:        "multiple @",
			email:       "test@example@com",
			expectError: true,
		},
		{
			name:        "domain starts with dash",
			email:       "test@-example.com",
			expectError: true,
		},
		{
			name:        "domain ends with dash",
			email:       "test@example-.com",
			expectError: true,
		},
		{
			name:        "local part starts with dot",
			email:       ".test@example.com",
			expectError: true,
		},
		{
			name:        "local part ends with dot",
			email:       "test.@example.com",
			expectError: true,
		},
		{
			name:        "consecutive dots in local",
			email:       "test..tag@example.com",
			expectError: true,
		},
		{
			name:        "consecutive dots in domain",
			email:       "test@example..com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckEmail(tt.email)
			if tt.expectError {
				assert.Error(t, err)
				if err != nil {
					assert.IsType(t, &errors.Error{}, err)
					kratosErr := err.(*errors.Error)
					assert.Equal(t, int32(400), kratosErr.Code)
					assert.Equal(t, "INVALID_EMAIL", kratosErr.Reason)
					assert.Equal(t, "The email format is incorrect.", kratosErr.Message)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid http url",
			url:         "http://example.com",
			expectError: false,
		},
		{
			name:        "valid https url",
			url:         "https://example.com",
			expectError: false,
		},
		{
			name:        "valid http url with path",
			url:         "http://example.com/path",
			expectError: false,
		},
		{
			name:        "valid https url with path",
			url:         "https://example.com/path",
			expectError: false,
		},
		{
			name:        "valid http url with query",
			url:         "http://example.com?param=value",
			expectError: false,
		},
		{
			name:        "valid https url with query",
			url:         "https://example.com?param=value",
			expectError: false,
		},
		{
			name:        "valid http url with port",
			url:         "http://example.com:8080",
			expectError: false,
		},
		{
			name:        "valid https url with port",
			url:         "https://example.com:8443",
			expectError: false,
		},
		{
			name:        "valid http url with subdomain",
			url:         "http://sub.example.com",
			expectError: false,
		},
		{
			name:        "valid https url with subdomain",
			url:         "https://sub.example.com",
			expectError: false,
		},
		{
			name:        "empty url",
			url:         "",
			expectError: true,
		},
		{
			name:        "whitespace url",
			url:         "   ",
			expectError: true,
		},
		{
			name:        "missing protocol",
			url:         "example.com",
			expectError: true,
		},
		{
			name:        "invalid protocol",
			url:         "ftp://example.com",
			expectError: true,
		},
		{
			name:        "protocol without colon",
			url:         "http//example.com",
			expectError: true,
		},
		{
			name:        "protocol without slashes",
			url:         "http:example.com",
			expectError: true,
		},
		{
			name:        "protocol with single slash",
			url:         "http:/example.com",
			expectError: true,
		},
		{
			name:        "protocol with three slashes",
			url:         "http:///example.com",
			expectError: false,
		},
		{
			name:        "uppercase protocol",
			url:         "HTTP://example.com",
			expectError: true,
		},
		{
			name:        "mixed case protocol",
			url:         "Http://example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckURL(tt.url)
			if tt.expectError {
				assert.Error(t, err)
				if err != nil {
					assert.IsType(t, &errors.Error{}, err)
					kratosErr := err.(*errors.Error)
					assert.Equal(t, int32(400), kratosErr.Code)
					assert.Equal(t, "INVALID_URL", kratosErr.Reason)
					assert.Equal(t, "The url format is incorrect.", kratosErr.Message)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 性能基准测试

func BenchmarkTextIsNull_Empty(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNull("")
	}
}

func BenchmarkTextIsNull_Whitespace(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNull("   ")
	}
}

func BenchmarkTextIsNull_ValidText(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNull("hello world")
	}
}

func BenchmarkTextIsNull_TextWithSpaces(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNull("  hello world  ")
	}
}

func BenchmarkTextIsNotNull_Empty(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNotNull("")
	}
}

func BenchmarkTextIsNotNull_Whitespace(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNotNull("   ")
	}
}

func BenchmarkTextIsNotNull_ValidText(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNotNull("hello world")
	}
}

func BenchmarkTextIsNotNull_TextWithSpaces(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = TextIsNotNull("  hello world  ")
	}
}

func BenchmarkCheckEmail_Valid(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckEmail("test@example.com")
	}
}

func BenchmarkCheckEmail_Invalid(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckEmail("invalid-email")
	}
}

func BenchmarkCheckEmail_Empty(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckEmail("")
	}
}

func BenchmarkCheckURL_ValidHTTP(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckURL("http://example.com")
	}
}

func BenchmarkCheckURL_ValidHTTPS(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckURL("https://example.com")
	}
}

func BenchmarkCheckURL_Invalid(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckURL("invalid-url")
	}
}

func BenchmarkCheckURL_Empty(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckURL("")
	}
}

// 复杂场景基准测试

func BenchmarkTextIsNull_LongText(t *testing.B) {
	longText := strings.Repeat("hello world ", 100)
	for i := 0; i < t.N; i++ {
		_ = TextIsNull(longText)
	}
}

func BenchmarkTextIsNull_LongTextWithSpaces(t *testing.B) {
	longText := "  " + strings.Repeat("hello world ", 100) + "  "
	for i := 0; i < t.N; i++ {
		_ = TextIsNull(longText)
	}
}

func BenchmarkCheckEmail_Complex(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckEmail("test.user+tag@subdomain.example.co.uk")
	}
}

func BenchmarkCheckURL_Complex(t *testing.B) {
	for i := 0; i < t.N; i++ {
		_ = CheckURL("https://subdomain.example.com:8443/path?param=value#fragment")
	}
}
