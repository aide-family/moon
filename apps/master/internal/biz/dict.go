package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/service"
)

type (
	IDictV1Repo interface {
		V1Repo
	}

	DictLogic struct {
		logger *log.Helper
		repo   IDictV1Repo
	}
)

var _ service.IDictV1Logic = (*DictLogic)(nil)

func NewDictLogic(repo IDictV1Repo, logger log.Logger) *DictLogic {
	return &DictLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Dict"))}
}

func (s *DictLogic) CreateDict(ctx context.Context, req *pb.CreateDictRequest) (*pb.CreateDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.CreateDict")
	defer span.End()
	return nil, nil
}
func (s *DictLogic) UpdateDict(ctx context.Context, req *pb.UpdateDictRequest) (*pb.UpdateDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.UpdateDict")
	defer span.End()
	return nil, nil
}
func (s *DictLogic) DeleteDict(ctx context.Context, req *pb.DeleteDictRequest) (*pb.DeleteDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.DeleteDict")
	defer span.End()
	return nil, nil
}
func (s *DictLogic) GetDict(ctx context.Context, req *pb.GetDictRequest) (*pb.GetDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.GetDict")
	defer span.End()
	return nil, nil
}
func (s *DictLogic) ListDict(ctx context.Context, req *pb.ListDictRequest) (*pb.ListDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.ListDict")
	defer span.End()
	return nil, nil
}
