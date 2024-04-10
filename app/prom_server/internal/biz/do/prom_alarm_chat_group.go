package do

import (
	"gorm.io/gorm"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
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

	PromAlarmChatGroupPreloadFieldCreateUser = "CreateUser"
)

// PromAlarmChatGroupInApp 根据app类型查询
func PromAlarmChatGroupInApp(apps ...vobj.NotifyApp) basescopes.ScopeMethod {
	apps = slices.Filter(apps, func(app vobj.NotifyApp) bool {
		return !app.IsUnknown()
	})
	return basescopes.WhereInColumn(PromAlarmChatGroupFieldNotifyApp, apps...)
}

// PromAlarmChatGroupPreloadCreateBy 预加载创建者
func PromAlarmChatGroupPreloadCreateBy() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromAlarmChatGroupPreloadFieldCreateUser)
	}
}

// PromAlarmChatGroup 告警通知群组机器人信息
type PromAlarmChatGroup struct {
	BaseModel
	Status    vobj.Status    `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark    string         `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Name      string         `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__cg__name,priority:1;comment:名称"`
	Hook      string         `gorm:"column:hook;type:varchar(255);not null;comment:钩子地址"`
	NotifyApp vobj.NotifyApp `gorm:"column:notify_app;type:tinyint;not null;default:1;comment:通知方式"`
	HookName  string         `gorm:"column:hook_name;type:varchar(64);not null;comment:钩子名称"`
	Secret    string         `gorm:"column:secret;type:varchar(128);not null;comment:通信密钥"`
	// 创建人ID
	CreateBy   uint32   `gorm:"column:create_by;type:bigint;UNSIGNED;not null;DEFAULT:0;comment:创建人ID"`
	CreateUser *SysUser `gorm:"foreignKey:CreateBy"`
}

func (*PromAlarmChatGroup) TableName() string {
	return TableNamePromChatGroup
}
