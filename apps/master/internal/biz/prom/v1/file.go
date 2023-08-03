package biz

import (
	"context"
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
	IFileRepo interface {
		V1Repo
		CreateFile(ctx context.Context, m *model.PromNodeDirFile) error
		UpdateFileById(ctx context.Context, id uint32, m *model.PromNodeDirFile) error
		DeleteFileById(ctx context.Context, id uint32) error
		GetFileById(ctx context.Context, id uint32) (*model.PromNodeDirFile, error)
		ListFile(ctx context.Context, q *FileListQueryParams) ([]*model.PromNodeDirFile, int64, error)
	}

	FileLogic struct {
		logger *log.Helper
		repo   IFileRepo
	}

	FileListQueryParams struct {
		Offset  int
		Limit   int
		Keyword string
	}
)

var _ service.IFileLogic = (*FileLogic)(nil)

func NewFileLogic(repo IFileRepo, logger log.Logger) *FileLogic {
	return &FileLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/File"))}
}

func (s *FileLogic) CreateFile(ctx context.Context, req *pb.CreateFileRequest) (*pb.CreateFileReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "FileLogic.CreateFile")
	defer span.End()

	if err := s.repo.CreateFile(ctx, s.buildFileModel(req.GetFile())); err != nil {
		s.logger.Errorf("CreateFile error: %v", err)
		return nil, err
	}

	return &pb.CreateFileReply{Response: &api.Response{Message: "add file success"}}, nil
}

func (s *FileLogic) UpdateFile(ctx context.Context, req *pb.UpdateFileRequest) (*pb.UpdateFileReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "FileLogic.UpdateFile")
	defer span.End()

	if err := s.repo.UpdateFileById(ctx, req.GetId(), s.buildFileModel(req.GetFile())); err != nil {
		s.logger.Errorf("UpdateFile error: %v", err)
		return nil, err
	}

	return &pb.UpdateFileReply{Response: &api.Response{Message: "update file success"}}, nil
}

func (s *FileLogic) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "FileLogic.DeleteFile")
	defer span.End()

	if err := s.repo.DeleteFileById(ctx, req.GetId()); err != nil {
		s.logger.Errorf("DeleteFile error: %v", err)
		return nil, err
	}

	return &pb.DeleteFileReply{Response: &api.Response{Message: "delete file success"}}, nil
}

func (s *FileLogic) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "FileLogic.GetFile")
	defer span.End()

	m, err := s.repo.GetFileById(ctx, req.GetId())
	if err != nil {
		s.logger.Errorf("GetFile error: %v", err)
		return nil, err
	}

	return &pb.GetFileReply{Response: &api.Response{Message: "get file success"}, File: s.buildFilePB(m)}, nil
}

func (s *FileLogic) ListFile(ctx context.Context, req *pb.ListFileRequest) (*pb.ListFileReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "FileLogic.ListFile")
	defer span.End()

	limit := req.GetPage().GetSize()
	offset := (req.GetPage().GetCurrent() - 1) * limit

	query := &FileListQueryParams{
		Offset:  int(offset),
		Limit:   int(limit),
		Keyword: req.GetParams().GetKeyword(),
	}

	list, total, err := s.repo.ListFile(ctx, query)
	if err != nil {
		s.logger.Errorf("ListFile error: %v", err)
		return nil, err
	}

	items := make([]*promPB.FileItem, 0, len(list))
	for _, item := range list {
		items = append(items, s.buildFilePB(item))
	}

	return &pb.ListFileReply{
		Response: &api.Response{Message: "list file success"},
		Page: &api.PageReply{
			Current: req.GetPage().GetCurrent(),
			Size:    req.GetPage().GetSize(),
			Total:   total,
		}, List: items,
	}, nil
}

func (s *FileLogic) buildFileModel(req *promPB.FileItem) *model.PromNodeDirFile {
	if req == nil {
		return nil
	}

	return &model.PromNodeDirFile{
		Filename: req.GetFilename(),
		DirID:    int32(req.GetDirId()),
	}
}

func (s *FileLogic) buildFilePB(m *model.PromNodeDirFile) *promPB.FileItem {
	if m == nil {
		return nil
	}

	return &promPB.FileItem{
		Id:        uint32(m.ID),
		Filename:  m.Filename,
		DirId:     uint32(m.DirID),
		CreatedAt: times.TimeToUnix(m.CreatedAt),
		UpdatedAt: times.TimeToUnix(m.UpdatedAt),
		Groups:    toGroups(m.Groups),
	}
}
