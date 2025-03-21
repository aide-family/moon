package format

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatter(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     interface{}
		expected string
	}{
		{
			name:     "empty format string",
			format:   "",
			data:     nil,
			expected: "",
		},
		{
			name:     "simple template",
			format:   "Hello {{.name}}!",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World!",
		},
		{
			name:   "template with functions",
			format: "{{toUpper .text}} {{toLower .name}}",
			data: map[string]interface{}{
				"text": "hello",
				"name": "WORLD",
			},
			expected: "HELLO world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Formatter(tt.format, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatterWithErr(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		data        interface{}
		expected    string
		expectError bool
	}{
		{
			name:        "empty format string",
			format:      "",
			data:        nil,
			expected:    "",
			expectError: true,
		},
		{
			name:        "nil data",
			format:      "test",
			data:        nil,
			expected:    "",
			expectError: true,
		},
		{
			name:        "valid template",
			format:      "Hello {{.name}}!",
			data:        map[string]interface{}{"name": "World"},
			expected:    "Hello World!",
			expectError: false,
		},
		{
			name:        "invalid template",
			format:      "Hello {{.name!",
			data:        map[string]interface{}{"name": "World"},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatterWithErr(tt.format, tt.data)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetObjectByPath(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    string
		expected interface{}
	}{
		{
			name:     "empty json",
			key:      "test",
			value:    "",
			expected: nil,
		},
		{
			name:     "simple json path",
			key:      "name",
			value:    `{"name": "test"}`,
			expected: "test",
		},
		{
			name:     "nested json path",
			key:      "user.name",
			value:    `{"user": {"name": "test"}}`,
			expected: "test",
		},
		{
			name:     "array json path",
			key:      "users.0.name",
			value:    `{"users": [{"name": "test"}]}`,
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetObjectByPath(tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetObjectByKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    map[string]interface{}
		expected interface{}
	}{
		{
			name:     "nil map",
			key:      "test",
			value:    nil,
			expected: nil,
		},
		{
			name:     "existing key",
			key:      "name",
			value:    map[string]interface{}{"name": "test"},
			expected: "test",
		},
		{
			name:     "non-existing key",
			key:      "age",
			value:    map[string]interface{}{"name": "test"},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetObjectByKey(tt.key, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatterWithTemplateFuncs(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     interface{}
		expected string
	}{
		{
			name:   "test getObjectByKey",
			format: `{{ getObjectByKey "name" .user }}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "alice",
					"age":  20,
				},
			},
			expected: "alice",
		},
		{
			name:   "test getObjectByKey with nested map",
			format: `{{ getObjectByKey "address" .user }}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name":    "bob",
					"address": "Beijing",
				},
			},
			expected: "Beijing",
		},
		{
			name:   "test getObjectByKey with non-existing key",
			format: `{{ getObjectByKey "phone" .user }}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "charlie",
				},
			},
			expected: "<nil>",
		},
		{
			name:   "test getObjectByPath with simple path",
			format: `{{ getObjectByPath "name" .jsonData }}`,
			data: map[string]interface{}{
				"jsonData": `{"name": "david", "age": 25}`,
			},
			expected: "david",
		},
		{
			name:   "test getObjectByPath with nested path",
			format: `{{ getObjectByPath "user.profile.name" .jsonData }}`,
			data: map[string]interface{}{
				"jsonData": `{
					"user": {
						"profile": {
							"name": "eve",
							"age": 30
						}
					}
				}`,
			},
			expected: "eve",
		},
		{
			name:   "test getObjectByPath with array",
			format: `{{ getObjectByPath "users.0.name" .jsonData }}`,
			data: map[string]interface{}{
				"jsonData": `{
					"users": [
						{"name": "frank", "age": 35},
						{"name": "grace", "age": 28}
					]
				}`,
			},
			expected: "frank",
		},
		{
			name:   "combine multiple template functions",
			format: `{{ toUpper (getObjectByKey "name" .user) }} - {{ getObjectByPath "location.city" .jsonData }}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "alice",
				},
				"jsonData": `{
					"location": {
						"city": "Shanghai",
						"country": "China"
					}
				}`,
			},
			expected: "ALICE - Shanghai",
		},
		{
			name:   "test with other template functions",
			format: `{{ toLower (getObjectByKey "name" .user) }} {{ toUpper (getObjectByPath "city" .jsonData) }}`,
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "ALICE",
				},
				"jsonData": `{"city": "beijing"}`,
			},
			expected: "alice BEIJING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Formatter(tt.format, tt.data)
			t.Log(result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatterWithPredefinedFuncs(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		data     interface{}
		expected string
	}{
		{
			name:     "test now function",
			format:   `Year: {{ (now).Year }}`,
			data:     nil,
			expected: fmt.Sprintf("Year: %d", time.Now().Year()),
		},
		{
			name:   "test string functions - hasPrefix",
			format: `{{ hasPrefix .text "hello" }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: "true",
		},
		{
			name:   "test string functions - hasSuffix",
			format: `{{ hasSuffix .text "world" }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: "true",
		},
		{
			name:   "test string functions - contains",
			format: `{{ contains .text "llo" }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: "true",
		},
		{
			name:   "test string functions - trimSpace",
			format: `"{{ trimSpace .text }}"`,
			data: map[string]interface{}{
				"text": "  hello world  ",
			},
			expected: `"hello world"`,
		},
		{
			name:   "test string functions - trimPrefix",
			format: `{{ trimPrefix .text "hello" }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: " world",
		},
		{
			name:   "test string functions - trimSuffix",
			format: `{{ trimSuffix .text "world" }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: "hello ",
		},
		{
			name:   "test string functions - toUpper",
			format: `{{ toUpper .text }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: "HELLO WORLD",
		},
		{
			name:   "test string functions - toLower",
			format: `{{ toLower .text }}`,
			data: map[string]interface{}{
				"text": "HELLO WORLD",
			},
			expected: "hello world",
		},
		{
			name:   "test string functions - replace",
			format: `{{ replace .text "hello" "hi" 1 }}`,
			data: map[string]interface{}{
				"text": "hello hello world",
			},
			expected: "hi hello world",
		},
		{
			name:   "test string functions - split",
			format: `{{ $parts := split .text " " }}{{ index $parts 0 }},{{ index $parts 1 }}`,
			data: map[string]interface{}{
				"text": "hello world",
			},
			expected: "hello,world",
		},
		{
			name:   "test multiple functions in pipeline",
			format: `{{ .text | toLower | trimSpace | toUpper }}`,
			data: map[string]interface{}{
				"text": "  Hello World  ",
			},
			expected: "HELLO WORLD",
		},
		{
			name:   "test functions with complex data",
			format: `{{ range $idx, $item := split .text "," }}{{ if ne $idx 0 }} + {{ end }}{{ toUpper (trimSpace $item) }}{{ end }}`,
			data: map[string]interface{}{
				"text": "go, python, java",
			},
			expected: "GO + PYTHON + JAVA",
		},
		{
			name:   "test getObjectByPath with json array",
			format: `{{ range $idx, $item := split (getObjectByPath "skills" .jsonData) "," }}{{ if ne $idx 0 }}, {{ end }}{{ toUpper (trimSpace $item) }}{{ end }}`,
			data: map[string]interface{}{
				"jsonData": `{"name": "Alice", "skills": "go, python, java"}`,
			},
			expected: "GO, PYTHON, JAVA",
		},
		{
			name: "test all string functions together",
			format: `{{ $text := .text | trimSpace }}
HasPrefix 'hello': {{ hasPrefix $text "hello" }}
HasSuffix 'world': {{ hasSuffix $text "world" }}
Contains 'lo wo': {{ contains $text "lo wo" }}
ToUpper: {{ toUpper $text }}
ToLower: {{ toLower $text }}
TrimPrefix 'hello': {{ trimPrefix $text "hello" }}
TrimSuffix 'world': {{ trimSuffix $text "world" }}
Replace 'o' with '0': {{ replace $text "o" "0" -1 }}`,
			data: map[string]interface{}{
				"text": "  hello world  ",
			},
			expected: `
HasPrefix 'hello': true
HasSuffix 'world': true
Contains 'lo wo': true
ToUpper: HELLO WORLD
ToLower: hello world
TrimPrefix 'hello':  world
TrimSuffix 'world': hello 
Replace 'o' with '0': hell0 w0rld`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Formatter(tt.format, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}
