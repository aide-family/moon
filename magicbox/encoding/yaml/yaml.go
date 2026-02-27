// Package yaml provides a YAML codec.
package yaml

import (
	"buf.build/go/protoyaml"
	kratosencoding "github.com/go-kratos/kratos/v2/encoding"
	kratosyaml "github.com/go-kratos/kratos/v2/encoding/yaml"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"

	"github.com/aide-family/magicbox/encoding"
)

const Name = kratosyaml.Name

func init() {
	encoding.RegisterCodec(Name, &yamlCodec{
		codec: kratosencoding.GetCodec(Name),
		marshalOptions: protoyaml.MarshalOptions{
			UseProtoNames:   true,  // use proto names
			EmitUnpopulated: false, // filter 0 values and empty values
			Indent:          2,     // indent 2 spaces
		},
		unmarshalOptions: protoyaml.UnmarshalOptions{
			DiscardUnknown: true, // discard unknown fields
		},
	})
}

type yamlCodec struct {
	codec            kratosencoding.Codec
	marshalOptions   protoyaml.MarshalOptions
	unmarshalOptions protoyaml.UnmarshalOptions
}

// Marshal implements [encoding.Codec].
func (y *yamlCodec) Marshal(v any) ([]byte, error) {
	switch v := v.(type) {
	case protoreflect.ProtoMessage:
		return y.marshalOptions.Marshal(v)
	default:
		return y.codec.Marshal(v)
	}
}

// Name implements [encoding.Codec].
func (y *yamlCodec) Name() string {
	return Name
}

// Unmarshal implements [encoding.Codec].
func (y *yamlCodec) Unmarshal(data []byte, v any) error {
	switch v := v.(type) {
	case protoreflect.ProtoMessage:
		return y.unmarshalOptions.Unmarshal(data, v)
	default:
		return y.codec.Unmarshal(data, v)
	}
}

// Valid implements [encoding.Codec].
func (y *yamlCodec) Valid(data []byte) bool {
	return yaml.Unmarshal(data, &yaml.Node{}) == nil
}
