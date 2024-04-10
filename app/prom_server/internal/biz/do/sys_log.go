package do

import (
	"gorm.io/gorm"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

const TableNameSysLog = "sys_logs"

const (
	SysLogFieldModule      = "module"
	SysLogFieldModuleId    = "module_id"
	SysLogFieldTitle       = "title"
	SysLogFieldContent     = "content"
	SysLogFieldUserId      = "user_id"
	SysLogFieldAction      = "action"
	SysLogPreloadFieldUser = "User"
)

// SysLogPreloadUsers 用户
func SysLogPreloadUsers(userIds ...uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) == 0 {
			return db.Preload(SysLogPreloadFieldUser)
		}
		return db.Preload(SysLogPreloadFieldUser, basescopes.WhereInColumn(basescopes.BaseFieldID, userIds...))
	}
}

// SysLogWhereModule .
func SysLogWhereModule(moduleName vobj.Module, moduleId uint32) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(SysLogFieldModule, moduleName).
			Where(SysLogFieldModuleId, moduleId)
	}
}

type SysLog struct {
	BaseModel
	Module   vobj.Module `gorm:"column:module;type:int;not null;default:0;comment:模块;index:syslog__m__idx"`
	ModuleId uint32      `gorm:"column:module_id;type:int;not null;default:0;comment:模块id;index:syslog__m__idx"`
	Title    string      `gorm:"column:title;type:varchar(255);not null;comment:日志标题"`
	Content  string      `gorm:"column:content;type:text;not null;comment:日志内容"`
	UserId   uint32      `gorm:"column:user_id;type:int;not null;default:0;comment:用户id"`
	Action   vobj.Action `gorm:"column:action;type:int;not null;default:0;comment:操作"`

	User *SysUser `gorm:"foreignKey:UserId;references:ID;comment:用户"`
}

func (l *SysLog) TableName() string {
	return TableNameSysLog
}

// GetUser .
func (l *SysLog) GetUser() *SysUser {
	if l == nil {
		return nil
	}
	return l.User
}
