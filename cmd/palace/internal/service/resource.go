package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
)

type ResourceService struct {
	palace.UnimplementedResourceServer

	resourceBiz *biz.ResourceBiz
	helper      *log.Helper
}

func NewResourceService(resourceBiz *biz.ResourceBiz, logger log.Logger) *ResourceService {
	return &ResourceService{
		resourceBiz: resourceBiz,
		helper:      log.NewHelper(log.With(logger, "module", "service.resource")),
	}
}

func (s *ResourceService) BatchUpdateResourceStatus(ctx context.Context, req *palace.BatchUpdateResourceStatusRequest) (*common.EmptyReply, error) {
	updateReq := &bo.BatchUpdateResourceStatusReq{
		IDs:    req.GetResourceIds(),
		Status: vobj.GlobalStatus(req.GetStatus()),
	}

	err := s.resourceBiz.BatchUpdateResourceStatus(ctx, updateReq)
	if err != nil {
		return nil, err
	}

	return &common.EmptyReply{Message: "更改资源接口状态成功"}, nil
}

func (s *ResourceService) GetResource(ctx context.Context, req *palace.GetResourceRequest) (*common.ResourceItem, error) {
	resource, err := s.resourceBiz.GetResource(ctx, req.GetResourceId())
	if err != nil {
		return nil, err
	}

	return build.ToResourceItem(resource), nil
}

func (s *ResourceService) ListResource(ctx context.Context, req *palace.ListResourceRequest) (*palace.ListResourceReply, error) {
	listReq := &bo.ListResourceReq{
		Statuses:          build.Statuses(req.GetStatus()),
		Keyword:           req.GetKeyword(),
		PaginationRequest: build.ToPaginationRequest(req.GetPagination()),
	}

	resourcesReply, err := s.resourceBiz.ListResource(ctx, listReq)
	if err != nil {
		return nil, err
	}

	return &palace.ListResourceReply{
		Items:      build.ToResourceItems(resourcesReply.Items),
		Pagination: build.ToPaginationReply(resourcesReply.PaginationReply),
	}, nil
}

func (s *ResourceService) GetResourceMenuTree(ctx context.Context, _ *common.EmptyRequest) (*palace.GetResourceMenuTreeReply, error) {
	menus, err := s.resourceBiz.Menus(ctx, vobj.MenuTypeMenuSystem)
	if err != nil {
		return nil, err
	}
	return &palace.GetResourceMenuTreeReply{
		Menus: build.ToMenuTree(menus),
	}, nil
}

func (s *ResourceService) GetTeamResourceMenuTree(ctx context.Context, _ *common.EmptyRequest) (*palace.GetResourceMenuTreeReply, error) {
	menus, err := s.resourceBiz.Menus(ctx, vobj.MenuTypeMenuTeam)
	if err != nil {
		return nil, err
	}
	return &palace.GetResourceMenuTreeReply{
		Menus: build.ToMenuTree(menus),
	}, nil
}

func (s *ResourceService) SaveResource(ctx context.Context, req *palace.SaveResourceRequest) (*common.EmptyReply, error) {
	saveReq := build.ToSaveResourceReq(req)
	if err := s.resourceBiz.SaveResource(ctx, saveReq); err != nil {
		return nil, err
	}

	return &common.EmptyReply{Message: "保存资源成功"}, nil
}

func (s *ResourceService) SaveMenu(ctx context.Context, req *palace.SaveMenuRequest) (*common.EmptyReply, error) {
	saveReq := build.ToSaveMenuReq(req)
	if err := s.resourceBiz.SaveMenu(ctx, saveReq); err != nil {
		return nil, err
	}

	return &common.EmptyReply{Message: "保存系统菜单成功"}, nil
}

func (s *ResourceService) GetMenu(ctx context.Context, req *palace.GetMenuRequest) (*common.MenuTreeItem, error) {
	menu, err := s.resourceBiz.GetMenu(ctx, req.GetMenuId())
	if err != nil {
		return nil, err
	}

	return build.ToMenuTreeItem(menu), nil
}
