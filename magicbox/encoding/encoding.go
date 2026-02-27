// Package encoding provides a codec interface.
package encoding

import (
	"github.com/aide-family/magicbox/safety"
	"github.com/go-kratos/kratos/v2/encoding"
)

type Codec interface {
	// Valid checks if the data is valid.
	Valid(data []byte) bool
	encoding.Codec
}

var registeredCodecs = safety.NewSyncMap(make(map[string]Codec))

// RegisterCodec registers a new codec.
func RegisterCodec(name string, codec Codec) {
	registeredCodecs.Set(name, codec)
}

// GetCodec gets a codec.
// If the codec is not found, it will return false.
// If the codec is found, it will return true and the codec.
func GetCodec(name string) (Codec, bool) {
	return registeredCodecs.Get(name)
}

// Codecs returns all registered codecs.
func Codecs() []string {
	return registeredCodecs.Keys()
}
