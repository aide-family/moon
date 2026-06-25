package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// UserMenuJoinRow 用于接收 JOIN 查询结果的结构体
type UserMenuJoinRow struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// UserMenuJoinToBo 将 JOIN 查询结果转换为业务对象
func UserMenuJoinToBo(row *UserMenuJoinRow) *bo.UserMenuItemBo {
	if row == nil {
		return nil
	}
	return &bo.UserMenuItemBo{
		MenuCode:       row.MenuCode,
		ParentMenuCode: row.ParentMenuCode,
		MenuName:       row.MenuName,
		MenuStatus:     row.MenuStatus,
	}
}

// UserMenuItemBoToAPIV1 将业务对象转换为 proto 响应
func UserMenuItemBoToAPIV1(b *bo.UserMenuItemBo) *goddessv1.UserMenuItem {
	return &goddessv1.UserMenuItem{
		MenuCode:       b.MenuCode,
		ParentMenuCode: b.ParentMenuCode,
		MenuName:       b.MenuName,
		MenuStatus:     b.MenuStatus,
	}
}
