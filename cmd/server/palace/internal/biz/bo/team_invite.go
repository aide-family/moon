package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	InviteUserParams struct {
		// 邀请userID
		UserID uint32 `json:"userID"`
		// 邀请人角色
		TeamRoleIds []uint32 `json:"sysTeamRoleIds"`
		// 邀请人(手机或邮箱)
		InviteCode string `json:"inviteCode"`
		// 团队id
		TeamID uint32 `json:"teamID"`
	}

	UpdateInviteStatusParams struct {
		// 邀请ID
		InviteID uint32 `json:"inviteID"`
		// 状态
		InviteType vobj.InviteType `json:"inviteType"`
	}

	QueryInviteListParams struct {
		Page       types.Pagination
		Keyword    string          `json:"keyword"`
		InviteType vobj.InviteType `json:"inviteType"`
	}

	QueryInviteParams struct {
		// 团队ID
		TeamID uint32 `json:"teamID"`
		// 邀请userID
		UserID uint32 `json:"userID"`
		// 状态
		InviteType vobj.InviteType `json:"inviteType"`
	}
)
