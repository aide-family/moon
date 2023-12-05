package model

import (
	query "github.com/aide-cloud/gorm-normalize"

	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromNotify = "prom_alarm_notifies"

// PromAlarmNotify 告警通知对象
type PromAlarmNotify struct {
	query.BaseModel
	Name            string                   `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:通知名称"`
	Status          valueobj.Status          `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark          string                   `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	ChatGroups      []*PromAlarmChatGroup    `gorm:"many2many:prom_notify_chat_groups;comment:通知组"`
	BeNotifyMembers []*PromAlarmNotifyMember `gorm:"comment:被通知成员"`
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
