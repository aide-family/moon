package resource

import (
	"context"

	"github.com/aide-cloud/moon/api/admin"
	pb "github.com/aide-cloud/moon/api/admin/resource"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/build"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

type Service struct {
	pb.UnimplementedResourceServer

	resourceBiz *biz.ResourceBiz
}

func NewResourceService(resourceBiz *biz.ResourceBiz) *Service {
	return &Service{
		resourceBiz: resourceBiz,
	}
}

func (s *Service) GetResource(ctx context.Context, req *pb.GetResourceRequest) (*pb.GetResourceReply, error) {
	resourceDo, err := s.resourceBiz.GetResource(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &pb.GetResourceReply{
		Resource: build.NewResourceBuild(resourceDo).ToApi(),
	}, nil
}

func (s *Service) ListResource(ctx context.Context, req *pb.ListResourceRequest) (*pb.ListResourceReply, error) {
	queryParams := &bo.QueryResourceListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
	}
	resourceDos, err := s.resourceBiz.ListResource(ctx, queryParams)
	if err != nil {
		return nil, err
	}
	return &pb.ListResourceReply{
		Pagination: build.NewPageBuild(queryParams.Page).ToApi(),
		List: types.SliceTo(resourceDos, func(item *model.SysAPI) *admin.ResourceItem {
			return build.NewResourceBuild(item).ToApi()
		}),
	}, nil
}

func (s *Service) BatchUpdateResourceStatus(ctx context.Context, req *pb.BatchUpdateResourceStatusRequest) (*pb.BatchUpdateResourceStatusReply, error) {
	if err := s.resourceBiz.UpdateResourceStatus(ctx, vobj.Status(req.GetStatus()), req.GetIds()...); err != nil {
		return nil, err
	}
	return &pb.BatchUpdateResourceStatusReply{}, nil
}

func (s *Service) GetResourceSelectList(ctx context.Context, req *pb.GetResourceSelectListRequest) (*pb.GetResourceSelectListReply, error) {
	queryParams := &bo.QueryResourceListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
	}
	resourceDos, err := s.resourceBiz.GetResourceSelectList(ctx, queryParams)
	if err != nil {
		return nil, err
	}

	return &pb.GetResourceSelectListReply{
		Pagination: build.NewPageBuild(queryParams.Page).ToApi(),
		List: types.SliceTo(resourceDos, func(item *bo.SelectOptionBo) *admin.Select {
			return build.NewSelectBuild(item).ToApi()
		}),
	}, nil
}
