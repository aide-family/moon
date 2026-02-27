package encoding

import (
	"testing"
)

// mockCodec is a minimal Codec implementation for testing.
type mockCodec struct {
	name string
}

func (m *mockCodec) Valid(data []byte) bool                     { return len(data) > 0 }
func (m *mockCodec) Marshal(v interface{}) ([]byte, error)      { return nil, nil }
func (m *mockCodec) Unmarshal(data []byte, v interface{}) error { return nil }
func (m *mockCodec) Name() string                               { return m.name }

func TestRegisterCodec_and_GetCodec(t *testing.T) {
	name := "test-codec"
	codec := &mockCodec{name: name}

	RegisterCodec(name, codec)

	got, ok := GetCodec(name)
	if !ok {
		t.Fatalf("GetCodec(%q) returned ok=false, want true", name)
	}
	if got.Name() != name {
		t.Errorf("GetCodec(%q).Name() = %q, want %q", name, got.Name(), name)
	}
}

func TestGetCodec_not_found(t *testing.T) {
	_, ok := GetCodec("nonexistent-codec")
	if ok {
		t.Error("GetCodec(\"nonexistent-codec\") returned ok=true, want false")
	}
}

func TestRegisterCodec_overwrite(t *testing.T) {
	name := "overwrite-codec"
	codec1 := &mockCodec{name: "v1"}
	codec2 := &mockCodec{name: "v2"}

	RegisterCodec(name, codec1)
	RegisterCodec(name, codec2)

	got, ok := GetCodec(name)
	if !ok {
		t.Fatalf("GetCodec(%q) returned ok=false after overwrite", name)
	}
	if got.Name() != "v2" {
		t.Errorf("GetCodec(%q).Name() = %q, want %q", name, got.Name(), "v2")
	}
}

func TestRegisteredCodec_Valid(t *testing.T) {
	name := "valid-test-codec"
	codec := &mockCodec{name: name}
	RegisterCodec(name, codec)

	got, ok := GetCodec(name)
	if !ok {
		t.Fatalf("GetCodec(%q) returned ok=false", name)
	}

	if !got.Valid([]byte("data")) {
		t.Error("Valid(non-empty) = false, want true")
	}
	if got.Valid([]byte("")) {
		t.Error("Valid(empty) = true, want false")
	}
}
