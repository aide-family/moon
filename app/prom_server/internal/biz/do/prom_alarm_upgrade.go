package do

const TableNamePromAlarmUpgrade = "prom_alarm_upgrades"

// PromAlarmUpgrade 告警升级信息
type PromAlarmUpgrade struct {
	BaseModel
	RealtimeAlarmID uint32   `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__u__realtime_alarm_id,priority:1;comment:告警ID"`
	UserID          uint32   `gorm:"column:user_id;type:int unsigned;not null;index:idx__u__user_id,priority:1;comment:用户ID"`
	UserInfo        *SysUser `gorm:"foreignKey:UserID"`
	UpgradedAt      int64    `gorm:"column:upgraded_at;type:bigint;not null;comment:升级时间"`
	Remark          string   `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
}

func (*PromAlarmUpgrade) TableName() string {
	return TableNamePromAlarmUpgrade
}
