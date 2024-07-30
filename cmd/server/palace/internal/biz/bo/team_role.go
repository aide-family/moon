package bo

import (
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// CreateTeamRoleParams 创建团队角色
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

	// UpdateTeamRoleParams 更新团队角色
	UpdateTeamRoleParams struct {
		ID uint32 `json:"id"`
		// 角色名称
		Name string `json:"name"`
		// 角色描述
		Remark string `json:"remark"`
		// 角色权限
		Permissions []uint32 `json:"permissions"`
	}

	// ListTeamRoleParams 获取团队角色列表
	ListTeamRoleParams struct {
		TeamID  uint32 `json:"teamID"`
		Keyword string `json:"keyword"`
	}
)
