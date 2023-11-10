package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/prom/strategy"
	"prometheus-manager/app/prom_server/internal/biz/prombiz"
)

type StrategyService struct {
	pb.UnimplementedStrategyServer

	log *log.Helper

	strategyBiz *prombiz.StrategyBiz
}

func NewStrategyService(strategyBiz *prombiz.StrategyBiz, logger log.Logger) *StrategyService {
	return &StrategyService{
		log:         log.NewHelper(log.With(logger, "module", "service.prom.strategy")),
		strategyBiz: strategyBiz,
	}
}

func (s *StrategyService) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	return &pb.CreateStrategyReply{}, nil
}

func (s *StrategyService) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	return &pb.UpdateStrategyReply{}, nil
}

func (s *StrategyService) BatchUpdateStrategyStatus(ctx context.Context, req *pb.BatchUpdateStrategyStatusRequest) (*pb.BatchUpdateStrategyStatusReply, error) {
	return &pb.BatchUpdateStrategyStatusReply{}, nil
}

func (s *StrategyService) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	return &pb.DeleteStrategyReply{}, nil
}

func (s *StrategyService) BatchDeleteStrategy(ctx context.Context, req *pb.BatchDeleteStrategyRequest) (*pb.BatchDeleteStrategyReply, error) {
	return &pb.BatchDeleteStrategyReply{}, nil
}

func (s *StrategyService) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	return &pb.GetStrategyReply{}, nil
}

func (s *StrategyService) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	return &pb.ListStrategyReply{}, nil
}

func (s *StrategyService) SelectStrategy(ctx context.Context, req *pb.SelectStrategyRequest) (*pb.SelectStrategyReply, error) {
	return &pb.SelectStrategyReply{}, nil
}

func (s *StrategyService) ExportStrategy(ctx context.Context, req *pb.ExportStrategyRequest) (*pb.ExportStrategyReply, error) {
	return &pb.ExportStrategyReply{}, nil
}
