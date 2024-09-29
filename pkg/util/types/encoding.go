package types

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/go-kratos/kratos/v2/encoding"
)

var defaultMarshalFn = json.Marshal
var defaultUnmarshalFn = json.Unmarshal
var setMarshalOnce sync.Once
var setUnmarshalOnce sync.Once

// Marshal marshals the given value into a byte slice.
func Marshal(v any) ([]byte, error) {
	return defaultMarshalFn(v)
}

// Unmarshal unmarshals the given byte slice into the given value.
func Unmarshal(data []byte, v any) error {
	return defaultUnmarshalFn(data, v)
}

// SetDefaultMarshalFn sets the default marshal function.
func SetDefaultMarshalFn(fn func(v any) ([]byte, error)) {
	setMarshalOnce.Do(func() {
		defaultMarshalFn = fn
	})
}

// SetDefaultUnmarshalFn sets the default unmarshal function.
func SetDefaultUnmarshalFn(fn func(data []byte, v any) error) {
	setUnmarshalOnce.Do(func() {
		defaultUnmarshalFn = fn
	})
}

type (

	// Decoder is a generic decoder that can be used to decode values.
	Decoder struct {
		unmarshalFn func(data []byte, v any) error
		r           io.Reader
	}

	// DecoderOption is a function that can be used to configure a decoder.
	DecoderOption func(*Decoder)
)

// NewDecoder creates a new decoder with the given options.
func NewDecoder(r io.Reader, opts ...DecoderOption) *Decoder {
	d := &Decoder{
		unmarshalFn: defaultUnmarshalFn,
		r:           r,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func (d *Decoder) Decode(v any) error {
	readAll, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}
	return d.unmarshalFn(readAll, v)
}

// RegisterCodec registers a codec with the default marshal and unmarshal functions.
func RegisterCodec(codec encoding.Codec) {
	encoding.RegisterCodec(codec)
	SetDefaultMarshalFn(codec.Marshal)
	SetDefaultUnmarshalFn(codec.Unmarshal)
}
