package codec

import (
	"gopkg.in/yaml.v3"
)

const yamlName = "yaml"

// YamlCodec is a Codec implementation with yaml.
//
//	func init() {
//	    encoding.RegisterCodec(codec.YamlCodec{})
//	}
type YamlCodec struct{}

// Marshal returns the wire format of v.
func (YamlCodec) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Unmarshal parses the wire format into v.
func (YamlCodec) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// Name returns the name of the Codec implementation. The returned string
// will be used as part of content type in transmission.  The result must be
// static; the result cannot change between calls.
func (YamlCodec) Name() string {
	return yamlName
}
