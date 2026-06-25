package convert

import (
	"github.com/aide-family/goddess/internal/biz/bo"
	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
)

// RoleMenuJoinRow 用于接收 JOIN 查询结果的结构体
type RoleMenuJoinRow struct {
	MenuCode       string
	ParentMenuCode string
	MenuName       string
	MenuStatus     int32
}

// RoleMenuJoinToBo 将 JOIN 查询结果转换为业务对象
func RoleMenuJoinToBo(row *RoleMenuJoinRow) *bo.RoleMenuItemBo {
	if row == nil {
		return nil
	}
	return &bo.RoleMenuItemBo{
		MenuCode:       row.MenuCode,
		ParentMenuCode: row.ParentMenuCode,
		MenuName:       row.MenuName,
		MenuStatus:     row.MenuStatus,
	}
}

// RoleMenuItemBoToAPIV1 将业务对象转换为 proto 响应
func RoleMenuItemBoToAPIV1(b *bo.RoleMenuItemBo) *goddessv1.RoleMenuItem {
	return &goddessv1.RoleMenuItem{
		MenuCode:       b.MenuCode,
		ParentMenuCode: b.ParentMenuCode,
		MenuName:       b.MenuName,
		MenuStatus:     b.MenuStatus,
	}
}
