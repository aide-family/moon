package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/alarm/hook"
)

type HookService struct {
	pb.UnimplementedHookServer

	log *log.Helper
}

func NewHookService(logger log.Logger) *HookService {
	return &HookService{
		log: log.NewHelper(log.With(logger, "module", "service.hook")),
	}
}

func (s *HookService) V1(ctx context.Context, req *pb.HookV1Request) (*pb.HookV1Reply, error) {
	// 直接转发到prom
	return &pb.HookV1Reply{}, nil
}
