package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameAlarmGroup = "alarm_group"

// AlarmGroup 告警组
type AlarmGroup struct {
	model.AllFieldModel
	Name        string             `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:告警组名称" json:"name"`
	Status      vobj.Status        `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用" json:"status"`
	Remark      string             `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`
	NoticeUsers []*AlarmNoticeUser `gorm:"foreignKey:AlarmGroupID;comment:通知人信息中间表" json:"notice_users"`
	AlarmHooks  []*AlarmHook       `gorm:"foreignKey:AlarmGroupID;comment:告警hook信息中间表" json:"alarm_hooks"`
}

// UnmarshalBinary redis存储实现
func (c *AlarmGroup) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmGroup) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName AlarmGroup's table name
func (*AlarmGroup) TableName() string {
	return tableNameAlarmGroup
}
