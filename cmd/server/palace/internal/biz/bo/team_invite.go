package bo

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// InviteUserParams 邀请用户参数
	InviteUserParams struct {
		// 邀请userID
		UserID uint32 `json:"userID"`
		// 邀请人角色
		TeamRoleIds *types.Slice[uint32] `json:"teamRoleIds"`
		// 邀请人(手机或邮箱)
		InviteCode string `json:"inviteCode"`
		// 团队ID
		TeamID uint32 `json:"teamID"`
		// 固定角色
		Role vobj.Role `json:"role"`
	}

	// UpdateInviteStatusParams 更新邀请状态参数
	UpdateInviteStatusParams struct {
		// 邀请ID
		InviteID uint32 `json:"inviteID"`
		// 状态
		InviteType vobj.InviteType `json:"inviteType"`
	}

	// QueryInviteListParams 查询邀请列表参数
	QueryInviteListParams struct {
		Page       types.Pagination
		Keyword    string          `json:"keyword"`
		InviteType vobj.InviteType `json:"inviteType"`
	}

	// InviteTeamInfoParams 邀请团队信息参数
	InviteTeamInfoParams struct {
		TeamMap   map[uint32]*model.SysTeam
		TeamRoles []*bizmodel.SysTeamRole
	}
)
