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

// Marshal 编码
func Marshal(v any) ([]byte, error) {
	return defaultMarshalFn(v)
}

// Unmarshal 解码
func Unmarshal(data []byte, v any) error {
	return defaultUnmarshalFn(data, v)
}

// SetDefaultMarshalFn 设置默认编码函数
func SetDefaultMarshalFn(fn func(v any) ([]byte, error)) {
	setMarshalOnce.Do(func() {
		defaultMarshalFn = fn
	})
}

// SetDefaultUnmarshalFn 设置默认解码函数
func SetDefaultUnmarshalFn(fn func(data []byte, v any) error) {
	setUnmarshalOnce.Do(func() {
		defaultUnmarshalFn = fn
	})
}

type (
	// Decoder 解码器
	Decoder struct {
		unmarshalFn func(data []byte, v any) error
		r           io.Reader
	}

	// DecoderOption 解码器选项
	DecoderOption func(*Decoder)
)

// NewDecoder 创建一个解码器
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

// Decode 解码
func (d *Decoder) Decode(v any) error {
	readAll, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}
	return d.unmarshalFn(readAll, v)
}

// RegisterCodec 注册编码器
func RegisterCodec(codec encoding.Codec) {
	encoding.RegisterCodec(codec)
	SetDefaultMarshalFn(codec.Marshal)
	SetDefaultUnmarshalFn(codec.Unmarshal)
}

type (
	// Encoder 编码器
	Encoder struct {
		marshalFn func(v any) ([]byte, error)
		w         io.Writer
	}
)

// NewEncoder 创建一个编码器
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		marshalFn: defaultMarshalFn,
		w:         w,
	}
}

// Encode 编码
func (e *Encoder) Encode(v any) error {
	data, err := e.marshalFn(v)
	if err != nil {
		return err
	}
	_, err = e.w.Write(data)
	return err
}
