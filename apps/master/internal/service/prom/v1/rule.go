package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IRuleLogic interface {
		CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error)
		UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error)
		DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error)
		GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleReply, error)
		ListRule(ctx context.Context, req *pb.ListRuleRequest) (*pb.ListRuleReply, error)
	}

	RuleService struct {
		pb.UnimplementedRuleServer

		logger *log.Helper
		logic  IRuleLogic
	}
)

var _ pb.RuleServer = (*RuleService)(nil)

func NewRuleService(logic IRuleLogic, logger log.Logger) *RuleService {
	return &RuleService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Rule"))}
}

func (l *RuleService) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "RuleService.CreateRule")
	defer span.End()
	return l.logic.CreateRule(ctx, req)
}

func (l *RuleService) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "RuleService.UpdateRule")
	defer span.End()
	return l.logic.UpdateRule(ctx, req)
}

func (l *RuleService) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "RuleService.DeleteRule")
	defer span.End()
	return l.logic.DeleteRule(ctx, req)
}

func (l *RuleService) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "RuleService.GetRule")
	defer span.End()
	return l.logic.GetRule(ctx, req)
}

func (l *RuleService) ListRule(ctx context.Context, req *pb.ListRuleRequest) (*pb.ListRuleReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "RuleService.ListRule")
	defer span.End()
	return l.logic.ListRule(ctx, req)
}
