package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	promV1 "prometheus-manager/apps/master/internal/service/prom/v1"
)

type (
	IDirRepo interface {
		V1Repo
	}

	DirLogic struct {
		logger *log.Helper
		repo   IDirRepo
	}
)

var _ promV1.IDirLogic = (*DirLogic)(nil)

func NewDirLogic(repo IDirRepo, logger log.Logger) *DirLogic {
	return &DirLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Dir"))}
}

func (s *DirLogic) CreateDir(ctx context.Context, req *pb.CreateDirRequest) (*pb.CreateDirReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DirLogic.CreateDir")
	defer span.End()
	return nil, nil
}
func (s *DirLogic) UpdateDir(ctx context.Context, req *pb.UpdateDirRequest) (*pb.UpdateDirReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DirLogic.UpdateDir")
	defer span.End()
	return nil, nil
}
func (s *DirLogic) DeleteDir(ctx context.Context, req *pb.DeleteDirRequest) (*pb.DeleteDirReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DirLogic.DeleteDir")
	defer span.End()
	return nil, nil
}
func (s *DirLogic) GetDir(ctx context.Context, req *pb.GetDirRequest) (*pb.GetDirReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DirLogic.GetDir")
	defer span.End()
	return nil, nil
}
func (s *DirLogic) ListDir(ctx context.Context, req *pb.ListDirRequest) (*pb.ListDirReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DirLogic.ListDir")
	defer span.End()
	return nil, nil
}
