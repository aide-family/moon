package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromAlarmNotifyMember = "prom_alarm_been_notify_members"

type PromAlarmBeenNotifyMember struct {
	query.BaseModel
	RealtimeAlarmID   uint                 `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__realtime_alarm_id,priority:1;comment:告警ID"`
	NotifyTypes       valueobj.NotifyTypes `gorm:"column:notify_types;type:json;not null;comment:通知方式"`
	MemberId          uint                 `gorm:"column:member_id;type:int unsigned;not null;comment:通知人员ID"`
	Member            *SysUser             `gorm:"foreignKey:MemberId;comment:成员"`
	Msg               string               `gorm:"column:msg;type:text;not null;comment:通知的消息"`
	Status            valueobj.Status      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	PromAlarmNotifyID uint                 `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;comment:通知ID"`
	PromAlarmNotify   *PromAlarmNotify     `gorm:"foreignKey:PromAlarmNotifyID"`
}

func (*PromAlarmBeenNotifyMember) TableName() string {
	return TableNamePromAlarmNotifyMember
}
