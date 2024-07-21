package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	menuapi "github.com/aide-family/moon/api/admin/menu"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	MenuModelBuilder interface {
		ToApi() *admin.Menu
	}

	MenuRequestBuilder interface {
		ToCreateMenuBO(menu *menuapi.CreateMenuRequest) *bo.CreateMenuParams

		ToBatchCreateMenuBO() []*bo.CreateMenuParams

		ToUpdateMenuBO() *bo.UpdateMenuParams
	}

	MenuTreeBuilder interface {
		ToTree() []*admin.MenuTree
	}

	menuBuilder struct {
		Menu                   *model.SysMenu
		BatchCreateMenuRequest *menuapi.BatchCreateMenuRequest
		UpdateMenuRequest      *menuapi.UpdateMenuRequest
		CreateMenuRequest      *menuapi.CreateMenuRequest
		ctx                    context.Context
	}

	menuTreeBuilder struct {
		MenuMap  map[uint32][]*admin.Menu
		ParentID uint32
		ctx      context.Context
	}
)

func (b *menuBuilder) ToApi() *admin.Menu {
	if types.IsNil(b) || types.IsNil(b.Menu) {
		return nil
	}
	return &admin.Menu{
		Id:         b.Menu.ID,
		Name:       b.Menu.Name,
		Path:       b.Menu.Path,
		Icon:       b.Menu.Icon,
		Status:     api.Status(b.Menu.Status),
		ParentId:   b.Menu.ParentID,
		Sort:       b.Menu.Sort,
		Type:       api.MenuType(b.Menu.Type),
		Level:      b.Menu.Level,
		Component:  b.Menu.Component,
		Permission: b.Menu.Permission,
		EnName:     b.Menu.EnName,
		CreatedAt:  b.Menu.CreatedAt.String(),
		UpdatedAt:  b.Menu.UpdatedAt.String(),
	}
}

func (b *menuBuilder) ToCreateMenuBO(menu *menuapi.CreateMenuRequest) *bo.CreateMenuParams {
	return &bo.CreateMenuParams{
		Name:       menu.GetName(),
		Path:       menu.GetPath(),
		Component:  menu.GetComponent(),
		Type:       vobj.MenuType(menu.GetMenuType()),
		Status:     vobj.Status(menu.GetStatus()),
		Icon:       menu.GetIcon(),
		Permission: menu.GetPermission(),
		ParentId:   menu.GetParentId(),
		EnName:     menu.GetEnName(),
		Sort:       menu.GetSort(),
		Level:      menu.GetLevel(),
	}
}

func (b *menuBuilder) ToUpdateMenuBO() *bo.UpdateMenuParams {
	data := b.UpdateMenuRequest.GetData()
	createParams := bo.CreateMenuParams{
		Name:       data.GetName(),
		Path:       data.GetPath(),
		Component:  data.GetComponent(),
		Type:       vobj.MenuType(data.GetMenuType()),
		Status:     vobj.Status(data.GetStatus()),
		Icon:       data.GetIcon(),
		Permission: data.GetPermission(),
		ParentId:   data.GetParentId(),
		EnName:     data.GetEnName(),
		Sort:       data.GetSort(),
		Level:      data.GetLevel(),
	}
	return &bo.UpdateMenuParams{
		ID:          b.UpdateMenuRequest.GetId(),
		UpdateParam: createParams,
	}
}

func (b *menuBuilder) ToBatchCreateMenuBO() []*bo.CreateMenuParams {
	params := types.SliceToWithFilter(b.BatchCreateMenuRequest.GetMenus(), func(menu *menuapi.CreateMenuRequest) (*bo.CreateMenuParams, bool) {
		createParam := bo.CreateMenuParams{
			Name:       menu.GetName(),
			Path:       menu.GetPath(),
			Component:  menu.GetComponent(),
			Type:       vobj.MenuType(menu.GetMenuType()),
			Status:     vobj.Status(menu.GetStatus()),
			Icon:       menu.GetIcon(),
			Permission: menu.GetPermission(),
			ParentId:   menu.GetParentId(),
			EnName:     menu.GetEnName(),
			Sort:       menu.GetSort(),
			Level:      menu.GetLevel(),
		}
		return &createParam, true
	})
	return params
}

// ToTree 转换为树形菜单
func (b *menuTreeBuilder) ToTree() []*admin.MenuTree {
	if types.IsNil(b) || types.IsNil(b.MenuMap) || len(b.MenuMap) == 0 {
		return nil
	}
	list := make([]*admin.MenuTree, 0)
	// 递归遍历
	for _, menu := range b.MenuMap[b.ParentID] {
		if menu.ParentId == b.ParentID {
			list = append(list, &admin.MenuTree{
				Id:        menu.GetId(),
				Name:      menu.GetName(),
				Path:      menu.GetPath(),
				Icon:      menu.GetIcon(),
				Status:    menu.GetStatus(),
				ParentId:  menu.GetParentId(),
				CreatedAt: menu.GetCreatedAt(),
				UpdatedAt: menu.GetUpdatedAt(),
				Level:     menu.GetLevel(),
				Sort:      menu.GetSort(),
				EnName:    menu.GetEnName(),
				Children:  NewBuilder().WithApiMenuTree(b.MenuMap[menu.GetId()], menu.GetId()).ToTree(),
			})
		}
	}
	return list
}
