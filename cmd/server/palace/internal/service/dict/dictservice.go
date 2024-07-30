package dict

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	dictapi "github.com/aide-family/moon/api/admin/dict"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// Service 字典服务
type Service struct {
	dictapi.UnimplementedDictServer

	dictBiz *biz.DictBiz
}

// NewDictService 创建字典服务
func NewDictService(dictBiz *biz.DictBiz) *Service {
	return &Service{
		dictBiz: dictBiz,
	}
}

// CreateDict 创建字典
func (s *Service) CreateDict(ctx context.Context, req *dictapi.CreateDictRequest) (*dictapi.CreateDictReply, error) {
	createParams := build.NewBuilder().WithCreateBoDict(req).ToCreateDictBO()
	_, err := s.dictBiz.CreateDict(ctx, createParams)
	if err != nil {
		return nil, err
	}
	if !types.IsNil(err) {
		return nil, err
	}
	return &dictapi.CreateDictReply{}, nil
}

// UpdateDict 更新字典
func (s *Service) UpdateDict(ctx context.Context, req *dictapi.UpdateDictRequest) (*dictapi.UpdateDictReply, error) {
	updateParams := build.NewBuilder().WithUpdateBoDict(req).ToUpdateDictBO()
	if err := s.dictBiz.UpdateDict(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &dictapi.UpdateDictReply{}, nil
}

// ListDict 获取字典列表
func (s *Service) ListDict(ctx context.Context, req *dictapi.GetDictSelectListRequest) (*dictapi.ListDictReply, error) {
	queryParams := &bo.QueryDictListParams{
		Keyword:  req.GetKeyword(),
		Page:     types.NewPagination(req.GetPagination()),
		Status:   vobj.Status(req.GetStatus()),
		DictType: vobj.DictType(req.GetDictType()),
	}
	dictPage, err := s.dictBiz.ListDict(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	resList := types.SliceTo(dictPage, func(dict imodel.IDict) *admin.Dict {
		return build.NewBuilder().WithContext(ctx).WithDict(dict).ToAPI()
	})
	return &dictapi.ListDictReply{
		Pagination: build.NewPageBuilder(queryParams.Page).ToAPI(),
		List:       resList,
	}, nil
}

// BatchUpdateDictStatus 批量更新字典状态
func (s *Service) BatchUpdateDictStatus(ctx context.Context, params *dictapi.BatchUpdateDictStatusRequest) (*dictapi.BatchUpdateDictStatusReply, error) {
	updateParams := bo.UpdateDictStatusParams{
		IDs:    params.GetIds(),
		Status: vobj.Status(params.Status),
	}
	err := s.dictBiz.UpdateDictStatusByIds(ctx, &updateParams)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &dictapi.BatchUpdateDictStatusReply{}, nil
}

// DeleteDict 删除字典
func (s *Service) DeleteDict(ctx context.Context, req *dictapi.DeleteDictRequest) (*dictapi.DeleteDictReply, error) {
	if err := s.dictBiz.DeleteDictByID(ctx, req.GetId()); !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return &dictapi.DeleteDictReply{}, nil
}

// GetDict 获取字典详情
func (s *Service) GetDict(ctx context.Context, req *dictapi.GetDictRequest) (*dictapi.GetDictReply, error) {
	dictDO, err := s.dictBiz.GetDict(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &dictapi.GetDictReply{
		Dict: build.NewBuilder().WithContext(ctx).WithDict(dictDO).ToAPI(),
	}, nil
}

// ListDictType 获取字典类型列表
func (s *Service) ListDictType(_ context.Context, _ *dictapi.ListDictTypeRequest) (*dictapi.ListDictTypeReply, error) {
	return &dictapi.ListDictTypeReply{
		List: build.NewDictTypeBuilder().ToAPI(),
	}, nil
}
