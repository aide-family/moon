package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	service "prometheus-manager/apps/master/internal/service/prom/v1"
)

type (
	IFileRepo interface {
		V1Repo
	}

	FileLogic struct {
		logger *log.Helper
		repo   IFileRepo
	}
)

var _ service.IFileLogic = (*FileLogic)(nil)

func NewFileLogic(repo IFileRepo, logger log.Logger) *FileLogic {
	return &FileLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/File"))}
}

func (s *FileLogic) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "FileLogic.CreateFile")
	defer span.End()
	return nil, nil
}
func (s *FileLogic) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "FileLogic.UpdateFile")
	defer span.End()
	return nil, nil
}
func (s *FileLogic) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "FileLogic.DeleteFile")
	defer span.End()
	return nil, nil
}
func (s *FileLogic) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "FileLogic.GetFile")
	defer span.End()
	return nil, nil
}
func (s *FileLogic) ListFile(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "FileLogic.ListFile")
	defer span.End()
	return nil, nil
}
