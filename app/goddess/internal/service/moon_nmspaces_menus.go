package service

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/goddess/internal/biz"
	"github.com/aide-family/goddess/internal/biz/bo"
	apiv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// NewNmspacesMenuService 创建租户菜单服务实例
func NewNmspacesMenuService(nmspacesMenuBiz *biz.NmspacesMenuBiz, sysMenuBiz *biz.SysMenu) *NmspacesMenuService {
	return &NmspacesMenuService{
		nmspacesMenuBiz: nmspacesMenuBiz,
		sysMenuBiz:      sysMenuBiz,
	}
}

// NmspacesMenuService 租户菜单服务
type NmspacesMenuService struct {
	apiv1.UnimplementedNmspacesMenuServer

	nmspacesMenuBiz *biz.NmspacesMenuBiz
	sysMenuBiz      *biz.SysMenu
}

// getNamespaceUID 从请求头 X-Namespace 中获取租户 UID
func (s *NmspacesMenuService) getNamespaceUID(ctx context.Context) (snowflake.ID, error) {
	if tr, ok := transport.FromServerContext(ctx); ok {
		namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
		namespaceUID, err := snowflake.ParseString(namespace)
		if err == nil && namespaceUID > 0 {
			return namespaceUID, nil
		}
	}
	return 0, merr.ErrorForbidden("namespace is required, please set %s header", cnst.HTTPHeaderXNamespace)
}

// ListNmspacesMenu 分页查询租户菜单列表
func (s *NmspacesMenuService) ListNmspacesMenu(ctx context.Context, req *apiv1.ListNmspacesMenuRequest) (*apiv1.ListNmspacesMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	listBo := bo.NewListNmspacesMenuBo(namespaceUID, req)
	items, total, err := s.nmspacesMenuBiz.ListNmspacesMenu(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListNmspacesMenuReply(items, total, req.Page, req.PageSize), nil
}

// GetNmspacesMenu 查询租户菜单详情
func (s *NmspacesMenuService) GetNmspacesMenu(ctx context.Context, req *apiv1.GetNmspacesMenuRequest) (*apiv1.GetNmspacesMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	result, err := s.nmspacesMenuBiz.GetNmspacesMenu(ctx, namespaceUID, req.MenuCode)
	if err != nil {
		return nil, err
	}
	return &apiv1.GetNmspacesMenuReply{
		Menu: result.ToAPIV1NmspacesMenuItem(),
	}, nil
}

// CreateNmspacesMenu 批量新增租户菜单（从系统菜单表复制）
func (s *NmspacesMenuService) CreateNmspacesMenu(ctx context.Context, req *apiv1.CreateNmspacesMenuRequest) (*apiv1.CreateNmspacesMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	creator := contextx.GetUserUID(ctx)

	var createdCodes []string
	for _, menuCode := range req.MenuCodes {
		// 查询系统菜单详情
		sysMenu, err := s.sysMenuBiz.GetSysMenu(ctx, menuCode)
		if err != nil {
			return nil, err
		}
		if sysMenu == nil {
			return nil, merr.ErrorNotFound("system menu %s not found", menuCode)
		}

		// 构造租户菜单业务对象
		createBo := bo.NewCreateNmspacesMenuBoFromSysMenu(namespaceUID, creator, sysMenu)
		if err := s.nmspacesMenuBiz.CreateNmspacesMenu(ctx, createBo); err != nil {
			return nil, err
		}
		createdCodes = append(createdCodes, menuCode)
	}

	return &apiv1.CreateNmspacesMenuReply{
		MenuCodes: createdCodes,
	}, nil
}

// UpdateNmspacesMenu 编辑租户菜单
func (s *NmspacesMenuService) UpdateNmspacesMenu(ctx context.Context, req *apiv1.UpdateNmspacesMenuRequest) (*apiv1.UpdateNmspacesMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	updateBo := bo.NewUpdateNmspacesMenuBo(namespaceUID, req)
	if err := s.nmspacesMenuBiz.UpdateNmspacesMenu(ctx, updateBo); err != nil {
		return nil, err
	}
	return &apiv1.UpdateNmspacesMenuReply{
		MenuCode: req.MenuCode,
	}, nil
}

// DeleteNmspacesMenu 删除租户菜单
func (s *NmspacesMenuService) DeleteNmspacesMenu(ctx context.Context, req *apiv1.DeleteNmspacesMenuRequest) (*apiv1.DeleteNmspacesMenuReply, error) {
	namespaceUID, err := s.getNamespaceUID(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.nmspacesMenuBiz.DeleteNmspacesMenu(ctx, namespaceUID, req.MenuCode); err != nil {
		return nil, err
	}
	return &apiv1.DeleteNmspacesMenuReply{
		Success: true,
	}, nil
}
