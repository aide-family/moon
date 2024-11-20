package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	menuapi "github.com/aide-family/moon/api/admin/menu"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IMenuModuleBuilder = (*menuModuleBuilder)(nil)

type (
	menuModuleBuilder struct {
		ctx context.Context
	}

	// IMenuModuleBuilder 菜单模块构造器
	IMenuModuleBuilder interface {
		// WithCreateMenuRequest 创建菜单请求参数构造器
		WithCreateMenuRequest(*menuapi.CreateMenuRequest) ICreateMenuRequestBuilder
		// WithBatchCreateMenuRequest 批量创建菜单请求参数构造器
		WithBatchCreateMenuRequest(*menuapi.BatchCreateMenuRequest) IBatchCreateMenuRequestBuilder
		// WithUpdateMenuRequest 更新菜单请求参数构造器
		WithUpdateMenuRequest(*menuapi.UpdateMenuRequest) IUpdateMenuRequestBuilder
		// WithListMenuRequest 获取菜单列表请求参数构造器
		WithListMenuRequest(*menuapi.ListMenuRequest) IListMenuRequestBuilder
		// WithBatchUpdateMenuStatusRequest 批量更新菜单状态请求参数构造器
		WithBatchUpdateMenuStatusRequest(*menuapi.BatchUpdateMenuStatusRequest) IBatchUpdateMenuStatusRequestBuilder
		// WithUpdateMenuTypeParams 更新菜单类型请求参数构造器
		WithUpdateMenuTypeParams(*menuapi.BatchUpdateMenuTypeRequest) IUpdateMenuTypeParams
		// DoMenuBuilder 菜单条目构造器
		DoMenuBuilder() IDoMenuBuilder
	}

	// ICreateMenuRequestBuilder 创建菜单请求参数构造器
	ICreateMenuRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.CreateMenuParams
	}

	createMenuRequestBuilder struct {
		ctx context.Context
		*menuapi.CreateMenuRequest
	}

	// IBatchCreateMenuRequestBuilder 批量创建菜单请求参数构造器
	IBatchCreateMenuRequestBuilder interface {
		// ToBos 转换为业务对象列表
		ToBos() []*bo.CreateMenuParams
	}

	batchCreateMenuRequestBuilder struct {
		ctx context.Context
		*menuapi.BatchCreateMenuRequest
	}

	// IUpdateMenuRequestBuilder 更新菜单请求参数构造器
	IUpdateMenuRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateMenuParams
	}

	updateMenuRequestBuilder struct {
		ctx context.Context
		*menuapi.UpdateMenuRequest
	}

	// IListMenuRequestBuilder 获取菜单列表请求参数构造器
	IListMenuRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryMenuListParams
	}

	listMenuRequestBuilder struct {
		ctx context.Context
		*menuapi.ListMenuRequest
	}

	// IBatchUpdateMenuStatusRequestBuilder 批量更新菜单状态请求参数构造器
	IBatchUpdateMenuStatusRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateMenuStatusParams
	}

	batchUpdateMenuStatusRequestBuilder struct {
		ctx context.Context
		*menuapi.BatchUpdateMenuStatusRequest
	}

	// IUpdateMenuTypeParams 更新菜单类型请求参数构造器
	IUpdateMenuTypeParams interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateMenuTypeParams
	}

	updateMenuTypeParamsBuilder struct {
		ctx context.Context
		*menuapi.BatchUpdateMenuTypeRequest
	}

	// IDoMenuBuilder 菜单条目构造器
	IDoMenuBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*model.SysMenu) *adminapi.MenuItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*model.SysMenu) []*adminapi.MenuItem
		// ToAPITree 转换为API树形对象列表
		ToAPITree([]*model.SysMenu) []*adminapi.MenuTree
	}

	doMenuBuilder struct {
		ctx context.Context
	}
)

func (c *createMenuRequestBuilder) ToBo() *bo.CreateMenuParams {
	if types.IsNil(c) || types.IsNil(c.CreateMenuRequest) {
		return nil
	}

	return &bo.CreateMenuParams{
		Name:       c.GetName(),
		ParentID:   c.GetParentId(),
		Path:       c.GetPath(),
		Icon:       c.GetIcon(),
		Type:       vobj.MenuType(c.GetMenuType()),
		Status:     vobj.Status(c.GetStatus()),
		Sort:       c.GetSort(),
		Level:      c.GetLevel(),
		Permission: c.GetPermission(),
		Component:  c.GetComponent(),
		EnName:     c.GetEnName(),
	}
}

func (m *menuModuleBuilder) WithCreateMenuRequest(request *menuapi.CreateMenuRequest) ICreateMenuRequestBuilder {
	return &createMenuRequestBuilder{ctx: m.ctx, CreateMenuRequest: request}
}

func (d *doMenuBuilder) ToAPI(menu *model.SysMenu) *adminapi.MenuItem {
	if types.IsNil(d) || types.IsNil(menu) {
		return nil
	}

	return &adminapi.MenuItem{
		Id:         menu.ID,
		Name:       menu.Name,
		Path:       menu.Path,
		Icon:       menu.Icon,
		Status:     api.Status(menu.Status),
		ParentId:   menu.ParentID,
		CreatedAt:  menu.CreatedAt.String(),
		UpdatedAt:  menu.UpdatedAt.String(),
		Level:      menu.Level,
		Type:       api.MenuType(menu.Type),
		Component:  menu.Component,
		Permission: menu.Permission,
		Sort:       menu.Sort,
		EnName:     menu.EnName,
	}
}

func (d *doMenuBuilder) ToAPIs(menus []*model.SysMenu) []*adminapi.MenuItem {
	if types.IsNil(d) || types.IsNil(menus) {
		return nil
	}

	return types.SliceTo(menus, func(item *model.SysMenu) *adminapi.MenuItem {
		return d.ToAPI(item)
	})
}

func (d *doMenuBuilder) toAPITree(menus []*model.SysMenu, parentID uint32) []*adminapi.MenuTree {
	if types.IsNil(d) || types.IsNil(menus) {
		return nil
	}

	return types.SliceToWithFilter(menus, func(item *model.SysMenu) (*adminapi.MenuTree, bool) {
		if item.ParentID != parentID {
			return nil, false
		}

		return &adminapi.MenuTree{
			Id:         item.ID,
			Name:       item.Name,
			Path:       item.Path,
			Icon:       item.Icon,
			Status:     api.Status(item.Status),
			ParentId:   item.ParentID,
			CreatedAt:  item.CreatedAt.String(),
			UpdatedAt:  item.UpdatedAt.String(),
			Level:      item.Level,
			Children:   d.toAPITree(menus, item.ID),
			Type:       api.MenuType(item.Type),
			Component:  item.Component,
			Permission: item.Permission,
			Sort:       item.Sort,
			EnName:     item.EnName,
		}, true
	})
}

func (d *doMenuBuilder) ToAPITree(menus []*model.SysMenu) []*adminapi.MenuTree {
	if types.IsNil(d) || types.IsNil(menus) {
		return nil
	}

	return d.toAPITree(menus, 0)
}

func (u *updateMenuTypeParamsBuilder) ToBo() *bo.UpdateMenuTypeParams {
	if types.IsNil(u) || types.IsNil(u.BatchUpdateMenuTypeRequest) {
		return nil
	}

	return &bo.UpdateMenuTypeParams{
		IDs:  u.GetIds(),
		Type: vobj.MenuType(u.GetType()),
	}
}

func (b *batchUpdateMenuStatusRequestBuilder) ToBo() *bo.UpdateMenuStatusParams {
	if types.IsNil(b) || types.IsNil(b.BatchUpdateMenuStatusRequest) {
		return nil
	}

	return &bo.UpdateMenuStatusParams{
		IDs:    b.GetIds(),
		Status: vobj.Status(b.GetStatus()),
	}
}

func (l *listMenuRequestBuilder) ToBo() *bo.QueryMenuListParams {
	if types.IsNil(l) || types.IsNil(l.ListMenuRequest) {
		return nil
	}

	return &bo.QueryMenuListParams{
		Keyword:  l.GetKeyword(),
		Page:     types.NewPagination(l.GetPagination()),
		Status:   vobj.Status(l.GetStatus()),
		MenuType: vobj.MenuType(l.GetMenuType()),
	}
}

func (u *updateMenuRequestBuilder) ToBo() *bo.UpdateMenuParams {
	if types.IsNil(u) || types.IsNil(u.UpdateMenuRequest) {
		return nil
	}

	return &bo.UpdateMenuParams{
		ID:          u.GetId(),
		UpdateParam: NewParamsBuild(u.ctx).MenuModuleBuilder().WithCreateMenuRequest(u.GetData()).ToBo(),
	}
}

func (b *batchCreateMenuRequestBuilder) ToBos() []*bo.CreateMenuParams {
	if types.IsNil(b) || types.IsNil(b.BatchCreateMenuRequest) {
		return nil
	}

	return types.SliceTo(b.GetMenus(), func(item *menuapi.CreateMenuRequest) *bo.CreateMenuParams {
		return NewParamsBuild(b.ctx).MenuModuleBuilder().WithCreateMenuRequest(item).ToBo()
	})
}

func (m *menuModuleBuilder) WithBatchCreateMenuRequest(request *menuapi.BatchCreateMenuRequest) IBatchCreateMenuRequestBuilder {
	return &batchCreateMenuRequestBuilder{ctx: m.ctx, BatchCreateMenuRequest: request}
}

func (m *menuModuleBuilder) WithUpdateMenuRequest(request *menuapi.UpdateMenuRequest) IUpdateMenuRequestBuilder {
	return &updateMenuRequestBuilder{ctx: m.ctx, UpdateMenuRequest: request}
}

func (m *menuModuleBuilder) WithListMenuRequest(request *menuapi.ListMenuRequest) IListMenuRequestBuilder {
	return &listMenuRequestBuilder{ctx: m.ctx, ListMenuRequest: request}
}

func (m *menuModuleBuilder) WithBatchUpdateMenuStatusRequest(request *menuapi.BatchUpdateMenuStatusRequest) IBatchUpdateMenuStatusRequestBuilder {
	return &batchUpdateMenuStatusRequestBuilder{ctx: m.ctx, BatchUpdateMenuStatusRequest: request}
}

func (m *menuModuleBuilder) WithUpdateMenuTypeParams(request *menuapi.BatchUpdateMenuTypeRequest) IUpdateMenuTypeParams {
	return &updateMenuTypeParamsBuilder{ctx: m.ctx, BatchUpdateMenuTypeRequest: request}
}

func (m *menuModuleBuilder) DoMenuBuilder() IDoMenuBuilder {
	return &doMenuBuilder{ctx: m.ctx}
}
