package service

import (
	"context"

	pb "github.com/aide-cloud/moon/api/rabbit/push"
)

type ConfigService struct {
	pb.UnimplementedConfigServer
}

func NewConfigService() *ConfigService {
	return &ConfigService{}
}

func (s *ConfigService) NotifyObject(ctx context.Context, req *pb.NotifyObjectRequest) (*pb.NotifyObjectReply, error) {
	return &pb.NotifyObjectReply{}, nil
}
