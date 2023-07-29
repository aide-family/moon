package biz

import (
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"
	pb "prometheus-manager/api/strategy/v1"
	"prometheus-manager/apps/master/internal/service"
	"prometheus-manager/dal/model"
)

type (
	ICrudRepo interface {
		CreateStrategies(ctx context.Context, m *model.PromNodeDirFileGroupStrategy) error
	}

	CrudLogic struct {
		logger *log.Helper
		repo   ICrudRepo
	}
)

var _ service.ICrudLogic = (*CrudLogic)(nil)

func NewCrudLogic(repo ICrudRepo, logger log.Logger) *CrudLogic {
	return &CrudLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Crud"))}
}

func (s *CrudLogic) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "CrudLogic.CreateStrategies")
	defer span.End()

	createStrategies := s.buildCreateStrategies(req)

	if err := s.repo.CreateStrategies(ctx, createStrategies); err != nil {
		s.logger.Errorf("CreateStrategies error: %v", err)
		return nil, err
	}

	return &pb.CreateRuleReply{Response: &api.Response{
		Code:     0,
		Message:  "success",
		Metadata: nil,
		Data:     nil,
	}}, nil
}
func (s *CrudLogic) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "CrudLogic.UpdateRule")
	defer span.End()
	return nil, nil
}

func (s *CrudLogic) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "CrudLogic.DeleteRule")
	defer span.End()
	return nil, nil
}

func (s *CrudLogic) RuleDetail(ctx context.Context, req *pb.GetRuleDetailRequest) (*pb.GetRuleDetailReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "CrudLogic.RuleDetail")
	defer span.End()
	return nil, nil
}

func (s *CrudLogic) Strategies(ctx context.Context, req *pb.StrategiesRequest) (*pb.StrategiesReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "CrudLogic.Strategies")
	defer span.End()
	return nil, nil
}

// buildCreateStrategies 构建创建策略model数据
func (s *CrudLogic) buildCreateStrategies(req *pb.CreateRuleRequest) *model.PromNodeDirFileGroupStrategy {
	ruleInfo := req.GetRule()
	labelsByte, _ := json.Marshal(ruleInfo.GetLabels())
	annotationsByte, _ := json.Marshal(ruleInfo.GetAnnotations())
	return &model.PromNodeDirFileGroupStrategy{
		GroupID:     int32(req.GetGroupId()),
		Alert:       ruleInfo.GetAlert(),
		Expr:        ruleInfo.GetExpr(),
		For:         ruleInfo.GetFor(),
		Labels:      string(labelsByte),
		Annotations: string(annotationsByte),
	}
}
