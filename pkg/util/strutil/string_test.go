package strutil

import "testing"

func TestTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "multiple words",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "already titled",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "mixed case",
			input:    "hElLo WoRlD",
			expected: "Hello World",
		},
		{
			name:     "with numbers",
			input:    "hello 123 world",
			expected: "Hello 123 World",
		},
		{
			name:     "with special characters",
			input:    "hello-world",
			expected: "Hello-World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Title(tt.input)
			if got != tt.expected {
				t.Errorf("Title() = %v, want %v", got, tt.expected)
			}
		})
	}
}
