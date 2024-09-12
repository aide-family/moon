package codec

import (
	"strings"

	"github.com/go-kratos/kratos/v2/encoding"
)

// RegisterCodec register codec
func RegisterCodec(ext string) {
	switch strings.ToLower(ext) {
	case tomlName:
		encoding.RegisterCodec(TomlCodec{})
	default:
		encoding.RegisterCodec(YamlCodec{})
	}
}
