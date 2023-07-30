package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"
	promPB "prometheus-manager/api/prom"
	"prometheus-manager/pkg/times"

	pb "prometheus-manager/api/prom/v1"
	promV1 "prometheus-manager/apps/master/internal/service/prom/v1"
	"prometheus-manager/dal/model"
)

type (
	IDirRepo interface {
		V1Repo
		CreateDir(ctx context.Context, m *model.PromNodeDir) error
		UpdateDirById(ctx context.Context, id uint32, m *model.PromNodeDir) error
		DeleteDirById(ctx context.Context, id uint32) error
		GetDirById(ctx context.Context, id uint32) (*model.PromNodeDir, error)
		ListDir(ctx context.Context, q *DirListQueryParams) ([]*model.PromNodeDir, int64, error)
	}

	DirLogic struct {
		logger *log.Helper
		repo   IDirRepo
	}

	DirListQueryParams struct {
		Offset  int
		Limit   int
		Keyword string
	}
)

var _ promV1.IDirLogic = (*DirLogic)(nil)

func NewDirLogic(repo IDirRepo, logger log.Logger) *DirLogic {
	return &DirLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Dir"))}
}

func (s *DirLogic) CreateDir(ctx context.Context, req *pb.CreateDirRequest) (*pb.CreateDirReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "DirLogic.CreateDir")
	defer span.End()

	if err := s.repo.CreateDir(ctx, s.buildDirModel(req.GetDir())); err != nil {
		s.logger.Errorf("CreateDir error: %v", err)
		return nil, err
	}

	return &pb.CreateDirReply{Response: &api.Response{
		Code:     0,
		Message:  "add dir success",
		Metadata: nil,
		Data:     nil,
	}}, nil
}
func (s *DirLogic) UpdateDir(ctx context.Context, req *pb.UpdateDirRequest) (*pb.UpdateDirReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "DirLogic.UpdateDir")
	defer span.End()

	if err := s.repo.UpdateDirById(ctx, req.GetId(), s.buildDirModel(req.GetDir())); err != nil {
		s.logger.Errorf("UpdateDir error: %v", err)
		return nil, err
	}

	return &pb.UpdateDirReply{Response: &api.Response{Message: "edit dir success"}}, nil
}
func (s *DirLogic) DeleteDir(ctx context.Context, req *pb.DeleteDirRequest) (*pb.DeleteDirReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "DirLogic.DeleteDir")
	defer span.End()

	if err := s.repo.DeleteDirById(ctx, req.GetId()); err != nil {
		s.logger.Errorf("DeleteDir error: %v", err)
		return nil, err
	}

	return &pb.DeleteDirReply{Response: &api.Response{Message: "delete dir success"}}, nil
}
func (s *DirLogic) GetDir(ctx context.Context, req *pb.GetDirRequest) (*pb.GetDirReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "DirLogic.GetDir")
	defer span.End()

	dir, err := s.repo.GetDirById(ctx, req.GetId())
	if err != nil {
		s.logger.Errorf("GetDir error: %v", err)
		return nil, err
	}

	return &pb.GetDirReply{Response: &api.Response{Message: "get dir success"}, Dir: s.buildNodeDirItem(dir)}, nil
}
func (s *DirLogic) ListDir(ctx context.Context, req *pb.ListDirRequest) (*pb.ListDirReply, error) {
	ctx, span := otel.Tracer("biz.prom.v1").Start(ctx, "DirLogic.ListDir")
	defer span.End()

	limit := req.GetPage().GetSize()
	offset := (req.GetPage().GetCurrent() - 1) * limit

	query := &DirListQueryParams{
		Offset:  int(offset),
		Limit:   int(limit),
		Keyword: req.GetParams().GetKeyword(),
	}

	list, total, err := s.repo.ListDir(ctx, query)
	if err != nil {
		s.logger.Errorf("ListFile error: %v", err)
		return nil, err
	}

	items := make([]*promPB.DirItem, 0, len(list))
	for _, item := range list {
		items = append(items, s.buildNodeDirItem(item))
	}

	return &pb.ListDirReply{
		Response: &api.Response{Message: "list file success"},
		Page: &api.PageReply{
			Current: req.GetPage().GetCurrent(),
			Size:    req.GetPage().GetSize(),
			Total:   total,
		}, List: items,
	}, nil
}

func (s *DirLogic) buildDirModel(req *promPB.DirItem) *model.PromNodeDir {
	if req == nil {
		return nil
	}

	return &model.PromNodeDir{
		NodeID: int32(req.NodeId),
		Path:   req.GetPath(),
	}
}

func (s *DirLogic) buildNodeDirItem(m *model.PromNodeDir) *promPB.DirItem {
	if m == nil {
		return nil
	}

	return &promPB.DirItem{
		CreatedAt: times.TimeToUnix(m.CreatedAt),
		UpdatedAt: times.TimeToUnix(m.UpdatedAt),
		Id:        uint32(m.ID),
		NodeId:    uint32(m.NodeID),
		Path:      m.Path,
	}
}
