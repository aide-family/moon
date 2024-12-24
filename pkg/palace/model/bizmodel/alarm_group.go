package bizmodel

import (
	"time"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

const tableNameAlarmNoticeGroup = "alarm_notice_group"

var _ types.TimeEngineer = (*AlarmNoticeGroup)(nil)

// AlarmNoticeGroup 告警通知组
type AlarmNoticeGroup struct {
	AllFieldModel
	DeletedAt     soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__alarm_notice_group__name,priority:1;" json:"deleted_at,omitempty"`
	Name          string                `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__alarm_notice_group__name,priority:1;comment:告警组名称" json:"name,omitempty"`
	Status        vobj.Status           `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用" json:"status,omitempty"`
	Remark        string                `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark,omitempty"`
	NoticeMembers []*AlarmNoticeMember  `gorm:"foreignKey:AlarmGroupID;comment:通知人信息中间表" json:"notice_members,omitempty"`
	AlarmHooks    []*AlarmHook          `gorm:"many2many:alarm_group_hook" json:"alarm_hooks,omitempty"`
	TimeEngines   []*TimeEngine         `gorm:"many2many:alarm_group_time_engine" json:"time_engines,omitempty"`
}

// IsAllowed 判断条件是否允许
func (c *AlarmNoticeGroup) IsAllowed(t time.Time) bool {
	if c == nil || len(c.TimeEngines) == 0 {
		return true
	}

	for _, engine := range c.TimeEngines {
		if engine.IsAllowed(t) {
			return true
		}
	}
	return false
}

// UnmarshalBinary redis存储实现
func (c *AlarmNoticeGroup) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *AlarmNoticeGroup) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName AlarmNoticeGroup's table name
func (*AlarmNoticeGroup) TableName() string {
	return tableNameAlarmNoticeGroup
}
