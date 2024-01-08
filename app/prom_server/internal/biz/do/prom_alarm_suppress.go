package do

const TableNamePromAlarmSuppress = "prom_alarm_suppress"

type PromAlarmSuppress struct {
	BaseModel
	RealtimeAlarmID uint32   `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__realtime_alarm_id,priority:1;comment:告警ID"`
	UserID          uint32   `gorm:"column:user_id;type:int unsigned;not null;index:idx__user_id,priority:1;comment:用户ID"`
	UserInfo        *SysUser `gorm:"foreignKey:UserID"`
	SuppressedAt    int64    `gorm:"column:suppressed_at;type:bigint;not null;comment:抑制时间"`
	Remark          string   `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Duration        int64    `gorm:"column:duration;type:bigint;not null;comment:抑制时长"`
}

func (*PromAlarmSuppress) TableName() string {
	return TableNamePromAlarmSuppress
}
