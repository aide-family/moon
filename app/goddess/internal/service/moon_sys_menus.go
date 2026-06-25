package service

import (
	"context"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
)

// NewSysMenuService 创建系统菜单服务实例
func NewSysMenuService(sysMenuBiz *biz.SysMenu) *SysMenuService {
	return &SysMenuService{
		sysMenuBiz: sysMenuBiz,
	}
}

// SysMenuService 系统菜单服务，实现 proto 定义的 SysMenuServer 接口
type SysMenuService struct {
	apiv1.UnimplementedSysMenuServer

	sysMenuBiz *biz.SysMenu
}

// ListSysMenu 分页查询系统菜单列表
func (s *SysMenuService) ListSysMenu(ctx context.Context, req *apiv1.ListSysMenuRequest) (*apiv1.ListSysMenuReply, error) {
	listBo := bo.NewListSysMenuBo(req)
	items, total, err := s.sysMenuBiz.ListSysMenu(ctx, listBo)
	if err != nil {
		return nil, err
	}

	// 联查父菜单名称
	parentMenuNames := make(map[string]string)
	for _, item := range items {
		if item.ParentMenuCode != "" {
			if _, ok := parentMenuNames[item.ParentMenuCode]; !ok {
				parent, err := s.sysMenuBiz.GetSysMenu(ctx, item.ParentMenuCode)
				if err == nil && parent != nil {
					parentMenuNames[item.ParentMenuCode] = parent.MenuName
				}
			}
		}
	}

	apiItems := make([]*apiv1.SysMenuListItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, item.ToAPIV1SysMenuListItem(parentMenuNames[item.ParentMenuCode]))
	}
	return &apiv1.ListSysMenuReply{
		Items:    apiItems,
		Total:    total,
		Page:     listBo.Page,
		PageSize: listBo.PageSize,
	}, nil
}

// TreeSysMenu 查询系统菜单树
func (s *SysMenuService) TreeSysMenu(ctx context.Context, req *apiv1.TreeSysMenuRequest) (*apiv1.TreeSysMenuReply, error) {
	treeBo := bo.NewTreeSysMenuBo(req)
	items, err := s.sysMenuBiz.TreeSysMenu(ctx, treeBo)
	if err != nil {
		return nil, err
	}
	roots := bo.BuildMenuTree(items)
	return &apiv1.TreeSysMenuReply{
		Items: roots,
	}, nil
}

// GetSysMenu 查询系统菜单详情
func (s *SysMenuService) GetSysMenu(ctx context.Context, req *apiv1.GetSysMenuRequest) (*apiv1.GetSysMenuReply, error) {
	result, err := s.sysMenuBiz.GetSysMenu(ctx, req.MenuCode)
	if err != nil {
		return nil, err
	}
	childrenCount, err := s.sysMenuBiz.CountChildren(ctx, req.MenuCode)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetSysMenuReply{
		Menu:          result.ToAPIV1SysMenuItem(),
		ChildrenCount: int32(childrenCount),
	}, nil
}

// CreateSysMenu 新增系统菜单
func (s *SysMenuService) CreateSysMenu(ctx context.Context, req *apiv1.CreateSysMenuRequest) (*apiv1.CreateSysMenuReply, error) {
	createBo := bo.NewCreateSysMenuBo(req)
	createBo.Creator = int64(contextx.GetUserUID(ctx))
	if err := s.sysMenuBiz.CreateSysMenu(ctx, createBo); err != nil {
		return nil, err
	}
	return &apiv1.CreateSysMenuReply{
		MenuCode: req.MenuCode,
	}, nil
}

// UpdateSysMenu 编辑系统菜单
func (s *SysMenuService) UpdateSysMenu(ctx context.Context, req *apiv1.UpdateSysMenuRequest) (*apiv1.UpdateSysMenuReply, error) {
	updateBo := bo.NewUpdateSysMenuBo(req)
	if err := s.sysMenuBiz.UpdateSysMenu(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateSysMenuReply{
		MenuCode: req.MenuCode,
	}, nil
}

// DeleteSysMenu 删除系统菜单
func (s *SysMenuService) DeleteSysMenu(ctx context.Context, req *apiv1.DeleteSysMenuRequest) (*apiv1.DeleteSysMenuReply, error) {
	if err := s.sysMenuBiz.DeleteSysMenu(ctx, req.MenuCode); err != nil {
		return nil, err
	}
	return &apiv1.DeleteSysMenuReply{
		Success: true,
	}, nil
}
