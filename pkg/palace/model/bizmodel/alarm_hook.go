package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameAlarmHook = "alarm_hook"

// AlarmHook mapped from table <alarm_hook>
type AlarmHook struct {
	AllFieldModel
	Name   string       `gorm:"column:name;type:varchar(64);not null;unique;comment:hook名称" json:"name"`
	Remark string       `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	URL    string       `gorm:"column:url;type:varchar(255);not null;comment:hook URL" json:"url"`
	APP    vobj.HookAPP `gorm:"column:app;type:tinyint;not null;comment:hook应用" json:"app"`
	Status vobj.Status  `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`
	Secret string       `gorm:"column:secret;type:varchar(255);not null;comment:secret" json:"secret"`
}

// String json string
func (c *AlarmHook) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *AlarmHook) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmHook) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName AlarmHook's table name
func (*AlarmHook) TableName() string {
	return tableNameAlarmHook
}
