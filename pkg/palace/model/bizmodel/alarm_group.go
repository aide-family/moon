package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

const tableNameAlarmNoticeGroup = "alarm_notice_group"

// AlarmNoticeGroup 告警通知组
type AlarmNoticeGroup struct {
	model.AllFieldModel
	DeletedAt     soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__alarm_notice_group__name,priority:1;" json:"deleted_at"`
	Name          string                `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__alarm_notice_group__name,priority:1;comment:告警组名称" json:"name"`
	Status        vobj.Status           `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用" json:"status"`
	Remark        string                `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`
	NoticeMembers []*AlarmNoticeMember  `gorm:"foreignKey:AlarmGroupID;comment:通知人信息中间表" json:"notice_members"`
	AlarmHooks    []*AlarmHook          `gorm:"many2many:alarm_group_hook" json:"alarm_hooks"`
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
