package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
)

type (
	IRuleRepo interface {
		V1Repo
	}

	RuleLogic struct {
		logger *log.Helper
		repo   IRuleRepo
	}
)

var _ service.IRuleLogic = (*RuleLogic)(nil)

func NewRuleLogic(repo IRuleRepo, logger log.Logger) *RuleLogic {
	return &RuleLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Rule"))}
}

func (s *RuleLogic) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "RuleLogic.CreateRule")
	defer span.End()
	return nil, nil
}
func (s *RuleLogic) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "RuleLogic.UpdateRule")
	defer span.End()
	return nil, nil
}
func (s *RuleLogic) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "RuleLogic.DeleteRule")
	defer span.End()
	return nil, nil
}
func (s *RuleLogic) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "RuleLogic.GetRule")
	defer span.End()
	return nil, nil
}
func (s *RuleLogic) ListRule(ctx context.Context, req *pb.ListRuleRequest) (*pb.ListRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "RuleLogic.ListRule")
	defer span.End()
	return nil, nil
}
