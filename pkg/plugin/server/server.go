package server

import (
	"github.com/go-kratos/kratos/v2/encoding/json"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	json.MarshalOptions = protojson.MarshalOptions{
		UseEnumNumbers:  true, // Emit enum values as numbers instead of their string representation (default is string).
		UseProtoNames:   true, // Use the field names defined in the proto file as the output field names.
		EmitUnpopulated: true, // Emit fields even if they are unset or empty.
	}
}

type Servers []transport.Server

func (s Servers) Append(servers ...transport.Server) Servers {
	return append(s, servers...)
}
