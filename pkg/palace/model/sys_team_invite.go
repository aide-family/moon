package model

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamInvite = "sys_team_invites"

// SysTeamInvite mapped from table <sys_team_invites>
type SysTeamInvite struct {
	AllFieldModel
	UserID     uint32               `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id__type,priority:1;comment:系统用户ID" json:"user_id"` // 系统用户ID
	TeamID     uint32               `gorm:"column:team_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id__type,priority:2;comment:团队ID" json:"team_id"`   // 团队ID
	InviteType vobj.InviteType      `gorm:"column:invite_type;type:int;not null;comment:邀请类型;uniqueIndex:idx__user_id__team__id__type,priority:2;" json:"invite_type"`   // 状态
	RolesIds   *types.Slice[uint32] `gorm:"column:roles_ids;type:varchar(255);not null;comment:团队角色id数组" json:"roles_ids"`
	Role       vobj.Role            `gorm:"column:role;type:int;not null;comment:角色" json:"role"`
}

// String json string
func (c *SysTeamInvite) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamInvite) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamInvite) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamInvite's table name
func (*SysTeamInvite) TableName() string {
	return tableNameSysTeamInvite
}

// GetRolesIds 获取角色id数组
func (c *SysTeamInvite) GetRolesIds() []uint32 {
	if types.IsNil(c) || types.IsNil(c.RolesIds) {
		return nil
	}
	return c.RolesIds.ToSlice()
}

// GetRole get role
func (c *SysTeamInvite) GetRole() vobj.Role {
	if types.IsNil(c) {
		return vobj.Role(0)
	}
	return c.Role
}

// GetInviteType get invite type
func (c *SysTeamInvite) GetInviteType() vobj.InviteType {
	if types.IsNil(c) {
		return vobj.InviteType(0)
	}
	return c.InviteType
}

// GetUserID get user id
func (c *SysTeamInvite) GetUserID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.UserID
}

// GetTeamID get team id
func (c *SysTeamInvite) GetTeamID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.TeamID
}
