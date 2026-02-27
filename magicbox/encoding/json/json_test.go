package json_test

import (
	"testing"

	"github.com/aide-family/magicbox/encoding"
	"github.com/aide-family/magicbox/encoding/json"
)

func TestJsonCodec_Valid(t *testing.T) {
	codec, ok := encoding.GetCodec(json.Name)
	if !ok {
		t.Fatal("json codec not registered")
	}

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid object",
			data: []byte(`{"key":"value"}`),
			want: true,
		},
		{
			name: "valid nested object",
			data: []byte(`{"parent":{"child":"value","list":[1,2,3]}}`),
			want: true,
		},
		{
			name: "valid array",
			data: []byte(`[1,"two",true,null]`),
			want: true,
		},
		{
			name: "valid string",
			data: []byte(`"hello"`),
			want: true,
		},
		{
			name: "valid number",
			data: []byte(`42`),
			want: true,
		},
		{
			name: "valid boolean",
			data: []byte(`true`),
			want: true,
		},
		{
			name: "valid null",
			data: []byte(`null`),
			want: true,
		},
		{
			name: "empty bytes",
			data: []byte(""),
			want: false,
		},
		{
			name: "invalid json - missing closing brace",
			data: []byte(`{"key":"value"`),
			want: false,
		},
		{
			name: "invalid json - trailing comma",
			data: []byte(`{"key":"value",}`),
			want: false,
		},
		{
			name: "invalid json - plain text",
			data: []byte(`not json`),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := codec.Valid(tt.data); got != tt.want {
				t.Errorf("Valid(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestJsonCodec_MarshalUnmarshal(t *testing.T) {
	codec, ok := encoding.GetCodec(json.Name)
	if !ok {
		t.Fatal("json codec not registered")
	}

	type sample struct {
		Name  string   `json:"name"`
		Age   int      `json:"age"`
		Items []string `json:"items"`
	}

	original := &sample{
		Name:  "test",
		Age:   18,
		Items: []string{"a", "b", "c"},
	}

	data, err := codec.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	if !codec.Valid(data) {
		t.Errorf("Valid() returned false for marshaled data: %s", data)
	}

	var decoded sample
	if err := codec.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Name != original.Name || decoded.Age != original.Age || len(decoded.Items) != len(original.Items) {
		t.Errorf("Unmarshal() got %+v, want %+v", decoded, original)
	}
}

func TestJsonCodec_Name(t *testing.T) {
	codec, ok := encoding.GetCodec(json.Name)
	if !ok {
		t.Fatal("json codec not registered")
	}

	if got := codec.Name(); got != json.Name {
		t.Errorf("Name() = %v, want %v", got, json.Name)
	}
}
