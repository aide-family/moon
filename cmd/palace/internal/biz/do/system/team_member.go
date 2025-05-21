package system

import (
	"encoding/json"
	"strconv"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.TeamMember = (*TeamMember)(nil)

const tableNameTeamMember = "team_members"

type TeamMember struct {
	do.TeamModel
	MemberName string            `gorm:"column:member_name;type:varchar(64);not null;comment:member name" json:"memberName"`
	Remark     string            `gorm:"column:remark;type:varchar(255);not null;comment:remark" json:"remark"`
	UserID     uint32            `gorm:"column:user_id;type:int unsigned;not null;comment:user ID" json:"userID"`
	InviterID  uint32            `gorm:"column:inviter_id;type:int unsigned;not null;comment:inviter ID" json:"inviterID"`
	Position   vobj.Position     `gorm:"column:position;type:tinyint(2);not null;comment:position" json:"position"`
	Status     vobj.MemberStatus `gorm:"column:status;type:tinyint(2);not null;comment:status" json:"status"`
	Roles      []*TeamRole       `gorm:"many2many:sys_team_member_roles" json:"roles"`
	User       *User             `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Inviter    *User             `gorm:"foreignKey:InviterID;references:ID" json:"inviter"`
}

func (u *TeamMember) MarshalBinary() (data []byte, err error) {
	if u == nil {
		return nil, nil
	}
	return json.Marshal(u)
}

func (u *TeamMember) UnmarshalBinary(data []byte) error {
	if u == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, u)
}

func (u *TeamMember) UniqueKey() string {
	return strconv.Itoa(int(u.ID))
}

func (u *TeamMember) GetUser() do.User {
	if u == nil {
		return nil
	}
	return u.User
}

func (u *TeamMember) GetInviter() do.User {
	if u == nil {
		return nil
	}
	return u.Inviter
}

func (u *TeamMember) GetMemberName() string {
	if u == nil {
		return ""
	}
	return u.MemberName
}

func (u *TeamMember) GetRemark() string {
	if u == nil {
		return ""
	}
	return u.Remark
}

func (u *TeamMember) GetTeamMemberID() uint32 {
	if u == nil {
		return 0
	}
	return u.ID
}

func (u *TeamMember) GetUserID() uint32 {
	if u == nil {
		return 0
	}
	return u.UserID
}

func (u *TeamMember) GetInviterID() uint32 {
	if u == nil {
		return 0
	}
	return u.InviterID
}

func (u *TeamMember) GetPosition() vobj.Position {
	if u == nil {
		return vobj.PositionUnknown
	}
	return u.Position
}

func (u *TeamMember) GetStatus() vobj.MemberStatus {
	if u == nil {
		return vobj.MemberStatusUnknown
	}
	return u.Status
}

func (u *TeamMember) GetRoles() []do.TeamRole {
	if u == nil {
		return nil
	}
	return slices.Map(u.Roles, func(r *TeamRole) do.TeamRole { return r })
}

func (u *TeamMember) TableName() string {
	return tableNameTeamMember
}
