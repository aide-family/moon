package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromAlarmNotifyMember = "prom_alarm_been_notify_members"

type PromAlarmBeenNotifyMember struct {
	query.BaseModel
	RealtimeAlarmID   uint32               `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__realtime_alarm_id,priority:1;comment:告警ID"`
	NotifyTypes       valueobj.NotifyTypes `gorm:"column:notify_types;type:json;not null;comment:通知方式"`
	MemberId          uint32               `gorm:"column:member_id;type:int unsigned;not null;comment:通知人员ID"`
	Member            *SysUser             `gorm:"foreignKey:MemberId;comment:成员"`
	Msg               string               `gorm:"column:msg;type:text;not null;comment:通知的消息"`
	Status            valueobj.Status      `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	PromAlarmNotifyID uint32               `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;comment:通知ID"`
	PromAlarmNotify   *PromAlarmNotify     `gorm:"foreignKey:PromAlarmNotifyID"`
}

func (*PromAlarmBeenNotifyMember) TableName() string {
	return TableNamePromAlarmNotifyMember
}

// GetNotifyTypes .
func (p *PromAlarmBeenNotifyMember) GetNotifyTypes() valueobj.NotifyTypes {
	if p.NotifyTypes == nil {
		return valueobj.NotifyTypes{}
	}
	return p.NotifyTypes
}

// GetMember .
func (p *PromAlarmBeenNotifyMember) GetMember() *SysUser {
	if p.Member == nil {
		return nil
	}
	return p.Member
}

// GetPromAlarmNotify .
func (p *PromAlarmBeenNotifyMember) GetPromAlarmNotify() *PromAlarmNotify {
	if p.PromAlarmNotify == nil {
		return nil
	}
	return p.PromAlarmNotify
}
