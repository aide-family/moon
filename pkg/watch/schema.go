package watch

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewDefaultSchemer() Schemer {
	return &defaultSchemer{}
}

type (
	Schemer interface {
		// Decode 解码
		Decode(in *Message, out any) error
		// Encode 编码
		Encode(in *Message, out any) error
	}

	defaultSchemer struct{}
)

func (d *defaultSchemer) Decode(in *Message, out any) error {
	switch in.GetTopic() {
	default:
		return status.Errorf(codes.Unimplemented, "decode unimplemented topic: %s", in.GetTopic())
	}
}

func (d *defaultSchemer) Encode(in *Message, out any) error {
	switch in.GetTopic() {
	default:
		return status.Errorf(codes.Unimplemented, "encode unimplemented topic: %s", in.GetTopic())
	}
}
