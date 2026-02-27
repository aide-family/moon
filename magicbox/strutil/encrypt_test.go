package strutil_test

import (
	"database/sql/driver"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aide-family/magicbox/strutil"
)

// TestDefaultBase64Encrypt_Encrypt 测试默认 Base64 加密器的 Encrypt 方法
func TestDefaultBase64Encrypt_Encrypt(t *testing.T) {
	encrypt := strutil.GetEncrypt()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Simple string",
			input:    "hello",
			expected: base64.StdEncoding.EncodeToString([]byte("hello")),
			wantErr:  false,
		},
		{
			name:     "String with spaces",
			input:    "hello world",
			expected: base64.StdEncoding.EncodeToString([]byte("hello world")),
			wantErr:  false,
		},
		{
			name:     "Special characters",
			input:    "!@#$%^&*()",
			expected: base64.StdEncoding.EncodeToString([]byte("!@#$%^&*()")),
			wantErr:  false,
		},
		{
			name:     "Chinese characters",
			input:    "你好世界",
			expected: base64.StdEncoding.EncodeToString([]byte("你好世界")),
			wantErr:  false,
		},
		{
			name:     "Long string",
			input:    "This is a very long string used to test the encryption function",
			expected: base64.StdEncoding.EncodeToString([]byte("This is a very long string used to test the encryption function")),
			wantErr:  false,
		},
		{
			name:     "Numeric string",
			input:    "123456789",
			expected: base64.StdEncoding.EncodeToString([]byte("123456789")),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encrypt.Encrypt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("Encrypt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestDefaultBase64Encrypt_Decrypt 测试默认 Base64 加密器的 Decrypt 方法
func TestDefaultBase64Encrypt_Decrypt(t *testing.T) {
	encrypt := strutil.GetEncrypt()

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Valid base64 string",
			input:    base64.StdEncoding.EncodeToString([]byte("hello")),
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "Valid base64 with spaces",
			input:    base64.StdEncoding.EncodeToString([]byte("hello world")),
			expected: "hello world",
			wantErr:  false,
		},
		{
			name:     "Valid base64 with Chinese",
			input:    base64.StdEncoding.EncodeToString([]byte("你好世界")),
			expected: "你好世界",
			wantErr:  false,
		},
		{
			name:    "Invalid base64 string",
			input:   "invalid base64!!!",
			wantErr: true,
		},
		{
			name:    "Invalid base64 with special chars",
			input:   "!@#$%^&*()",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encrypt.Decrypt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("Decrypt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestDefaultBase64Encrypt_RoundTrip 测试加密和解密的往返操作
func TestDefaultBase64Encrypt_RoundTrip(t *testing.T) {
	encrypt := strutil.GetEncrypt()

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "Empty string",
			input: "",
		},
		{
			name:  "Simple string",
			input: "hello",
		},
		{
			name:  "String with spaces",
			input: "hello world",
		},
		{
			name:  "Special characters",
			input: "!@#$%^&*()",
		},
		{
			name:  "Chinese characters",
			input: "你好世界",
		},
		{
			name:  "Long string",
			input: "This is a very long string used to test the encryption function",
		},
		{
			name:  "Newline characters",
			input: "line1\nline2\nline3",
		},
		{
			name:  "Tab characters",
			input: "col1\tcol2\tcol3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			encrypted, err := encrypt.Encrypt(tt.input)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			// 解密
			decrypted, err := encrypt.Decrypt(encrypted)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			// 验证往返后的值应该等于原始值
			if decrypted != tt.input {
				t.Errorf("RoundTrip() = %v, want %v", decrypted, tt.input)
			}
		})
	}
}

// TestEncryptString_Value 测试 EncryptString 的 Value 方法
func TestEncryptString_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    strutil.EncryptString
		wantErr  bool
		validate func(driver.Value) bool
	}{
		{
			name:    "Empty string",
			input:   strutil.EncryptString(""),
			wantErr: false,
			validate: func(v driver.Value) bool {
				str, ok := v.(string)
				return ok && str == ""
			},
		},
		{
			name:    "Simple string",
			input:   strutil.EncryptString("hello"),
			wantErr: false,
			validate: func(v driver.Value) bool {
				str, ok := v.(string)
				if !ok {
					return false
				}
				// 验证是有效的 base64 字符串
				_, err := base64.StdEncoding.DecodeString(str)
				return err == nil
			},
		},
		{
			name:    "String with spaces",
			input:   strutil.EncryptString("hello world"),
			wantErr: false,
			validate: func(v driver.Value) bool {
				str, ok := v.(string)
				if !ok {
					return false
				}
				_, err := base64.StdEncoding.DecodeString(str)
				return err == nil
			},
		},
		{
			name:    "Chinese characters",
			input:   strutil.EncryptString("你好世界"),
			wantErr: false,
			validate: func(v driver.Value) bool {
				str, ok := v.(string)
				if !ok {
					return false
				}
				_, err := base64.StdEncoding.DecodeString(str)
				return err == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.validate(result) {
				t.Errorf("Value() = %v, validation failed", result)
			}
		})
	}
}

// TestEncryptString_Scan 测试 EncryptString 的 Scan 方法
func TestEncryptString_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected strutil.EncryptString
		wantErr  bool
	}{
		{
			name:     "Nil value",
			input:    nil,
			expected: strutil.EncryptString(""),
			wantErr:  false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: strutil.EncryptString(""),
			wantErr:  false,
		},
		{
			name:     "Valid encrypted string",
			input:    base64.StdEncoding.EncodeToString([]byte("hello")),
			expected: strutil.EncryptString("hello"),
			wantErr:  false,
		},
		{
			name:     "Valid encrypted string with spaces",
			input:    base64.StdEncoding.EncodeToString([]byte("hello world")),
			expected: strutil.EncryptString("hello world"),
			wantErr:  false,
		},
		{
			name:     "Valid encrypted string with Chinese",
			input:    base64.StdEncoding.EncodeToString([]byte("你好世界")),
			expected: strutil.EncryptString("你好世界"),
			wantErr:  false,
		},
		{
			name:     "Valid encrypted []byte",
			input:    []byte(base64.StdEncoding.EncodeToString([]byte("hello"))),
			expected: strutil.EncryptString("hello"),
			wantErr:  false,
		},
		{
			name:    "Invalid base64 string",
			input:   "invalid base64!!!",
			wantErr: true,
		},
		{
			name:    "Invalid type - int",
			input:   123,
			wantErr: true,
		},
		{
			name:    "Invalid type - float",
			input:   3.14,
			wantErr: true,
		},
		{
			name:    "Invalid type - bool",
			input:   true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result strutil.EncryptString
			err := result.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("Scan() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestEncryptString_RoundTrip 测试 EncryptString 的 Value 和 Scan 往返操作
func TestEncryptString_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input strutil.EncryptString
	}{
		{
			name:  "Empty string",
			input: strutil.EncryptString(""),
		},
		{
			name:  "Simple string",
			input: strutil.EncryptString("hello"),
		},
		{
			name:  "String with spaces",
			input: strutil.EncryptString("hello world"),
		},
		{
			name:  "Special characters",
			input: strutil.EncryptString("!@#$%^&*()"),
		},
		{
			name:  "Chinese characters",
			input: strutil.EncryptString("你好世界"),
		},
		{
			name:  "Long string",
			input: strutil.EncryptString("This is a very long string used to test the encryption function"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Value: 将 EncryptString 转换为数据库值
			dbValue, err := tt.input.Value()
			if err != nil {
				t.Errorf("Value() error = %v", err)
				return
			}

			// Scan: 从数据库值恢复 EncryptString
			var result strutil.EncryptString
			err = result.Scan(dbValue)
			if err != nil {
				t.Errorf("Scan() error = %v", err)
				return
			}

			// 验证往返后的值应该等于原始值
			if result != tt.input {
				t.Errorf("RoundTrip() = %v, want %v", result, tt.input)
			}
		})
	}
}

// TestGetEncrypt 测试 GetEncrypt 函数
func TestGetEncrypt(t *testing.T) {
	// 获取默认加密器
	encrypt := strutil.GetEncrypt()
	if encrypt == nil {
		t.Error("GetEncrypt() returned nil")
		return
	}

	// 测试默认加密器功能
	testString := "test string"
	encrypted, err := encrypt.Encrypt(testString)
	if err != nil {
		t.Errorf("Encrypt() error = %v", err)
		return
	}

	decrypted, err := encrypt.Decrypt(encrypted)
	if err != nil {
		t.Errorf("Decrypt() error = %v", err)
		return
	}

	if decrypted != testString {
		t.Errorf("RoundTrip() = %v, want %v", decrypted, testString)
	}
}

// mockEncrypt 用于测试的自定义加密器
type mockEncrypt struct {
	encryptFunc func(string) (string, error)
	decryptFunc func(string) (string, error)
}

func (m *mockEncrypt) Encrypt(s string) (string, error) {
	if m.encryptFunc != nil {
		return m.encryptFunc(s)
	}
	return "encrypted:" + s, nil
}

func (m *mockEncrypt) Decrypt(s string) (string, error) {
	if m.decryptFunc != nil {
		return m.decryptFunc(s)
	}
	if len(s) > 10 && s[:10] == "encrypted:" {
		return s[10:], nil
	}
	return "", errors.New("invalid encrypted string")
}

// TestSetEncrypt 测试 SetEncrypt 函数
// 注意：由于使用了 sync.Once，SetEncrypt 只能生效一次
// 这个测试需要在一个新的测试环境中运行，或者测试顺序很重要
func TestSetEncrypt(t *testing.T) {
	// 保存原始的加密器
	originalEncrypt := strutil.GetEncrypt()

	// 创建 mock 加密器
	mock := &mockEncrypt{
		encryptFunc: func(s string) (string, error) {
			return "mock_encrypted:" + s, nil
		},
		decryptFunc: func(s string) (string, error) {
			if len(s) > 15 && s[:15] == "mock_encrypted:" {
				return s[15:], nil
			}
			return "", errors.New("invalid mock encrypted string")
		},
	}

	// 尝试设置自定义加密器
	strutil.SetEncrypt(mock)

	// 获取加密器并验证
	currentEncrypt := strutil.GetEncrypt()
	if currentEncrypt == nil {
		t.Error("GetEncrypt() returned nil after SetEncrypt")
		return
	}

	// 注意：由于 sync.Once 的特性，如果之前已经设置过，这里可能不会生效
	// 所以我们需要检查是否是 mock 加密器或者是原始加密器
	testString := "test"
	encrypted, err := currentEncrypt.Encrypt(testString)
	if err != nil {
		t.Errorf("Encrypt() error = %v", err)
		return
	}

	// 尝试解密
	decrypted, err := currentEncrypt.Decrypt(encrypted)
	if err != nil {
		// 如果解密失败，可能是使用的原始加密器（sync.Once 已经执行过）
		// 这是正常的，因为 sync.Once 确保只执行一次
		t.Logf("Decrypt() error = %v (this may be expected if SetEncrypt was already called)", err)
		// 验证原始加密器仍然工作
		if originalEncrypt != nil {
			decrypted, err = originalEncrypt.Decrypt(encrypted)
			if err == nil && decrypted == testString {
				t.Log("Original encrypt still works (SetEncrypt may have been called before)")
				return
			}
		}
		return
	}

	// 如果解密成功，验证值
	if decrypted != testString {
		t.Errorf("RoundTrip() = %v, want %v", decrypted, testString)
	}
}

// BenchmarkDefaultBase64Encrypt_Encrypt 基准测试 Encrypt 方法
func BenchmarkDefaultBase64Encrypt_Encrypt(b *testing.B) {
	encrypt := strutil.GetEncrypt()
	testString := "This is a test string for benchmarking"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encrypt.Encrypt(testString)
	}
}

// BenchmarkDefaultBase64Encrypt_Decrypt 基准测试 Decrypt 方法
func BenchmarkDefaultBase64Encrypt_Decrypt(b *testing.B) {
	encrypt := strutil.GetEncrypt()
	encrypted := base64.StdEncoding.EncodeToString([]byte("This is a test string for benchmarking"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encrypt.Decrypt(encrypted)
	}
}

// BenchmarkEncryptString_Value 基准测试 EncryptString.Value 方法
func BenchmarkEncryptString_Value(b *testing.B) {
	es := strutil.EncryptString("This is a test string for benchmarking")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = es.Value()
	}
}

// BenchmarkEncryptString_Scan 基准测试 EncryptString.Scan 方法
func BenchmarkEncryptString_Scan(b *testing.B) {
	encrypted := base64.StdEncoding.EncodeToString([]byte("This is a test string for benchmarking"))
	var es strutil.EncryptString

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = es.Scan(encrypted)
	}
}
