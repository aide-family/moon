package model

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/helper/valueobj"
)

const TableNamePromChatGroup = "prom_alarm_chat_groups"

// PromAlarmChatGroup 告警通知群组机器人信息
type PromAlarmChatGroup struct {
	query.BaseModel
	Status    valueobj.Status    `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark    string             `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Name      string             `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_name,priority:1;comment:名称"`
	Hook      string             `gorm:"column:hook;type:varchar(255);not null;comment:钩子地址"`
	NotifyApp valueobj.NotifyApp `gorm:"column:notify_app;type:tinyint;not null;default:1;comment:通知方式"`
	HookName  string             `gorm:"column:hook_name;type:varchar(64);not null;comment:钩子名称"`
}

func (*PromAlarmChatGroup) TableName() string {
	return TableNamePromChatGroup
}
