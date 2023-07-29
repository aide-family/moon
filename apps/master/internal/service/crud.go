package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/strategy/v1"
)

type (
	ICrudLogic interface {
		CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error)
		UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error)
		DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error)
		RuleDetail(ctx context.Context, req *pb.GetRuleDetailRequest) (*pb.GetRuleDetailReply, error)
		Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error)
	}

	CrudService struct {
		pb.UnimplementedCrudServer

		logger *log.Helper
		logic  ICrudLogic
	}
)

var _ pb.CrudServer = (*CrudService)(nil)

func NewCrudService(logic ICrudLogic, logger log.Logger) *CrudService {
	return &CrudService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Crud"))}
}

func (l *CrudService) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "CrudService.CreateRule")
	defer span.End()
	return l.logic.CreateRule(ctx, req)
}

func (l *CrudService) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "CrudService.UpdateRule")
	defer span.End()
	return l.logic.UpdateRule(ctx, req)
}

func (l *CrudService) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "CrudService.DeleteRule")
	defer span.End()
	return l.logic.DeleteRule(ctx, req)
}

func (l *CrudService) RuleDetail(ctx context.Context, req *pb.GetRuleDetailRequest) (*pb.GetRuleDetailReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "CrudService.RuleDetail")
	defer span.End()
	return l.logic.RuleDetail(ctx, req)
}

func (l *CrudService) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "CrudService.Strategies")
	defer span.End()
	return l.logic.Strategies(ctx, req)
}
