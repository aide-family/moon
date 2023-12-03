package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromAlarmBeenNotifyChatGroup = "prom_alarm_been_notify_chat_groups"

type PromAlarmBeenNotifyChatGroup struct {
	query.BaseModel
	RealtimeAlarmID   uint                `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__realtime_alarm_id,priority:1;comment:告警ID"`
	ChatGroup         *PromAlarmChatGroup `gorm:"foreignKey:ChatGroupId;comment:通知组"`
	ChatGroupId       uint                `gorm:"column:chat_group_id;type:int unsigned;not null;index:idx__chat_group_id,priority:1;comment:通知组ID"`
	Status            valueobj.Status     `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Msg               string              `gorm:"column:msg;type:text;not null;comment:通知的消息"`
	PromAlarmNotifyID uint                `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;comment:通知ID"`
}

func (*PromAlarmBeenNotifyChatGroup) TableName() string {
	return TableNamePromAlarmBeenNotifyChatGroup
}
