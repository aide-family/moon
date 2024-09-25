package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamInvite = "sys_team_invites"

// SysTeamInvite mapped from table <sys_team_invites>
type SysTeamInvite struct {
	model.AllFieldModel
	UserID     uint32          `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id__type,priority:1;comment:系统用户ID" json:"user_id"` // 系统用户ID
	TeamID     uint32          `gorm:"column:team_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id__type,priority:2;comment:团队ID" json:"team_id"`   // 团队ID
	InviteType vobj.InviteType `gorm:"column:invite_type;type:int;not null;comment:邀请类型;uniqueIndex:idx__user_id__team__id__type,priority:2;" json:"invite_type"`   // 状态
	TeamRoles  []*SysTeamRole  `gorm:"many2many:sys_team_invite_roles" json:"team_roles"`
}

// String json string
func (c *SysTeamInvite) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamInvite) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamInvite) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysTeamInvite's table name
func (*SysTeamInvite) TableName() string {
	return tableNameSysTeamInvite
}
