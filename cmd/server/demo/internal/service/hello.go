package service

import (
	"context"

	pb "github.com/aide-family/moon/api/v1"
	"github.com/aide-family/moon/cmd/server/demo/internal/biz"
)

type HelloService struct {
	pb.UnimplementedHelloServer

	helloBiz *biz.HelloBiz
}

func NewHelloService(helloBiz *biz.HelloBiz) *HelloService {
	return &HelloService{
		helloBiz: helloBiz,
	}
}

func (s *HelloService) SayHello(ctx context.Context, req *pb.SayHelloRequest) (*pb.SayHelloReply, error) {
	hello, err := s.helloBiz.SayHello(ctx, req.GetName())
	if err != nil {
		return nil, err
	}
	return &pb.SayHelloReply{
		Message: hello,
	}, nil
}
