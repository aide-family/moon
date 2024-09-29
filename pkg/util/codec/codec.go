package codec

import (
	"strings"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/encoding"
)

// RegisterCodec register codec
func RegisterCodec(ext string) {
	switch strings.ToLower(ext) {
	case tomlName:
		encoding.RegisterCodec(TomlCodec{})
		types.RegisterCodec(TomlCodec{})
	default:
		encoding.RegisterCodec(YamlCodec{})
		types.RegisterCodec(YamlCodec{})
	}
}
