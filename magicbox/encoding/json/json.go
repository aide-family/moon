// Package json provides a JSON codec.
package json

import (
	"encoding/json"

	"github.com/aide-family/magicbox/encoding"
	kratosencoding "github.com/go-kratos/kratos/v2/encoding"
	kratosjson "github.com/go-kratos/kratos/v2/encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
)

const Name = kratosjson.Name

func init() {
	kratosjson.MarshalOptions = protojson.MarshalOptions{
		// UseEnumNumbers:  true, // Emit enum values as numbers instead of their string representation (default is string).
		UseProtoNames:   true, // Use the field names defined in the proto file as the output field names.
		EmitUnpopulated: true, // Emit fields even if they are unset or empty.
	}
	kratosjson.UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true, // discard unknown fields
	}
	encoding.RegisterCodec(Name, &jsonCodec{
		Codec: kratosencoding.GetCodec(Name),
	})
}

type jsonCodec struct {
	kratosencoding.Codec
}

// Valid implements [encoding.Codec].
func (j *jsonCodec) Valid(data []byte) bool {
	return json.Valid(data)
}
