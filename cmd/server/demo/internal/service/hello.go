package service

import (
	"context"

	v1api "github.com/aide-family/moon/api/v1"
	"github.com/aide-family/moon/cmd/server/demo/internal/biz"
)

type HelloService struct {
	v1api.UnimplementedHelloServer

	helloBiz *biz.HelloBiz
}

func NewHelloService(helloBiz *biz.HelloBiz) *HelloService {
	return &HelloService{
		helloBiz: helloBiz,
	}
}

func (s *HelloService) SayHello(ctx context.Context, req *v1api.SayHelloRequest) (*v1api.SayHelloReply, error) {
	hello, err := s.helloBiz.SayHello(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1api.SayHelloReply{
		Message: hello,
	}, nil
}
