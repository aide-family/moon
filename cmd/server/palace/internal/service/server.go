package service

import (
	"context"
	pb "github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
)

type ServerService struct {
	pb.UnimplementedServerServer

	serverRegisterBiz *biz.ServerRegisterBiz
}

// NewServerService create new server service
func NewServerService(serverRegisterBiz *biz.ServerRegisterBiz) *ServerService {
	return &ServerService{
		serverRegisterBiz: serverRegisterBiz,
	}
}

func (s *ServerService) GetServerInfo(ctx context.Context, req *pb.GetServerInfoRequest) (*pb.GetServerInfoReply, error) {
	return &pb.GetServerInfoReply{}, nil
}

func (s *ServerService) GetServerList(ctx context.Context, req *pb.GetServerListRequest) (*pb.GetServerListReply, error) {
	list, err := s.serverRegisterBiz.GetServerList(ctx, req)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ServerService) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	if err := s.serverRegisterBiz.Heartbeat(ctx, req); err != nil {
		return nil, err
	}
	return &pb.HeartbeatReply{}, nil
}
