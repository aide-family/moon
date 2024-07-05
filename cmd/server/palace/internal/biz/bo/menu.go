package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateMenuParams struct {
		// 菜单名称
		Name string `json:"name"`
		// 父id
		ParentId uint32 `json:"parent_id"`
		// 路径
		Path string `json:"path"`
		// 图标
		Icon string `json:"icon"`
		// 菜单类型
		Type vobj.MenuType `json:"type"`
		// 状态
		Status vobj.Status `json:"status"`
		// 排序
		Sort int32 `json:"sort"`
		// 级别
		Level int32 `json:"level"`
		// 权限标识
		Permission string `json:"permission"`
		// 组件路径
		Component string `json:"component"`
		// 英文名称
		EnName string `json:"en_name"`
	}

	UpdateMenuParams struct {
		ID          uint32 `json:"id"`
		UpdateParam CreateMenuParams
	}

	UpdateMenuStatusParams struct {
		IDs    []uint32    `json:"ids"`
		Status vobj.Status `json:"status"`
	}

	UpdateMenuTypeParams struct {
		IDs  []uint32      `json:"ids"`
		Type vobj.MenuType `json:"type"`
	}

	QueryMenuListParams struct {
		Keyword  string           `json:"keyword"`
		Page     types.Pagination `json:"page"`
		Status   vobj.Status      `json:"status"`
		MenuType vobj.MenuType    `json:"menu_type"`
	}
)
