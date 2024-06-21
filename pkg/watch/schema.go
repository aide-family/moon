package watch

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewDefaultSchemer() Schemer {
	return &defaultSchemer{}
}

func NewEmptySchemer() Schemer {
	return &emptySchemer{}
}

type (
	Schemer interface {
		// Decode 解码
		Decode(in *Message, out any) error

		// Encode 编码
		Encode(in *Message, out any) error
	}

	defaultSchemer struct{}

	emptySchemer struct{}
)

func (d *defaultSchemer) Decode(in *Message, out any) error {
	switch in.GetTopic() {
	// TODO 待实现
	default:
		return status.Errorf(codes.Unimplemented, "decode unimplemented topic: %s", in.GetTopic())
	}
}

func (d *defaultSchemer) Encode(in *Message, out any) error {
	switch in.GetTopic() {
	// TODO 待实现
	default:
		return status.Errorf(codes.Unimplemented, "encode unimplemented topic: %s", in.GetTopic())
	}
}

func (e *emptySchemer) Decode(_ *Message, _ any) error {
	return nil
}

func (e *emptySchemer) Encode(_ *Message, _ any) error {
	return nil
}
