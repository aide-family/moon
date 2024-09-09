package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameAlarmNoticeMembers = "alarm_notice_member"

// AlarmNoticeMember 告警通知用户
type AlarmNoticeMember struct {
	model.AllFieldModel
	AlarmGroup      *AlarmNoticeGroup `gorm:"foreignKey:AlarmGroupID" json:"alarm_group"`
	AlarmNoticeType vobj.NotifyType   `gorm:"column:notice_type;type:int;not null;comment:通知类型;" json:"alarm_notice_type"`
	MemberID        uint32            `gorm:"column:member_id;type:int;not null;comment:通知人id;uniqueIndex:idx__notice__alarm_group_member_id,priority:1" json:"member_id"`

	AlarmGroupID uint32 `gorm:"column:alarm_group_id;type:int;comment:告警分组id;uniqueIndex:idx__notice__alarm_group_member_id,priority:2" json:"alarm_group_id"`
}

// UnmarshalBinary redis存储实现
func (c *AlarmNoticeMember) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmNoticeMember) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName AlarmNoticeMember's table name
func (*AlarmNoticeMember) TableName() string {
	return tableNameAlarmNoticeMembers
}
