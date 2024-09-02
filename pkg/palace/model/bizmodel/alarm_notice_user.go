package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameAlarmNoticeUsers = "alarm_notice_user"

// AlarmNoticeUser 告警通知用户
type AlarmNoticeUser struct {
	model.AllFieldModel
	AlarmGroup      *AlarmGroup     `gorm:"foreignKey:AlarmGroupID" json:"alarm_group"`
	AlarmNoticeType vobj.NotifyType `gorm:"column:notice_type;type:int;not null;comment:通知类型;" json:"alarm_notice_type"`
	UserID          uint32          `gorm:"column:user_id;type:int;not null;comment:通知人id;uniqueIndex:idx__notice__alarm_group_user_id,priority:1" json:"user_id"`
	AlarmGroupID    uint32          `gorm:"column:alarm_group_id;type:int;comment:告警分组id;uniqueIndex:idx__notice__alarm_group_user_id,priority:2" json:"alarm_group_id"`
}

// UnmarshalBinary redis存储实现
func (c *AlarmNoticeUser) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmNoticeUser) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName AlarmNoticeUser's table name
func (*AlarmNoticeUser) TableName() string {
	return tableNameAlarmNoticeUsers
}
