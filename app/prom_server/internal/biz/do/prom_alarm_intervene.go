package do

const TableNamePromIntervene = "prom_alarm_intervenes"

// PromAlarmIntervene 告警介入信息
type PromAlarmIntervene struct {
	BaseModel
	RealtimeAlarmID uint32   `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__i__realtime_alarm_id,priority:1;comment:告警ID"`
	UserID          uint32   `gorm:"column:user_id;type:int unsigned;not null;index:idx__i__user_id,priority:1;comment:用户ID"`
	UserInfo        *SysUser `gorm:"foreignKey:UserID"`
	IntervenedAt    int64    `gorm:"column:intervened_at;type:bigint;not null;comment:干预时间"`
	Remark          string   `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
}

func (*PromAlarmIntervene) TableName() string {
	return TableNamePromIntervene
}
