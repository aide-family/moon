package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"
	"prometheus-manager/api/perrors"
	"prometheus-manager/api/prom"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/service"
	"prometheus-manager/dal/model"
)

type (
	IDictV1Repo interface {
		V1Repo
		CreateDict(ctx context.Context, m *model.PromDict) error
		UpdateDictById(ctx context.Context, id int32, m *model.PromDict) error
		UpdateDictByIds(ctx context.Context, ids []int32, status prom.Status) error
		DeleteDictById(ctx context.Context, id int32) error
		GetDictById(ctx context.Context, id int32) (*model.PromDict, error)
		ListDict(ctx context.Context, req *pb.ListDictRequest) ([]*model.PromDict, int64, error)
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

// CreateDict 创建字典
//
//	ctx 上下文
//	req 请求参数
func (s *DictLogic) CreateDict(ctx context.Context, req *pb.CreateDictRequest) (*pb.CreateDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.CreateDict")
	defer span.End()

	if err := s.repo.CreateDict(ctx, dictToModel(req.GetDict())); err != nil {
		s.logger.WithContext(ctx).Errorf("CreateDict error: %v", err)
		return nil, perrors.ErrorLogicCreateDictFailed("创建字典失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.CreateDictReply{Response: &api.Response{Message: "创建字典成功"}}, nil
}

// UpdateDict 更新字典
//
//	ctx 上下文
//	req 请求参数
func (s *DictLogic) UpdateDict(ctx context.Context, req *pb.UpdateDictRequest) (*pb.UpdateDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.UpdateDict")
	defer span.End()

	if err := s.repo.UpdateDictById(ctx, req.GetId(), dictToModel(req.GetDict())); err != nil {
		s.logger.WithContext(ctx).Errorf("UpdateDict error: %v", err)
		return nil, perrors.ErrorLogicEditDictFailed("更新字典失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.UpdateDictReply{Response: &api.Response{Message: "更新字典成功"}}, nil
}

// UpdateDictsStatus 更新字典状态
//
//	ctx 上下文
//	req 请求参数
func (s *DictLogic) UpdateDictsStatus(ctx context.Context, req *pb.UpdateDictsStatusRequest) (*pb.UpdateDictsStatusReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.UpdateDictsStatus")
	defer span.End()

	if err := s.repo.UpdateDictByIds(ctx, req.GetIds(), req.GetStatus()); err != nil {
		s.logger.WithContext(ctx).Errorf("UpdateDictsStatus error: %v", err)
		return nil, perrors.ErrorLogicEditDictFailed("更新字典状态失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.UpdateDictsStatusReply{Response: &api.Response{Message: "更新字典状态成功"}}, nil
}

// DeleteDict 删除字典
//
//	ctx 上下文
//	req 请求参数
func (s *DictLogic) DeleteDict(ctx context.Context, req *pb.DeleteDictRequest) (*pb.DeleteDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.DeleteDict")
	defer span.End()

	if err := s.repo.DeleteDictById(ctx, req.GetId()); err != nil {
		s.logger.WithContext(ctx).Errorf("DeleteDict error: %v", err)
		return nil, perrors.ErrorLogicDeleteDictFailed("删除字典失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.DeleteDictReply{Response: &api.Response{Message: "删除字典成功"}}, nil
}

// GetDict 获取字典
//
//	ctx 上下文
//	req 请求参数
func (s *DictLogic) GetDict(ctx context.Context, req *pb.GetDictRequest) (*pb.GetDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.GetDict")
	defer span.End()

	detail, err := s.repo.GetDictById(ctx, req.GetId())
	if err != nil {
		s.logger.WithContext(ctx).Errorf("GetDict error: %v", err)
		if perrors.IsLogicDataNotFound(err) {
			return nil, perrors.ErrorClientNotFound("字典不存在").WithCause(err).WithMetadata(map[string]string{
				"req": req.String(),
			})
		}

		return nil, perrors.ErrorServerDatabaseError("获取字典失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	return &pb.GetDictReply{Response: &api.Response{Message: "获取字典成功"}, Dict: dictToProm(detail)}, nil
}

// ListDict 列表字典
//
//	ctx 上下文
//	req 请求参数
func (s *DictLogic) ListDict(ctx context.Context, req *pb.ListDictRequest) (*pb.ListDictReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "DictLogic.ListDict")
	defer span.End()

	dicts, total, err := s.repo.ListDict(ctx, req)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("ListDict error: %v", err)
		return nil, perrors.ErrorServerDatabaseError("列表字典失败").WithCause(err).WithMetadata(map[string]string{
			"req": req.String(),
		})
	}

	list := make([]*prom.DictItem, 0, len(dicts))
	for _, dict := range dicts {
		list = append(list, dictToProm(dict))
	}

	return &pb.ListDictReply{
		Response: &api.Response{Message: "获取告警页面列表成功"},
		Result: &api.ListQueryResult{
			Page: &api.PageReply{
				Current: req.GetQuery().GetPage().GetCurrent(),
				Size:    req.GetQuery().GetPage().GetSize(),
				Total:   total,
			},
			// TODO 暂时不返回fields
			Fields: nil,
		},
		Dicts: list,
	}, nil
}
