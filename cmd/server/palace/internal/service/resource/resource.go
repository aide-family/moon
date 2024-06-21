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

type Service struct {
	resourceapi.UnimplementedResourceServer

	resourceBiz *biz.ResourceBiz
}

func NewResourceService(resourceBiz *biz.ResourceBiz) *Service {
	return &Service{
		resourceBiz: resourceBiz,
	}
}

func (s *Service) GetResource(ctx context.Context, req *resourceapi.GetResourceRequest) (*resourceapi.GetResourceReply, error) {
	resourceDo, err := s.resourceBiz.GetResource(ctx, req.GetId())
	if !types.IsNil(err) {
		return nil, err
	}
	return &resourceapi.GetResourceReply{
		Resource: build.NewResourceBuilder(resourceDo).ToApi(),
	}, nil
}

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
		Pagination: build.NewPageBuilder(queryParams.Page).ToApi(),
		List: types.SliceTo(resourceDos, func(item *model.SysAPI) *admin.ResourceItem {
			return build.NewResourceBuilder(item).ToApi()
		}),
	}, nil
}

func (s *Service) BatchUpdateResourceStatus(ctx context.Context, req *resourceapi.BatchUpdateResourceStatusRequest) (*resourceapi.BatchUpdateResourceStatusReply, error) {
	if err := s.resourceBiz.UpdateResourceStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()...); !types.IsNil(err) {
		return nil, err
	}
	return &resourceapi.BatchUpdateResourceStatusReply{}, nil
}

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
		Pagination: build.NewPageBuilder(queryParams.Page).ToApi(),
		List: types.SliceTo(resourceDos, func(item *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuilder(item).ToApi()
		}),
	}, nil
}
