package dict

import (
	"context"

	dictapi "github.com/aide-family/moon/api/admin/dict"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
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
	createParams := builder.NewParamsBuild(ctx).DictModuleBuilder().WithCreateDictRequest(req).ToBo()
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
	updateParams := builder.NewParamsBuild(ctx).DictModuleBuilder().WithUpdateDictRequest(req).ToBo()
	if err := s.dictBiz.UpdateDict(ctx, updateParams); !types.IsNil(err) {
		return nil, err
	}
	return &dictapi.UpdateDictReply{}, nil
}

// ListDict 获取字典列表
func (s *Service) ListDict(ctx context.Context, req *dictapi.ListDictRequest) (*dictapi.ListDictReply, error) {
	queryParams := builder.NewParamsBuild(ctx).DictModuleBuilder().WithListDictRequest(req).ToBo()
	dictList, err := s.dictBiz.ListDict(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}

	return &dictapi.ListDictReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(queryParams.Page),
		List:       builder.NewParamsBuild(ctx).DictModuleBuilder().DoDictBuilder().ToAPIs(dictList),
	}, nil
}

// BatchUpdateDictStatus 批量更新字典状态
func (s *Service) BatchUpdateDictStatus(ctx context.Context, req *dictapi.BatchUpdateDictStatusRequest) (*dictapi.BatchUpdateDictStatusReply, error) {
	updateParams := builder.NewParamsBuild(ctx).DictModuleBuilder().WithUpdateDictStatusParams(req).ToBo()
	err := s.dictBiz.UpdateDictStatusByIds(ctx, updateParams)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return &dictapi.BatchUpdateDictStatusReply{}, nil
}

// DeleteDict 删除字典
func (s *Service) DeleteDict(ctx context.Context, req *dictapi.DeleteDictRequest) (*dictapi.DeleteDictReply, error) {
	if err := s.dictBiz.DeleteDictByID(ctx, req.GetId()); !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
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
		Detail: builder.NewParamsBuild(ctx).DictModuleBuilder().DoDictBuilder().ToAPI(dictDO),
	}, nil
}

// ListDictType 获取字典类型列表
func (s *Service) ListDictType(ctx context.Context, _ *dictapi.ListDictTypeRequest) (*dictapi.ListDictTypeReply, error) {
	return &dictapi.ListDictTypeReply{
		List: builder.NewParamsBuild(ctx).DictModuleBuilder().DictTypeList(),
	}, nil
}

// DictSelectList 获取字典下拉列表
func (s *Service) DictSelectList(ctx context.Context, req *dictapi.ListDictRequest) (*dictapi.DictSelectListReply, error) {
	queryParams := builder.NewParamsBuild(ctx).DictModuleBuilder().WithListDictRequest(req).ToBo()
	dictList, err := s.dictBiz.ListDict(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}

	return &dictapi.DictSelectListReply{
		List: builder.NewParamsBuild(ctx).DictModuleBuilder().DoDictBuilder().ToSelects(dictList),
	}, nil
}
