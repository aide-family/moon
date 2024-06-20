package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	CreateTeamParams struct {
		// 团队名称
		Name string `json:"name"`
		// 团队备注
		Remark string `json:"remark"`
		// 创建人
		CreatorID uint32 `json:"creatorID"`
		// 团队logo
		Logo string `json:"logo"`
		// 状态
		Status vobj.Status `json:"status"`
		// 团队负责人
		LeaderID uint32 `json:"leaderID"`
		// 管理员列表
		Admins []uint32 `json:"admins"`
	}

	UpdateTeamParams struct {
		// 团队ID
		ID uint32 `json:"id"`
		// 团队名称
		Name string `json:"name"`
		// 团队备注
		Remark string `json:"remark"`
		// 团队logo
		Logo string `json:"logo"`
		// 状态
		Status vobj.Status `json:"status"`
	}

	QueryTeamListParams struct {
		Page    types.Pagination
		Keyword string      `json:"keyword"`
		Status  vobj.Status `json:"status"`
		// 创建人
		CreatorID uint32 `json:"creatorID"`
		// 负责人
		LeaderID uint32 `json:"leaderID"`

		// 指定人员的团队列表
		UserID uint32 `json:"userID"`
		// 团队ID列表
		IDs []uint32 `json:"ids"`
	}

	AddTeamMemberItem struct {
		// 用户ID
		UserID uint32 `json:"userID"`
		// 是否为管理员
		Role vobj.Role `json:"role"`
		// 角色列表
		RoleIds []uint32 `json:"roleIds"`
	}

	AddTeamMemberParams struct {
		// 团队ID
		ID uint32 `json:"id"`
		// 成员列表
		Members []*AddTeamMemberItem `json:"members"`
	}

	RemoveTeamMemberParams struct {
		// 团队ID
		ID uint32 `json:"id"`
		// 成员列表
		MemberIds []uint32 `json:"memberIds"`
	}

	SetMemberAdminParams struct {
		// 团队ID
		ID uint32 `json:"id"`
		// 成员列表
		MemberIds []uint32 `json:"memberIds"`
		// 是否为管理员
		Role vobj.Role `json:"role"`
	}

	SetMemberRoleParams struct {
		// 团队ID
		ID uint32 `json:"id"`
		// 成员列表
		MemberID uint32   `json:"memberID"`
		RoleIds  []uint32 `json:"roleIds"`
	}

	ListTeamMemberParams struct {
		Page types.Pagination
		// 团队ID
		ID uint32 `json:"id"`
		// 模糊查询
		Keyword string `json:"keyword"`
		// 是否为管理员
		Role vobj.Role `json:"role"`
		// 性别
		Gender vobj.Gender `json:"gender"`
		// 状态
		Status vobj.Status `json:"status"`

		// 成员ID列表
		MemberIDs []uint32 `json:"memberIDs"`
	}

	TransferTeamLeaderParams struct {
		// 团队ID
		ID uint32 `json:"id"`
		// 新负责人ID
		LeaderID uint32 `json:"leaderID"`
		// 旧负责人ID
		OldLeaderID uint32 `json:"oldLeaderID"`
	}
)
