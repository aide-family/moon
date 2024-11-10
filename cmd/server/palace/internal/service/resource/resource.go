package resource

import (
	"context"

	resourceapi "github.com/aide-family/moon/api/admin/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// Service 资源管理服务
type Service struct {
	resourceapi.UnimplementedResourceServer

	resourceBiz *biz.ResourceBiz
}

// NewResourceService 创建资源管理服务
func NewResourceService(resourceBiz *biz.ResourceBiz) *Service {
	return &Service{
		resourceBiz: resourceBiz,
	}
}

// GetResource 获取资源
func (s *Service) GetResource(ctx context.Context, req *resourceapi.GetResourceRequest) (*resourceapi.GetResourceReply, error) {
	resourceDo, err := s.resourceBiz.GetResource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &resourceapi.GetResourceReply{
		Detail: builder.NewParamsBuild(ctx).ResourceModuleBuilder().DoResourceBuilder().ToAPI(resourceDo),
	}, nil
}

// ListResource 获取资源列表
func (s *Service) ListResource(ctx context.Context, req *resourceapi.ListResourceRequest) (*resourceapi.ListResourceReply, error) {
	queryParams := builder.NewParamsBuild(ctx).ResourceModuleBuilder().WithListResourceRequest(req).ToBo()
	resourceDos, err := s.resourceBiz.ListResource(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &resourceapi.ListResourceReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(queryParams.Page),
		List:       builder.NewParamsBuild(ctx).ResourceModuleBuilder().DoResourceBuilder().ToAPIs(resourceDos),
	}, nil
}

// BatchUpdateResourceStatus 批量更新资源状态
func (s *Service) BatchUpdateResourceStatus(ctx context.Context, req *resourceapi.BatchUpdateResourceStatusRequest) (*resourceapi.BatchUpdateResourceStatusReply, error) {
	if err := s.resourceBiz.UpdateResourceStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()...); !types.IsNil(err) {
		return nil, err
	}
	return &resourceapi.BatchUpdateResourceStatusReply{}, nil
}

// GetResourceSelectList 获取资源下拉列表
func (s *Service) GetResourceSelectList(ctx context.Context, req *resourceapi.ListResourceRequest) (*resourceapi.GetResourceSelectListReply, error) {
	queryParams := builder.NewParamsBuild(ctx).ResourceModuleBuilder().WithListResourceRequest(req).ToBo()
	resourceDos, err := s.resourceBiz.ListResource(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}

	return &resourceapi.GetResourceSelectListReply{
		Pagination: builder.NewParamsBuild(ctx).PaginationModuleBuilder().ToAPI(queryParams.Page),
		List:       builder.NewParamsBuild(ctx).ResourceModuleBuilder().DoResourceBuilder().ToSelects(resourceDos),
	}, nil
}
