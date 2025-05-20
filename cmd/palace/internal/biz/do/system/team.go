package system

import (
	"encoding/json"
	"strconv"

	"github.com/google/uuid"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.Team = (*Team)(nil)

const tableNameTeam = "sys_teams"

type Team struct {
	do.CreatorModel
	Name          string                           `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__team__name,priority:1;comment:team space name" json:"name"`
	Status        vobj.TeamStatus                  `gorm:"column:status;type:tinyint(2);not null;comment:status" json:"status"`
	Remark        string                           `gorm:"column:remark;type:varchar(255);not null;comment:remark" json:"remark"`
	Logo          string                           `gorm:"column:logo;type:varchar(255);not null;comment:team logo" json:"logo"`
	LeaderID      uint32                           `gorm:"column:leader_id;type:int unsigned;not null;index:sys_teams__sys_users,priority:1;comment:leader" json:"leaderID"`
	UUID          uuid.UUID                        `gorm:"column:uuid;type:varchar(36);not null" json:"uuid"`
	Capacity      vobj.TeamCapacity                `gorm:"column:capacity;type:tinyint(2);not null;comment:team capacity (package)" json:"capacity"`
	Leader        *User                            `gorm:"foreignKey:LeaderID;references:ID" json:"leader"`
	Admins        []*User                          `gorm:"many2many:sys_team_admins" json:"admins"`
	BizDBConfig   *crypto.Object[*config.Database] `gorm:"column:biz_db_config;type:text;not null;comment:database name" json:"bizDBConfig"`
	AlarmDBConfig *crypto.Object[*config.Database] `gorm:"column:alarm_db_config;type:text;not null;comment:database name" json:"alarmDBConfig"`
	Menus         []*Menu                          `gorm:"many2many:sys_team_menus" json:"menus"`
}

func (u *Team) MarshalBinary() (data []byte, err error) {
	if u == nil {
		return nil, nil
	}
	return json.Marshal(u)
}

func (u *Team) UnmarshalBinary(data []byte) error {
	if u == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, u)
}

func (u *Team) UniqueKey() string {
	return strconv.Itoa(int(u.ID))
}

func (u *Team) GetLeader() do.User {
	if u == nil {
		return nil
	}
	return u.Leader
}

func (u *Team) GetAdmins() []do.User {
	if u == nil {
		return nil
	}
	return slices.Map(u.Admins, func(v *User) do.User { return v })
}

func (u *Team) GetTeamID() uint32 {
	if u == nil {
		return 0
	}
	return u.ID
}

func (u *Team) GetName() string {
	if u == nil {
		return ""
	}
	return u.Name
}

func (u *Team) GetRemark() string {
	if u == nil {
		return ""
	}
	return u.Remark
}

func (u *Team) GetLogo() string {
	if u == nil {
		return ""
	}
	return u.Logo
}

func (u *Team) GetStatus() vobj.TeamStatus {
	if u == nil {
		return vobj.TeamStatusUnknown
	}
	return u.Status
}

func (u *Team) GetLeaderID() uint32 {
	if u == nil {
		return 0
	}
	return u.LeaderID
}

func (u *Team) GetUUID() uuid.UUID {
	if u == nil {
		return uuid.UUID{}
	}
	return u.UUID
}

func (u *Team) GetCapacity() vobj.TeamCapacity {
	if u == nil {
		return vobj.TeamCapacityUnknown
	}
	return u.Capacity
}

func (u *Team) GetBizDBConfig() *config.Database {
	if u == nil {
		return nil
	}
	return u.BizDBConfig.Get()
}

func (u *Team) GetAlarmDBConfig() *config.Database {
	if u == nil {
		return nil
	}
	return u.AlarmDBConfig.Get()
}

func (u *Team) GetMenus() []do.Menu {
	if u == nil {
		return nil
	}
	return slices.Map(u.Menus, func(v *Menu) do.Menu { return v })
}

func (u *Team) TableName() string {
	return tableNameTeam
}
