package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameAlarmPageSelf = "alarm_page_self"

// AlarmPageSelf mapped from table <alarm_page_self>
type AlarmPageSelf struct {
	model.AllFieldModel
	UserID      uint `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__alarm_page__id,priority:1;comment:用户ID" json:"user_id"`
	MemberID    uint `gorm:"column:member_id;type:int unsigned;not null;comment:成员ID" json:"member_id"`
	Sort        uint `gorm:"column:sort;type:int unsigned;not null;default:0;comment:排序(值越小越靠前， 默认为0)" json:"sort"`
	AlarmPageID uint `gorm:"column:alarm_page_id;type:int unsigned;not null;uniqueIndex:idx__user_id__alarm_page__id,priority:2;comment:报警页面ID" json:"alarm_page_id"`

	Member    *SysTeamMember `gorm:"foreignKey:MemberID" json:"member"`
	AlarmPage *SysDict       `gorm:"foreignKey:AlarmPageID" json:"alarm_page"`
}

// String json string
func (c *AlarmPageSelf) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *AlarmPageSelf) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmPageSelf) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName Dashboard's table name
func (*AlarmPageSelf) TableName() string {
	return tableNameAlarmPageSelf
}
