package promservice

import (
	"context"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/api"
	pb "prometheus-manager/api/server/prom/strategy"
	"prometheus-manager/pkg/strategy"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
)

type StrategyService struct {
	pb.UnimplementedStrategyServer

	log *log.Helper

	strategyBiz *biz.StrategyBiz
}

func NewStrategyService(strategyBiz *biz.StrategyBiz, logger log.Logger) *StrategyService {
	return &StrategyService{
		log:         log.NewHelper(log.With(logger, "module", "service.prom.strategy")),
		strategyBiz: strategyBiz,
	}
}

func (s *StrategyService) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	strategyBO, err := s.strategyBiz.CreateStrategy(ctx, &bo.StrategyBO{
		Alert:        req.GetAlert(),
		Expr:         req.GetExpr(),
		Duration:     bo.BuildApiDurationString(req.GetDuration()),
		Labels:       strategy.MapToLabels(req.GetLabels()),
		Annotations:  strategy.MapToAnnotations(req.GetAnnotations()),
		Remark:       req.GetRemark(),
		GroupId:      req.GetGroupId(),
		AlarmLevelId: req.GetAlarmLevelId(),
		AlarmPageIds: req.GetAlarmPageIds(),
		CategoryIds:  req.GetCategoryIds(),
		EndpointId:   req.GetDataSourceId(),
		MaxSuppress:  bo.BuildApiDurationString(req.GetMaxSuppress()),
		SendInterval: bo.BuildApiDurationString(req.GetSendInterval()),
		SendRecover:  vo.NewIsSendRecover(req.GetSendRecover()),
	})

	if err != nil {
		return nil, err
	}
	return &pb.CreateStrategyReply{Id: strategyBO.Id}, nil
}

func (s *StrategyService) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	strategyBO, err := s.strategyBiz.UpdateStrategyById(ctx, req.GetId(), &bo.StrategyBO{
		Id:           req.GetId(),
		Alert:        req.GetAlert(),
		Expr:         req.GetExpr(),
		Duration:     bo.BuildApiDurationString(req.GetDuration()),
		Labels:       strategy.MapToLabels(req.GetLabels()),
		Annotations:  strategy.MapToAnnotations(req.GetAnnotations()),
		Remark:       req.GetRemark(),
		GroupId:      req.GetGroupId(),
		AlarmLevelId: req.GetAlarmLevelId(),
		AlarmPageIds: req.GetAlarmPageIds(),
		CategoryIds:  req.GetCategoryIds(),
		EndpointId:   req.GetDataSourceId(),
		MaxSuppress:  bo.BuildApiDurationString(req.GetMaxSuppress()),
		SendInterval: bo.BuildApiDurationString(req.GetSendInterval()),
		SendRecover:  vo.NewIsSendRecover(req.GetSendRecover()),
	})
	if err != nil {
		return nil, err
	}

	return &pb.UpdateStrategyReply{Id: strategyBO.Id}, nil
}

func (s *StrategyService) BatchUpdateStrategyStatus(ctx context.Context, req *pb.BatchUpdateStrategyStatusRequest) (*pb.BatchUpdateStrategyStatusReply, error) {
	if err := s.strategyBiz.BatchUpdateStrategyStatusByIds(ctx, vo.Status(req.GetStatus()), req.GetIds()); err != nil {
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

	detail := strategyBO.ToApiV1()

	return &pb.GetStrategyReply{Detail: detail}, nil
}

func (s *StrategyService) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	pageReq := req.GetPage()
	pgInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	strategyBos, err := s.strategyBiz.ListStrategy(ctx, &bo.ListStrategyRequest{
		Page:       pgInfo,
		Keyword:    req.GetKeyword(),
		GroupId:    req.GetGroupId(),
		Status:     vo.Status(req.GetStatus()),
		StrategyId: req.GetStrategyId(),
	})
	if err != nil {
		return nil, err
	}

	list := bo.ListToApiPromStrategyV1(strategyBos...)
	pg := &api.PageReply{
		Curr:  pgInfo.GetCurr(),
		Size:  pgInfo.GetSize(),
		Total: pgInfo.GetTotal(),
	}
	return &pb.ListStrategyReply{Page: pg, List: list}, nil
}

func (s *StrategyService) SelectStrategy(ctx context.Context, req *pb.SelectStrategyRequest) (*pb.SelectStrategyReply, error) {
	pageReq := req.GetPage()
	pgInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	strategyBos, err := s.strategyBiz.SelectStrategy(ctx, &bo.SelectStrategyRequest{
		Page:    pgInfo,
		Keyword: req.GetKeyword(),
	})
	if err != nil {
		return nil, err
	}
	list := bo.ListToApiPromStrategySelectV1(strategyBos...)
	pg := &api.PageReply{
		Curr:  pgInfo.GetCurr(),
		Size:  pgInfo.GetSize(),
		Total: pgInfo.GetTotal(),
	}
	return &pb.SelectStrategyReply{Page: pg, List: list}, nil
}

func (s *StrategyService) ExportStrategy(_ context.Context, req *pb.ExportStrategyRequest) (*pb.ExportStrategyReply, error) {
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

// GetStrategyNotifyObject 获取策略的告警对象
func (s *StrategyService) GetStrategyNotifyObject(ctx context.Context, req *pb.GetStrategyNotifyObjectRequest) (*pb.GetStrategyNotifyObjectReply, error) {
	strategyBo, err := s.strategyBiz.GetStrategyWithNotifyObjectById(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetStrategyNotifyObjectReply{
		Detail: strategyBo.ToApiV1(),
		NotifyObjectList: slices.To(strategyBo.GetPromNotifies(), func(item *bo.NotifyBO) *api.NotifyV1 {
			return item.ToApi()
		}),
	}, nil
}

// BindStrategyNotifyObject 绑定策略的告警对象
func (s *StrategyService) BindStrategyNotifyObject(ctx context.Context, req *pb.BindStrategyNotifyObjectRequest) (*pb.BindStrategyNotifyObjectReply, error) {
	if err := s.strategyBiz.BindStrategyNotifyObject(ctx, req.GetId(), req.GetNotifyObjectIds()); err != nil {
		return nil, err
	}
	return &pb.BindStrategyNotifyObjectReply{Id: req.GetId()}, nil
}
