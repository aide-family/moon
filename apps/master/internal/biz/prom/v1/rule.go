package biz

import (
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"
	promPB "prometheus-manager/api/prom"
	"prometheus-manager/pkg/times"

	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
	"prometheus-manager/dal/model"
)

type (
	IRuleRepo interface {
		V1Repo
		CreateRule(ctx context.Context, m *model.PromNodeDirFileGroupStrategy) error
		UpdateRuleById(ctx context.Context, id uint32, m *model.PromNodeDirFileGroupStrategy) error
		DeleteRuleById(ctx context.Context, id uint32) error
		GetRuleById(ctx context.Context, id uint32) (*model.PromNodeDirFileGroupStrategy, error)
		ListRule(ctx context.Context, q *RuleListQueryParams) ([]*model.PromNodeDirFileGroupStrategy, int64, error)
	}

	RuleLogic struct {
		logger *log.Helper
		repo   IRuleRepo
	}

	RuleListQueryParams struct {
		Offset  int
		Limit   int
		Keyword string
	}
)

var _ service.IRuleLogic = (*RuleLogic)(nil)

func NewRuleLogic(repo IRuleRepo, logger log.Logger) *RuleLogic {
	return &RuleLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Rule"))}
}

func (s *RuleLogic) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "RuleLogic.CreateRule")
	defer span.End()

	if err := s.repo.CreateRule(ctx, s.buildRuleModel(req.GetRule())); err != nil {
		s.logger.Errorf("CreateRule error: %v", err)
		return nil, err
	}

	return &pb.CreateRuleReply{Response: &api.Response{Message: "add rule success"}}, nil
}

func (s *RuleLogic) UpdateRule(ctx context.Context, req *pb.UpdateRuleRequest) (*pb.UpdateRuleReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "RuleLogic.UpdateRule")
	defer span.End()

	if err := s.repo.UpdateRuleById(ctx, req.GetId(), s.buildRuleModel(req.GetRule())); err != nil {
		s.logger.Errorf("UpdateRule error: %v", err)
		return nil, err
	}

	return &pb.UpdateRuleReply{Response: &api.Response{Message: "update rule success"}}, nil
}

func (s *RuleLogic) DeleteRule(ctx context.Context, req *pb.DeleteRuleRequest) (*pb.DeleteRuleReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "RuleLogic.DeleteRule")
	defer span.End()

	if err := s.repo.DeleteRuleById(ctx, req.GetId()); err != nil {
		s.logger.Errorf("DeleteRule error: %v", err)
		return nil, err
	}

	return &pb.DeleteRuleReply{Response: &api.Response{Message: "delete rule success"}}, nil
}

func (s *RuleLogic) GetRule(ctx context.Context, req *pb.GetRuleRequest) (*pb.GetRuleReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "RuleLogic.GetRule")
	defer span.End()

	m, err := s.repo.GetRuleById(ctx, req.GetId())
	if err != nil {
		s.logger.Errorf("GetRule error: %v", err)
		return nil, err
	}

	return &pb.GetRuleReply{Response: &api.Response{Message: "get rule success"}, Rule: s.buildRuleItem(m)}, nil
}

func (s *RuleLogic) ListRule(ctx context.Context, req *pb.ListRuleRequest) (*pb.ListRuleReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "RuleLogic.ListRule")
	defer span.End()

	limit := int(req.GetPage().GetSize())
	offset := int(req.GetPage().GetCurrent()*req.GetPage().GetSize()) - limit

	query := &RuleListQueryParams{
		Offset:  offset,
		Limit:   limit,
		Keyword: req.GetParams().GetKeyword(),
	}

	list, total, err := s.repo.ListRule(ctx, query)
	if err != nil {
		s.logger.Errorf("ListRule error: %v", err)
		return nil, err
	}

	items := make([]*promPB.RuleItem, 0, len(list))
	for _, m := range list {
		items = append(items, s.buildRuleItem(m))
	}

	return &pb.ListRuleReply{Response: &api.Response{Message: "list rule success"}, Page: &api.PageReply{
		Current: req.GetPage().GetCurrent(),
		Size:    req.GetPage().GetSize(),
		Total:   total,
	}, List: items}, nil
}

func (s *RuleLogic) buildRuleModel(req *promPB.RuleItem) *model.PromNodeDirFileGroupStrategy {
	if req == nil {
		return nil
	}

	labels, _ := json.Marshal(req.GetLabels())
	annotations, _ := json.Marshal(req.GetAnnotations())

	return &model.PromNodeDirFileGroupStrategy{
		GroupID:     int32(req.GetGroupId()),
		Alert:       req.GetAlert(),
		Expr:        req.GetExpr(),
		For:         req.GetFor(),
		Labels:      string(labels),
		Annotations: string(annotations),
	}
}

func (s *RuleLogic) buildRuleItem(m *model.PromNodeDirFileGroupStrategy) *promPB.RuleItem {
	if m == nil {
		return nil
	}

	labels := make(map[string]string)
	annotations := make(map[string]string)

	if err := json.Unmarshal([]byte(m.Labels), &labels); err != nil {
		s.logger.Errorf("json.Unmarshal error: %v", err)
	}

	if err := json.Unmarshal([]byte(m.Annotations), &annotations); err != nil {
		s.logger.Errorf("json.Unmarshal error: %v", err)
	}

	return &promPB.RuleItem{
		GroupId:     uint32(m.GroupID),
		Alert:       m.Alert,
		Expr:        m.Expr,
		For:         m.For,
		Labels:      labels,
		Annotations: annotations,
		CreatedAt:   times.TimeToUnix(m.CreatedAt),
		UpdatedAt:   times.TimeToUnix(m.UpdatedAt),
		Id:          uint32(m.ID),
	}
}
