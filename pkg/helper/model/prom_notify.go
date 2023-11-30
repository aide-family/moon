package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromNotify = "prom_notifies"

type PromNotify struct {
	query.BaseModel
	Name            string              `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_name,priority:1;comment:通知名称" json:"name"`
	Status          int32               `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`
	Remark          string              `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	ChatGroups      []*PromChatGroup    `gorm:"many2many:prom_notify_chat_groups;comment:通知组" json:"chatGroups"`
	BeNotifyMembers []*PromNotifyMember `gorm:"hasMany:PromNotifyMember;comment:被通知成员" json:"beNotifyMembers"`
}

func (*PromNotify) TableName() string {
	return TableNamePromNotify
}
