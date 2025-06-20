package hash_test

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/aide-family/moon/pkg/util/hash"
)

func TestMD5(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "hello world",
			input:    "hello world",
			expected: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name:     "chinese characters",
			input:    "你好世界",
			expected: "65396ee4aad0b4f17aacd1c6112ee364",
		},
		{
			name:     "special characters",
			input:    "!@#$%^&*()",
			expected: "05b28d17a7b6e7024b6e5d8cc43a8bf7",
		},
		{
			name:     "numbers",
			input:    "1234567890",
			expected: "e807f1fcf82d132f9bb018ca6738a19f",
		},
		{
			name:     "long string",
			input:    strings.Repeat("a", 1000),
			expected: "cabe45dcc9ae5b66ba86600cca6b8ba8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hash.MD5(tt.input)
			if result != tt.expected {
				t.Errorf("MD5(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMD5Bytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty bytes",
			input:    []byte{},
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "hello world bytes",
			input:    []byte("hello world"),
			expected: "5eb63bbbe01eeed093cb22bb8f5acdc3",
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05},
			expected: "d15ae53931880fd7b724dd7888b4b4ed",
		},
		{
			name:     "unicode bytes",
			input:    []byte("你好世界"),
			expected: "65396ee4aad0b4f17aacd1c6112ee364",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hash.MD5Bytes(tt.input)
			if result != tt.expected {
				t.Errorf("MD5Bytes(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMD5File(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "md5_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Test data
	testData := "hello world"
	expectedHash := "5eb63bbbe01eeed093cb22bb8f5acdc3"

	// Write test data to file
	_, err = tempFile.WriteString(testData)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test successful file hash
	t.Run("valid file", func(t *testing.T) {
		result, err := hash.MD5File(tempFile.Name())
		if err != nil {
			t.Errorf("MD5File() error = %v", err)
			return
		}
		if result != expectedHash {
			t.Errorf("MD5File() = %v, want %v", result, expectedHash)
		}
	})

	// Test non-existent file
	t.Run("non-existent file", func(t *testing.T) {
		_, err := hash.MD5File("non_existent_file.txt")
		if err == nil {
			t.Error("MD5File() should return error for non-existent file")
		}
	})

	// Test empty file
	t.Run("empty file", func(t *testing.T) {
		emptyFile, err := os.CreateTemp("", "md5_empty_test_*.txt")
		if err != nil {
			t.Fatalf("Failed to create empty temp file: %v", err)
		}
		defer os.Remove(emptyFile.Name())
		emptyFile.Close()

		result, err := hash.MD5File(emptyFile.Name())
		if err != nil {
			t.Errorf("MD5File() error = %v", err)
			return
		}
		expectedEmptyHash := "d41d8cd98f00b204e9800998ecf8427e"
		if result != expectedEmptyHash {
			t.Errorf("MD5File() for empty file = %v, want %v", result, expectedEmptyHash)
		}
	})
}

func TestMD5Consistency(t *testing.T) {
	// Test that MD5 and MD5Bytes produce the same result for the same input
	testString := "test consistency"

	md5Result := hash.MD5(testString)
	md5BytesResult := hash.MD5Bytes([]byte(testString))

	if md5Result != md5BytesResult {
		t.Errorf("MD5(%q) = %v, MD5Bytes(%v) = %v, should be equal",
			testString, md5Result, []byte(testString), md5BytesResult)
	}
}

func TestMD5AgainstStandardLibrary(t *testing.T) {
	// Test that our MD5 function produces the same result as the standard library
	testString := "test against standard library"

	// Our implementation
	ourResult := hash.MD5(testString)

	// Standard library implementation
	h := md5.New()
	h.Write([]byte(testString))
	standardResult := hex.EncodeToString(h.Sum(nil))

	if ourResult != standardResult {
		t.Errorf("Our MD5(%q) = %v, standard library = %v, should be equal",
			testString, ourResult, standardResult)
	}
}

func BenchmarkMD5(b *testing.B) {
	testString := "benchmark test string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.MD5(testString)
	}
}

func BenchmarkMD5Bytes(b *testing.B) {
	testData := []byte("benchmark test bytes")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.MD5Bytes(testData)
	}
}

func BenchmarkMD5File(b *testing.B) {
	// Create a temporary file for benchmarking
	tempFile, err := os.CreateTemp("", "md5_benchmark_*.txt")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write some test data
	testData := strings.Repeat("benchmark data ", 100)
	_, err = tempFile.WriteString(testData)
	if err != nil {
		b.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash.MD5File(tempFile.Name())
	}
}
