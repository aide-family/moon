package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IFileLogic interface {
		CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileReply, error)
		UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileReply, error)
		DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileReply, error)
		GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileReply, error)
		ListFile(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error)
	}

	FileService struct {
		pb.UnimplementedFileServer

		logger *log.Helper
		logic  IFileLogic
	}
)

var _ pb.FileServer = (*FileService)(nil)

func NewFileService(logic IFileLogic, logger log.Logger) *FileService {
	return &FileService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/File"))}
}

func (l *FileService) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "FileService.CreateFile")
	defer span.End()
	return l.logic.CreateFile(ctx, req)
}

func (l *FileService) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "FileService.UpdateFile")
	defer span.End()
	return l.logic.UpdateFile(ctx, req)
}

func (l *FileService) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "FileService.DeleteFile")
	defer span.End()
	return l.logic.DeleteFile(ctx, req)
}

func (l *FileService) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "FileService.GetFile")
	defer span.End()
	return l.logic.GetFile(ctx, req)
}

func (l *FileService) ListFile(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "FileService.ListFile")
	defer span.End()
	return l.logic.ListFile(ctx, req)
}
