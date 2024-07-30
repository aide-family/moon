package resource

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	resourceapi "github.com/aide-family/moon/api/admin/resource"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/palace/model"
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
		Resource: build.NewResourceBuilder(resourceDo).ToAPI(),
	}, nil
}

// ListResource 获取资源列表
func (s *Service) ListResource(ctx context.Context, req *resourceapi.ListResourceRequest) (*resourceapi.ListResourceReply, error) {
	queryParams := &bo.QueryResourceListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
	}
	resourceDos, err := s.resourceBiz.ListResource(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}
	return &resourceapi.ListResourceReply{
		Pagination: build.NewPageBuilder(queryParams.Page).ToAPI(),
		List: types.SliceTo(resourceDos, func(item *model.SysAPI) *admin.ResourceItem {
			return build.NewResourceBuilder(item).ToAPI()
		}),
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
func (s *Service) GetResourceSelectList(ctx context.Context, req *resourceapi.GetResourceSelectListRequest) (*resourceapi.GetResourceSelectListReply, error) {
	queryParams := &bo.QueryResourceListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
	}
	resourceDos, err := s.resourceBiz.GetResourceSelectList(ctx, queryParams)
	if !types.IsNil(err) {
		return nil, err
	}

	return &resourceapi.GetResourceSelectListReply{
		Pagination: build.NewPageBuilder(queryParams.Page).ToAPI(),
		List: types.SliceTo(resourceDos, func(item *bo.SelectOptionBo) *admin.SelectItem {
			return build.NewSelectBuilder(item).ToAPI()
		}),
	}, nil
}
