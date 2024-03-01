package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNameSysLog = "sys_logs"

type SysLog struct {
	BaseModel
	ModuleName vo.Module `gorm:"column:module;type:int;not null;default:0;comment:模块;index:syslog__m__idx"`
	ModuleId   uint32    `gorm:"column:module_id;type:int;not null;default:0;comment:模块id;index:syslog__m__idx"`
	Title      string    `gorm:"column:title;type:varchar(255);not null;comment:日志标题"`
	Content    string    `gorm:"column:content;type:varchar(255);not null;comment:日志内容"`
	UserId     uint32    `gorm:"column:user_id;type:int;not null;default:0;comment:用户id"`
	User       *SysUser  `gorm:"foreignKey:UserId;references:ID;comment:用户"`
	Action     vo.Action `gorm:"column:action;type:int;not null;default:0;comment:操作"`
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
