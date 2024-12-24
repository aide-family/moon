package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameAlarmPageSelf = "alarm_page_self"

// AlarmPageSelf mapped from table <alarm_page_self>
type AlarmPageSelf struct {
	AllFieldModel
	UserID      uint32 `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__alarm_page__id,priority:1;comment:用户ID" json:"user_id"`
	MemberID    uint32 `gorm:"column:member_id;type:int unsigned;not null;comment:成员ID" json:"member_id"`
	Sort        uint32 `gorm:"column:sort;type:int unsigned;not null;default:0;comment:排序(值越小越靠前， 默认为0)" json:"sort"`
	AlarmPageID uint32 `gorm:"column:alarm_page_id;type:int unsigned;not null;uniqueIndex:idx__user_id__alarm_page__id,priority:2;comment:报警页面ID" json:"alarm_page_id"`

	Member    *SysTeamMember `gorm:"foreignKey:MemberID" json:"member"`
	AlarmPage *SysDict       `gorm:"foreignKey:AlarmPageID" json:"alarm_page"`
}

// String json string
func (c *AlarmPageSelf) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *AlarmPageSelf) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmPageSelf) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName Dashboard's table name
func (*AlarmPageSelf) TableName() string {
	return tableNameAlarmPageSelf
}
