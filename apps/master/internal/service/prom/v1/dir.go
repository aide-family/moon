package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IDirLogic interface {
		CreateDir(ctx context.Context, req *pb.CreateDirRequest) (*pb.CreateDirReply, error)
		UpdateDir(ctx context.Context, req *pb.UpdateDirRequest) (*pb.UpdateDirReply, error)
		DeleteDir(ctx context.Context, req *pb.DeleteDirRequest) (*pb.DeleteDirReply, error)
		GetDir(ctx context.Context, req *pb.GetDirRequest) (*pb.GetDirReply, error)
		ListDir(ctx context.Context, req *pb.ListDirRequest) (*pb.ListDirReply, error)
	}

	DirService struct {
		pb.UnimplementedDirServer

		logger *log.Helper
		logic  IDirLogic
	}
)

var _ pb.DirServer = (*DirService)(nil)

func NewDirService(logic IDirLogic, logger log.Logger) *DirService {
	return &DirService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Dir"))}
}

func (l *DirService) CreateDir(ctx context.Context, req *pb.CreateDirRequest) (*pb.CreateDirReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DirService.CreateDir")
	defer span.End()
	return l.logic.CreateDir(ctx, req)
}

func (l *DirService) UpdateDir(ctx context.Context, req *pb.UpdateDirRequest) (*pb.UpdateDirReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DirService.UpdateDir")
	defer span.End()
	return l.logic.UpdateDir(ctx, req)
}

func (l *DirService) DeleteDir(ctx context.Context, req *pb.DeleteDirRequest) (*pb.DeleteDirReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DirService.DeleteDir")
	defer span.End()
	return l.logic.DeleteDir(ctx, req)
}

func (l *DirService) GetDir(ctx context.Context, req *pb.GetDirRequest) (*pb.GetDirReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DirService.GetDir")
	defer span.End()
	return l.logic.GetDir(ctx, req)
}

func (l *DirService) ListDir(ctx context.Context, req *pb.ListDirRequest) (*pb.ListDirReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DirService.ListDir")
	defer span.End()
	return l.logic.ListDir(ctx, req)
}
