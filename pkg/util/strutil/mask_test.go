package strutil

import "testing"

func TestMaskString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		end      int
		expected string
	}{
		{
			name:     "normal string",
			input:    "1234567890",
			start:    3,
			end:      2,
			expected: "123******90",
		},
		{
			name:     "empty string",
			input:    "",
			start:    3,
			end:      2,
			expected: "",
		},
		{
			name:     "short string",
			input:    "123",
			start:    2,
			end:      2,
			expected: "123******",
		},
		{
			name:     "negative start",
			input:    "1234567890",
			start:    -1,
			end:      2,
			expected: "******90",
		},
		{
			name:     "negative end",
			input:    "1234567890",
			start:    3,
			end:      -1,
			expected: "123******",
		},
		{
			name:     "UTF-8 string",
			input:    "你好世界",
			start:    1,
			end:      1,
			expected: "你******界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskString(tt.input, tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("MaskString(%q, %d, %d) = %q, want %q",
					tt.input, tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal email",
			input:    "user@example.com",
			expected: "u******r@example.com",
		},
		{
			name:     "short local part",
			input:    "u@example.com",
			expected: "u******@example.com",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no @ symbol",
			input:    "userexample.com",
			expected: "userexample.com******",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskEmail(tt.input)
			if result != tt.expected {
				t.Errorf("MaskEmail(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal phone",
			input:    "13812345678",
			expected: "138******5678",
		},
		{
			name:     "short number",
			input:    "12345",
			expected: "12345******",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskPhone(tt.input)
			if result != tt.expected {
				t.Errorf("MaskPhone(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskBankCard(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal card number",
			input:    "6225123412341234",
			expected: "6225******1234",
		},
		{
			name:     "short number",
			input:    "12345678",
			expected: "12345678******",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskBankCard(tt.input)
			if result != tt.expected {
				t.Errorf("MaskBankCard(%q) = %q, want %q",
					tt.input, result, tt.expected)
			}
		})
	}
}
