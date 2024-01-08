package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromNotify = "prom_alarm_notifies"

// PromAlarmNotify 告警通知对象
type PromAlarmNotify struct {
	BaseModel
	Name            string                   `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:通知名称"`
	Status          vo.Status                `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark          string                   `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	ChatGroups      []*PromAlarmChatGroup    `gorm:"many2many:prom_notify_chat_groups;comment:通知组"`
	BeNotifyMembers []*PromAlarmNotifyMember `gorm:"comment:被通知成员"`
	// 外部体系通知对象(不在用户体系内的人和hook), 多对多
	ExternalNotifyObjs []*ExternalNotifyObj `gorm:"many2many:prom_alarm_notify_external_notify_objs;comment:外部体系通知对象"`
}

func (*PromAlarmNotify) TableName() string {
	return TableNamePromNotify
}

// GetChatGroups 获取通知组
func (p *PromAlarmNotify) GetChatGroups() []*PromAlarmChatGroup {
	if p == nil {
		return nil
	}
	return p.ChatGroups
}

// GetBeNotifyMembers 获取被通知成员
func (p *PromAlarmNotify) GetBeNotifyMembers() []*PromAlarmNotifyMember {
	if p == nil {
		return nil
	}
	return p.BeNotifyMembers
}

// GetExternalNotifyObjs 获取外部体系通知对象
func (p *PromAlarmNotify) GetExternalNotifyObjs() []*ExternalNotifyObj {
	if p == nil {
		return nil
	}
	return p.ExternalNotifyObjs
}
