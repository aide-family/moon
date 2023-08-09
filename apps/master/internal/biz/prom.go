package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/service"
	"prometheus-manager/dal/model"
)

type (
	IPromRepo interface {
		V1Repo
		CreateGroup(ctx context.Context, m *model.PromGroup) error
		UpdateGroupByID(ctx context.Context, id int32, m map[string]any) error
		DeleteGroupByID(ctx context.Context, id int32) error
		GroupDetail(ctx context.Context, id int32) (*model.PromGroup, error)

		CreateStrategy(ctx context.Context, m *model.PromStrategy) error
		UpdateStrategyByID(ctx context.Context, id int32, m map[string]any) error
		DeleteStrategyByID(ctx context.Context, id int32) error
		StrategyDetail(ctx context.Context, id int32) (*model.PromStrategy, error)
	}

	PromLogic struct {
		logger *log.Helper
		repo   IPromRepo
	}
)

var _ service.IPromLogic = (*PromLogic)(nil)

func NewPromLogic(repo IPromRepo, logger log.Logger) *PromLogic {
	return &PromLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Prom"))}
}

func (s *PromLogic) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CreateGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.CreateGroup")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.UpdateGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.UpdateGroup")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.DeleteGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.DeleteGroup")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.GetGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.GetGroup")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.ListGroupReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.ListGroup")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) CreateStrategy(ctx context.Context, req *pb.CreateStrategyRequest) (*pb.CreateStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.CreateStrategy")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) UpdateStrategy(ctx context.Context, req *pb.UpdateStrategyRequest) (*pb.UpdateStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.UpdateStrategy")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyRequest) (*pb.DeleteStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.DeleteStrategy")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) GetStrategy(ctx context.Context, req *pb.GetStrategyRequest) (*pb.GetStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.GetStrategy")
	defer span.End()
	return nil, nil
}

func (s *PromLogic) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) (*pb.ListStrategyReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "PromLogic.ListStrategy")
	defer span.End()
	return nil, nil
}
