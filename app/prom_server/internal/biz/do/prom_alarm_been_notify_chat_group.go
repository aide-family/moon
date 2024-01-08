package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromAlarmBeenNotifyChatGroup = "prom_alarm_been_notify_chat_groups"

type PromAlarmBeenNotifyChatGroup struct {
	BaseModel
	RealtimeAlarmID   uint32              `gorm:"column:realtime_alarm_id;type:int unsigned;not null;index:idx__realtime_alarm_id,priority:1;comment:告警ID"`
	ChatGroup         *PromAlarmChatGroup `gorm:"foreignKey:ChatGroupId;comment:通知组"`
	ChatGroupId       uint32              `gorm:"column:chat_group_id;type:int unsigned;not null;index:idx__chat_group_id,priority:1;comment:通知组ID"`
	Status            vo.Status           `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Msg               string              `gorm:"column:msg;type:text;not null;comment:通知的消息"`
	PromAlarmNotifyID uint32              `gorm:"column:prom_alarm_notify_id;type:int unsigned;not null;comment:通知ID"`
}

func (*PromAlarmBeenNotifyChatGroup) TableName() string {
	return TableNamePromAlarmBeenNotifyChatGroup
}

// GetChatGroup .
func (p *PromAlarmBeenNotifyChatGroup) GetChatGroup() *PromAlarmChatGroup {
	if p.ChatGroup == nil {
		return nil
	}
	return p.ChatGroup
}
