package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	"github.com/aide-family/goddess/internal/data/impl/do"
)

// NmspacesMenuToBo 将数据库模型转换为业务对象
func NmspacesMenuToBo(menu *do.NmspacesMenu) *bo.NmspacesMenuItemBo {
	if menu == nil {
		return nil
	}
	return &bo.NmspacesMenuItemBo{
		MenuCode:        menu.MenuCode,
		ParentMenuCode:  menu.ParentMenuCode,
		MenuName:        menu.MenuName,
		MenuPath:        menu.MenuPath,
		MenuIcon:        menu.MenuIcon,
		MenuComponent:   menu.MenuComponent,
		MenuRedirect:    menu.MenuRedirect,
		MenuIsHidden:    menu.MenuIsHidden,
		MenuIsKeepAlive: menu.MenuIsKeepAlive,
		MenuSortOrder:   menu.MenuSortOrder,
		MenuStatus:      menu.MenuStatus,
		MenuRemark:      menu.MenuRemark,
		Creator:         menu.Creator.Int64(),
		CreatedAt:       menu.CreatedAt,
		UpdatedAt:       menu.UpdatedAt,
	}
}
