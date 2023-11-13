package promservice

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/strategy"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/prombiz"
)

type StrategyService struct {
	pb.UnimplementedStrategyServer

	log *log.Helper

	strategyBiz *prombiz.StrategyXBiz
}

func NewStrategyService(strategyBiz *prombiz.StrategyXBiz, logger log.Logger) *StrategyService {
	return &StrategyService{
		log:         log.NewHelper(log.With(logger, "module", "service.prom.strategy")),
		strategyBiz: strategyBiz,
	}
}

func (s *StrategyService) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	strategyBO, err := s.strategyBiz.CreateStrategy(ctx, &biz.StrategyBO{
		Alert:        req.GetAlert(),
		Expr:         req.GetExpr(),
		Duration:     req.GetDuration(),
		Labels:       req.GetLabels(),
		Annotations:  req.GetAnnotations(),
		Remark:       req.GetRemark(),
		GroupId:      req.GetGroupId(),
		AlarmLevelId: req.GetAlarmLevelId(),
		AlarmPageIds: req.GetAlarmPageIds(),
		CategoryIds:  req.GetCategoryIds(),
	})

	if err != nil {
		return nil, err
	}
	return &pb.CreateStrategyReply{Id: strategyBO.Id}, nil
}

func (s *StrategyService) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	strategyBO, err := s.strategyBiz.UpdateStrategyById(ctx, req.GetId(), &biz.StrategyBO{
		Alert:        req.GetAlert(),
		Expr:         req.GetExpr(),
		Duration:     req.GetDuration(),
		Labels:       req.GetLabels(),
		Annotations:  req.GetAnnotations(),
		Remark:       req.GetRemark(),
		GroupId:      req.GetGroupId(),
		AlarmLevelId: req.GetAlarmLevelId(),
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateStrategyReply{Id: strategyBO.Id}, nil
}

func (s *StrategyService) BatchUpdateStrategyStatus(ctx context.Context, req *pb.BatchUpdateStrategyStatusRequest) (*pb.BatchUpdateStrategyStatusReply, error) {
	if err := s.strategyBiz.BatchUpdateStrategyStatusByIds(ctx, req.GetStatus(), req.GetIds()); err != nil {
		return nil, err
	}

	return &pb.BatchUpdateStrategyStatusReply{Ids: req.GetIds()}, nil
}

func (s *StrategyService) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	if err := s.strategyBiz.DeleteStrategyByIds(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &pb.DeleteStrategyReply{Id: req.GetId()}, nil
}

func (s *StrategyService) BatchDeleteStrategy(ctx context.Context, req *pb.BatchDeleteStrategyRequest) (*pb.BatchDeleteStrategyReply, error) {
	if err := s.strategyBiz.DeleteStrategyByIds(ctx, req.GetIds()...); err != nil {
		return nil, err
	}
	return &pb.BatchDeleteStrategyReply{Ids: req.GetIds()}, nil
}

func (s *StrategyService) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	strategyBO, err := s.strategyBiz.GetStrategyById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	detail := strategyBO.ToApiPromStrategyV1()

	return &pb.GetStrategyReply{Detail: detail}, nil
}

func (s *StrategyService) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	strategyBos, pgInfo, err := s.strategyBiz.ListStrategy(ctx, req)
	if err != nil {
		return nil, err
	}

	list := biz.ListToApiPromStrategyV1(strategyBos...)
	pg := &api.PageReply{
		Curr:  int32(pgInfo.GetCurr()),
		Size:  int32(pgInfo.GetSize()),
		Total: pgInfo.GetTotal(),
	}
	return &pb.ListStrategyReply{Page: pg, List: list}, nil
}

func (s *StrategyService) SelectStrategy(ctx context.Context, req *pb.SelectStrategyRequest) (*pb.SelectStrategyReply, error) {
	strategyBos, pgInfo, err := s.strategyBiz.SelectStrategy(ctx, req)
	if err != nil {
		return nil, err
	}
	list := biz.ListToApiPromStrategySelectV1(strategyBos...)
	pg := &api.PageReply{
		Curr:  int32(pgInfo.GetCurr()),
		Size:  int32(pgInfo.GetSize()),
		Total: pgInfo.GetTotal(),
	}
	return &pb.SelectStrategyReply{Page: pg, List: list}, nil
}

func (s *StrategyService) ExportStrategy(ctx context.Context, req *pb.ExportStrategyRequest) (*pb.ExportStrategyReply, error) {
	fmt.Println("ids: ", req.GetIds())
	filename := "config.yaml"

	var buff []byte
	var err error
	if buff, err = os.ReadFile(filename); err != nil {
		return nil, err
	}

	return &pb.ExportStrategyReply{
		File:     buff,
		FileName: filename,
	}, nil
}
