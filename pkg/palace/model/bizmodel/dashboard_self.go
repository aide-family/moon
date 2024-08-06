package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameDashboardSelf = "dashboard_self"

// DashboardSelf mapped from table <dashboard_self>
type DashboardSelf struct {
	model.AllFieldModel
	DashboardID uint `gorm:"column:dashboard_id;type:int unsigned;not null;uniqueIndex:idx__user_id__dashboard__id,priority:2;comment:DashboardID" json:"dashboard_id"`
	UserID      uint `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__dashboard__id,priority:1;comment:用户ID" json:"user_id"`
	MemberID    uint `gorm:"column:member_id;type:int unsigned;not null;comment:成员ID" json:"member_id"`
	Sort        uint `gorm:"column:sort;type:int unsigned;not null;default:0;comment:排序(值越小越靠前， 默认为0)" json:"sort"`

	Member    *SysTeamMember `gorm:"foreignKey:MemberID" json:"member"`
	Dashboard *Dashboard     `gorm:"foreignKey:DashboardID" json:"dashboard"`
}

// String json string
func (c *DashboardSelf) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *DashboardSelf) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *DashboardSelf) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName Dashboard's table name
func (*DashboardSelf) TableName() string {
	return tableNameDashboardSelf
}
