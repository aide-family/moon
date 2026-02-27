package yaml_test

import (
	"testing"

	"github.com/aide-family/magicbox/encoding"
	"github.com/aide-family/magicbox/encoding/yaml"
)

func TestYamlCodec_Valid(t *testing.T) {
	// trigger init() to register the codec
	codec, ok := encoding.GetCodec(yaml.Name)
	if !ok {
		t.Fatal("yaml codec not registered")
	}

	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid simple key-value",
			data: []byte("key: value"),
			want: true,
		},
		{
			name: "valid nested yaml",
			data: []byte("parent:\n  child: value\n  list:\n    - item1\n    - item2"),
			want: true,
		},
		{
			name: "valid empty document",
			data: []byte(""),
			want: true,
		},
		{
			name: "valid yaml list",
			data: []byte("- one\n- two\n- three"),
			want: true,
		},
		{
			name: "valid scalar",
			data: []byte("hello"),
			want: true,
		},
		{
			name: "invalid yaml - bad indentation",
			data: []byte("parent:\n  child: value\n bad: indent"),
			want: false,
		},
		{
			name: "invalid yaml - tab indentation",
			data: []byte("parent:\n\tchild: value"),
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

func TestYamlCodec_MarshalUnmarshal(t *testing.T) {
	codec, ok := encoding.GetCodec(yaml.Name)
	if !ok {
		t.Fatal("yaml codec not registered")
	}

	type sample struct {
		Name  string   `yaml:"name" json:"name"`
		Age   int      `yaml:"age" json:"age"`
		Items []string `yaml:"items" json:"items"`
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

func TestYamlCodec_Name(t *testing.T) {
	codec, ok := encoding.GetCodec(yaml.Name)
	if !ok {
		t.Fatal("yaml codec not registered")
	}

	if got := codec.Name(); got != yaml.Name {
		t.Errorf("Name() = %v, want %v", got, yaml.Name)
	}
}
