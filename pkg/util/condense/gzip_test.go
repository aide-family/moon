package condense_test

import (
	"bytes"
	"encoding/json"
	"io"
	"reflect"
	"testing"

	"github.com/aide-family/moon/pkg/util/condense"
)

func TestGzip(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "normal data",
			data:    []byte("Hello, World!"),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "large data",
			data:    bytes.Repeat([]byte("test"), 1000),
			wantErr: false,
		},
		{
			name:    "unicode data",
			data:    []byte("Hello, 世界!"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := condense.Gzip(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Gzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			compressed, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("Failed to read compressed data: %v", err)
				return
			}

			// Verify it's actually compressed (should be smaller for non-empty data)
			if len(tt.data) > 0 && len(compressed) >= len(tt.data) {
				t.Logf("Original size: %d, Compressed size: %d", len(tt.data), len(compressed))
			}

			// Test decompression
			decompressed, err := condense.UnGzip(bytes.NewReader(compressed))
			if err != nil {
				t.Errorf("Failed to decompress: %v", err)
				return
			}

			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("Decompressed data doesn't match original. Got %q, want %q", decompressed, tt.data)
			}
		})
	}
}

func TestGzipBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "normal data",
			data:    []byte("Hello, World!"),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "large data",
			data:    bytes.Repeat([]byte("test"), 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := condense.GzipBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GzipBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			// Test decompression
			decompressed, err := condense.UnGzipBytes(compressed)
			if err != nil {
				t.Errorf("Failed to decompress: %v", err)
				return
			}

			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("Decompressed data doesn't match original. Got %q, want %q", decompressed, tt.data)
			}
		})
	}
}

func TestGzipJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "simple struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name: "map data",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
				"key3": []string{"a", "b", "c"},
			},
			wantErr: false,
		},
		{
			name:    "nil data",
			data:    nil,
			wantErr: false,
		},
		{
			name: "complex nested struct",
			data: struct {
				User struct {
					Name string `json:"name"`
					Info struct {
						Email string `json:"email"`
						Phone string `json:"phone"`
					} `json:"info"`
				} `json:"user"`
				Tags []string `json:"tags"`
			}{
				User: struct {
					Name string `json:"name"`
					Info struct {
						Email string `json:"email"`
						Phone string `json:"phone"`
					} `json:"info"`
				}{
					Name: "Alice",
					Info: struct {
						Email string `json:"email"`
						Phone string `json:"phone"`
					}{
						Email: "alice@example.com",
						Phone: "123-456-7890",
					},
				},
				Tags: []string{"admin", "user"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader, err := condense.GzipJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GzipJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			compressed, err := io.ReadAll(reader)
			if err != nil {
				t.Errorf("Failed to read compressed data: %v", err)
				return
			}

			if tt.data == nil {
				return // nil数据不做比较
			}

			// 对于map类型，使用JSON字符串比较
			if reflect.TypeOf(tt.data).Kind() == reflect.Map {
				var result interface{}
				err = condense.UnGzipJSONUnmarshal(bytes.NewReader(compressed), &result)
				if err != nil {
					t.Errorf("Failed to decompress and unmarshal JSON: %v", err)
					return
				}

				// 转换为JSON字符串进行比较
				originalJSON, _ := json.Marshal(tt.data)
				resultJSON, _ := json.Marshal(result)
				if !bytes.Equal(originalJSON, resultJSON) {
					t.Errorf("JSON data doesn't match. Original: %s, Result: %s", originalJSON, resultJSON)
				}
			} else {
				// 对于结构体类型，使用reflect
				got := reflect.New(reflect.TypeOf(tt.data)).Interface()
				err = condense.UnGzipJSONUnmarshal(bytes.NewReader(compressed), got)
				if err != nil {
					t.Errorf("Failed to decompress and unmarshal JSON: %v", err)
					return
				}

				if !reflect.DeepEqual(tt.data, reflect.ValueOf(got).Elem().Interface()) {
					t.Errorf("JSON data doesn't match. Original: %+v, Result: %+v", tt.data, reflect.ValueOf(got).Elem().Interface())
				}
			}
		})
	}
}

func TestGzipJSONBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "simple struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name:    "slice data",
			data:    []string{"apple", "banana", "cherry"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := condense.GzipJSONBytes(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GzipJSONBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			got := reflect.New(reflect.TypeOf(tt.data)).Interface()
			err = condense.UnGzipJSONUnmarshalBytes(compressed, got)
			if err != nil {
				t.Errorf("Failed to decompress and unmarshal JSON: %v", err)
				return
			}

			if !reflect.DeepEqual(tt.data, reflect.ValueOf(got).Elem().Interface()) {
				t.Errorf("JSON data doesn't match. Original: %+v, Result: %+v", tt.data, reflect.ValueOf(got).Elem().Interface())
			}
		})
	}
}

func TestUnGzip(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "normal data",
			data:    []byte("Hello, World!"),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First compress the data
			compressed, err := condense.GzipBytes(tt.data)
			if err != nil {
				t.Errorf("Failed to compress data for test: %v", err)
				return
			}

			// Then decompress it
			decompressed, err := condense.UnGzip(bytes.NewReader(compressed))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnGzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("UnGzip() = %q, want %q", decompressed, tt.data)
			}
		})
	}
}

func TestUnGzip_InvalidData(t *testing.T) {
	// Test with invalid gzip data
	invalidData := []byte("This is not gzip data")
	_, err := condense.UnGzip(bytes.NewReader(invalidData))
	if err == nil {
		t.Error("UnGzip() should return error for invalid gzip data")
	}
}

func TestUnGzipBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "normal data",
			data:    []byte("Hello, World!"),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: false,
		},
		{
			name:    "large data",
			data:    bytes.Repeat([]byte("test"), 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First compress the data
			compressed, err := condense.GzipBytes(tt.data)
			if err != nil {
				t.Errorf("Failed to compress data for test: %v", err)
				return
			}

			// Then decompress it
			decompressed, err := condense.UnGzipBytes(compressed)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnGzipBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if !bytes.Equal(decompressed, tt.data) {
				t.Errorf("UnGzipBytes() = %q, want %q", decompressed, tt.data)
			}
		})
	}
}

func TestUnGzipBytes_InvalidData(t *testing.T) {
	// Test with invalid gzip data
	invalidData := []byte("This is not gzip data")
	_, err := condense.UnGzipBytes(invalidData)
	if err == nil {
		t.Error("UnGzipBytes() should return error for invalid gzip data")
	}
}

func TestUnGzipJSONUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "simple struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name:    "array data",
			data:    []int{1, 2, 3, 4, 5},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := condense.GzipJSONBytes(tt.data)
			if err != nil {
				t.Errorf("Failed to compress JSON data for test: %v", err)
				return
			}

			got := reflect.New(reflect.TypeOf(tt.data)).Interface()
			err = condense.UnGzipJSONUnmarshal(bytes.NewReader(compressed), got)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnGzipJSONUnmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if !reflect.DeepEqual(tt.data, reflect.ValueOf(got).Elem().Interface()) {
				t.Errorf("JSON data doesn't match. Original: %+v, Result: %+v", tt.data, reflect.ValueOf(got).Elem().Interface())
			}
		})
	}
}

func TestUnGzipJSONUnmarshalBytes(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantErr bool
	}{
		{
			name: "simple struct",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			wantErr: false,
		},
		{
			name: "map data",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compressed, err := condense.GzipJSONBytes(tt.data)
			if err != nil {
				t.Errorf("Failed to compress JSON data for test: %v", err)
				return
			}

			// 对于map类型，使用JSON字符串比较
			if reflect.TypeOf(tt.data).Kind() == reflect.Map {
				var result interface{}
				err = condense.UnGzipJSONUnmarshalBytes(compressed, &result)
				if (err != nil) != tt.wantErr {
					t.Errorf("UnGzipJSONUnmarshalBytes() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err != nil {
					return
				}

				// 转换为JSON字符串进行比较
				originalJSON, _ := json.Marshal(tt.data)
				resultJSON, _ := json.Marshal(result)
				if !bytes.Equal(originalJSON, resultJSON) {
					t.Errorf("JSON data doesn't match. Original: %s, Result: %s", originalJSON, resultJSON)
				}
			} else {
				// 对于结构体类型，使用reflect
				got := reflect.New(reflect.TypeOf(tt.data)).Interface()
				err = condense.UnGzipJSONUnmarshalBytes(compressed, got)
				if (err != nil) != tt.wantErr {
					t.Errorf("UnGzipJSONUnmarshalBytes() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if err != nil {
					return
				}

				if !reflect.DeepEqual(tt.data, reflect.ValueOf(got).Elem().Interface()) {
					t.Errorf("JSON data doesn't match. Original: %+v, Result: %+v", tt.data, reflect.ValueOf(got).Elem().Interface())
				}
			}
		})
	}
}

func TestUnGzipJSONUnmarshal_InvalidData(t *testing.T) {
	// Test with invalid gzip data
	invalidData := []byte("This is not gzip data")
	var result interface{}
	err := condense.UnGzipJSONUnmarshal(bytes.NewReader(invalidData), &result)
	if err == nil {
		t.Error("UnGzipJSONUnmarshal() should return error for invalid gzip data")
	}
}

func TestUnGzipJSONUnmarshalBytes_InvalidData(t *testing.T) {
	// Test with invalid gzip data
	invalidData := []byte("This is not gzip data")
	var result interface{}
	err := condense.UnGzipJSONUnmarshalBytes(invalidData, &result)
	if err == nil {
		t.Error("UnGzipJSONUnmarshalBytes() should return error for invalid gzip data")
	}
}

// Benchmark tests
func BenchmarkGzip(b *testing.B) {
	data := bytes.Repeat([]byte("test data for compression"), 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := condense.Gzip(data)
		if err != nil {
			b.Errorf("Gzip() error = %v", err)
		}
	}
}

func BenchmarkGzipBytes(b *testing.B) {
	data := bytes.Repeat([]byte("test data for compression"), 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := condense.GzipBytes(data)
		if err != nil {
			b.Errorf("GzipBytes() error = %v", err)
		}
	}
}

func BenchmarkGzipJSON(b *testing.B) {
	data := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
		"tags":  []string{"user", "admin"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := condense.GzipJSON(data)
		if err != nil {
			b.Errorf("GzipJSON() error = %v", err)
		}
	}
}

func BenchmarkGzipJSONBytes(b *testing.B) {
	data := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
		"tags":  []string{"user", "admin"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := condense.GzipJSONBytes(data)
		if err != nil {
			b.Errorf("GzipJSONBytes() error = %v", err)
		}
	}
}

func BenchmarkUnGzip(b *testing.B) {
	originalData := bytes.Repeat([]byte("test data for compression"), 100)
	compressed, err := condense.GzipBytes(originalData)
	if err != nil {
		b.Fatalf("Failed to prepare test data: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := condense.UnGzip(bytes.NewReader(compressed))
		if err != nil {
			b.Errorf("UnGzip() error = %v", err)
		}
	}
}

func BenchmarkUnGzipBytes(b *testing.B) {
	originalData := bytes.Repeat([]byte("test data for compression"), 100)
	compressed, err := condense.GzipBytes(originalData)
	if err != nil {
		b.Fatalf("Failed to prepare test data: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := condense.UnGzipBytes(compressed)
		if err != nil {
			b.Errorf("UnGzipBytes() error = %v", err)
		}
	}
}
