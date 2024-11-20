package service

import (
	"context"

	pb "github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
)

// ServerService 服务管理
type ServerService struct {
	pb.UnimplementedServerServer

	serverRegisterBiz *biz.ServerRegisterBiz
}

// NewServerService 创建服务管理
func NewServerService(serverRegisterBiz *biz.ServerRegisterBiz) *ServerService {
	return &ServerService{
		serverRegisterBiz: serverRegisterBiz,
	}
}

// GetServerInfo 获取服务信息
func (s *ServerService) GetServerInfo(ctx context.Context, req *pb.GetServerInfoRequest) (*pb.GetServerInfoReply, error) {
	return &pb.GetServerInfoReply{}, nil
}

// GetServerList 获取服务列表
func (s *ServerService) GetServerList(ctx context.Context, req *pb.GetServerListRequest) (*pb.GetServerListReply, error) {
	list, err := s.serverRegisterBiz.GetServerList(ctx, req)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Heartbeat 心跳
func (s *ServerService) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	if err := s.serverRegisterBiz.Heartbeat(ctx, req); err != nil {
		return nil, err
	}
	return &pb.HeartbeatReply{}, nil
}
