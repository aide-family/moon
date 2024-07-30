package service

import (
	"context"

	v1api "github.com/aide-family/moon/api/v1"
	"github.com/aide-family/moon/cmd/server/demo/internal/biz"
)

// HelloService is a greeter service.
type HelloService struct {
	v1api.UnimplementedHelloServer

	helloBiz *biz.HelloBiz
}

// NewHelloService new a greeter service.
func NewHelloService(helloBiz *biz.HelloBiz) *HelloService {
	return &HelloService{
		helloBiz: helloBiz,
	}
}

// SayHello implements helloworld.GreeterServer
func (s *HelloService) SayHello(ctx context.Context, req *v1api.SayHelloRequest) (*v1api.SayHelloReply, error) {
	hello, err := s.helloBiz.SayHello(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1api.SayHelloReply{
		Message: hello,
	}, nil
}
