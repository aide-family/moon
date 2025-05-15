package system

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

var _ do.TeamInviteUser = (*TeamInviteUser)(nil)

const tableNameTeamInviteUser = "team_invite_users"

type TeamInviteUser struct {
	do.CreatorModel
	TeamID       uint32    `gorm:"index:idx_team_invite_user_team_id;column:team_id;not null;type:int(10) unsigned;comment:团队ID" json:"teamID"`
	InviteUserID uint32    `gorm:"index:idx_team_invite_user_invite_user_id;column:invite_user_id;not null;type:int(10) unsigned;comment:被邀请用户ID" json:"inviteUserID"`
	InviteUser   *User     `gorm:"foreignKey:InviteUserID;references:ID" json:"inviteUser"`
	Position     vobj.Role `gorm:"column:position;type:tinyint(2);not null;comment:职位" json:"position"`
	Roles        RoleSlice `gorm:"column:roles;type:text;not null;comment:角色id列表" json:"roles"`
	Team         *Team     `gorm:"foreignKey:TeamID;references:ID" json:"team"`
}

func (t *TeamInviteUser) GetTeamID() uint32 {
	if t == nil {
		return 0
	}
	return t.TeamID
}

func (t *TeamInviteUser) GetInviteUserID() uint32 {
	if t == nil {
		return 0
	}
	return t.InviteUserID
}

func (t *TeamInviteUser) GetPosition() vobj.Role {
	if t == nil {
		return vobj.RoleUnknown
	}
	return t.Position
}

func (t *TeamInviteUser) GetRoles() []uint32 {
	if t == nil {
		return nil
	}
	return t.Roles
}

func (t *TeamInviteUser) GetInviteUser() do.User {
	if t == nil {
		return nil
	}
	return t.InviteUser
}

func (t *TeamInviteUser) GetTeam() do.Team {
	if t == nil {
		return nil
	}
	return t.Team
}

func (t *TeamInviteUser) TableName() string {
	return tableNameTeamInviteUser
}

var _ do.ORMModel = (*RoleSlice)(nil)

type RoleSlice []uint32

func (t *RoleSlice) Scan(src any) error {
	val := ""
	switch origin := src.(type) {
	case string:
		val = origin
	case []byte:
		val = string(origin)
	}
	return json.Unmarshal([]byte(val), t)
}

func (t RoleSlice) Value() (driver.Value, error) {
	return json.Marshal(t)
}
