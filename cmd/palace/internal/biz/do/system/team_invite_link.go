package system

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
)

const tableNameTeamInviteLink = "team_invite_links"

type TeamInviteLink struct {
	do.CreatorModel
	TeamID      uint32    `gorm:"index:idx_team_invite_link_team_id;column:team_id;not null;type:int(10) unsigned;comment:团队ID" json:"teamID"`
	Link        string    `gorm:"column:link;type:varchar(255);not null;comment:链接" json:"link"`
	Position    vobj.Role `gorm:"column:position;type:tinyint(2);not null;comment:职位" json:"position"`
	InviteUsers []*User   `gorm:"many2many:team_invite_link_users" json:"inviteUsers"`
}

func (t *TeamInviteLink) TableName() string {
	return tableNameTeamInviteLink
}
