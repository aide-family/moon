package bo

import (
	"github.com/aide-cloud/moon/pkg/vobj"
)

type (
	CreateTeamRoleParams struct {
		TeamID uint32 `json:"teamID"`
		// 角色名称
		Name string `json:"name"`
		// 角色描述
		Remark string `json:"remark"`
		// 角色状态
		Status vobj.Status `json:"status"`
		// 角色权限
		Permissions []uint32 `json:"permissions"`
	}

	UpdateTeamRoleParams struct {
		ID uint32 `json:"id"`
		// 角色名称
		Name string `json:"name"`
		// 角色描述
		Remark string `json:"remark"`
		// 角色权限
		Permissions []uint32 `json:"permissions"`
	}

	ListTeamRoleParams struct {
		TeamID  uint32 `json:"teamID"`
		Keyword string `json:"keyword"`
	}
)
