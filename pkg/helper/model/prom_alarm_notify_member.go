package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromNotifyMember = "prom_alarm_notify_members"

type PromAlarmNotifyMember struct {
	query.BaseModel
	PromAlarmNotifyID uint                 `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;index:idx__prom_alarm_notify_id,priority:1;comment:通知ID"`
	Status            valueobj.Status      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	NotifyTypes       valueobj.NotifyTypes `gorm:"column:notify_types;type:json;not null;comment:通知方式"`
	MemberId          uint                 `gorm:"column:member_id;type:int unsigned;not null;index:idx__member_id,priority:1;comment:成员ID"`
	Member            *SysUser             `gorm:"foreignKey:MemberId;comment:成员"`
}

func (*PromAlarmNotifyMember) TableName() string {
	return TableNamePromNotifyMember
}
