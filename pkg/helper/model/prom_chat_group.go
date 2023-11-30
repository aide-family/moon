package model

import (
	query "github.com/aide-cloud/gorm-normalize"
)

const TableNamePromChatGroup = "prom_chat_groups"

type PromChatGroup struct {
	query.BaseModel
	Status    int32  `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`
	Remark    string `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Name      string `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx_name,priority:1;comment:名称" json:"name"`
	Hook      string `gorm:"column:hook;type:varchar(255);not null;comment:钩子地址" json:"hook"`
	NotifyApp int32  `gorm:"column:notify_app;type:tinyint;not null;default:1;comment:通知方式" json:"notifyApp"`
	HookName  string `gorm:"column:hook_name;type:varchar(64);not null;comment:钩子名称" json:"hookName"`
}

func (*PromChatGroup) TableName() string {
	return TableNamePromChatGroup
}
