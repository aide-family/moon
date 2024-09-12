package codec

import (
	"github.com/BurntSushi/toml"
	"github.com/go-kratos/kratos/v2/encoding"
)

var _ encoding.Codec = (*TomlCodec)(nil)

const tomlName = "toml"

// TomlCodec is a Codec implementation with toml.
//
//	func init() {
//	    encoding.RegisterCodec(codec.TomlCodec{})
//	}
type TomlCodec struct{}

// Marshal returns the wire format of v.
func (TomlCodec) Marshal(v interface{}) ([]byte, error) {
	return toml.Marshal(v)
}

// Unmarshal parses the wire format into v.
func (TomlCodec) Unmarshal(data []byte, v interface{}) error {
	return toml.Unmarshal(data, v)
}

// Name returns the name of the Codec implementation. The returned string
// will be used as part of content type in transmission.  The result must be
// static; the result cannot change between calls.
func (TomlCodec) Name() string {
	return tomlName
}
