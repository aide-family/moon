package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromChatGroup = "prom_alarm_chat_groups"

const (
	PromAlarmChatGroupFieldStatus    = "status"
	PromAlarmChatGroupFieldRemark    = "remark"
	PromAlarmChatGroupFieldName      = "name"
	PromAlarmChatGroupFieldHook      = "hook"
	PromAlarmChatGroupFieldNotifyApp = "notify_app"
	PromAlarmChatGroupFieldHookName  = "hook_name"
	PromAlarmChatGroupFieldTemplate  = "template"
	PromAlarmChatGroupFieldTitle     = "title"
	PromAlarmChatGroupFieldSecret    = "secret"
	PromAlarmChatGroupFieldCreateBy  = "create_by"
)

// PromAlarmChatGroup 告警通知群组机器人信息
type PromAlarmChatGroup struct {
	BaseModel
	Status    vo.Status    `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark    string       `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Name      string       `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__cg__name,priority:1;comment:名称"`
	Hook      string       `gorm:"column:hook;type:varchar(255);not null;comment:钩子地址"`
	NotifyApp vo.NotifyApp `gorm:"column:notify_app;type:tinyint;not null;default:1;comment:通知方式"`
	HookName  string       `gorm:"column:hook_name;type:varchar(64);not null;comment:钩子名称"`
	// 消息模板
	Template string `gorm:"column:template;type:text;not null;comment:消息模板"`
	Title    string `gorm:"column:title;type:varchar(64);not null;comment:消息标题"`
	Secret   string `gorm:"column:secret;type:varchar(128);not null;comment:通信密钥"`
	// 创建人ID
	CreateBy uint32 `gorm:"column:create_by;type:int;not null;comment:创建人ID"`
}

func (*PromAlarmChatGroup) TableName() string {
	return TableNamePromChatGroup
}
