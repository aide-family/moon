package model

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&PromAlarmChatGroup{},
		&PromAlarmHistory{},
		&PromAlarmIntervene{},
		&PromAlarmNotify{},
		&PromAlarmNotifyMember{},
		&PromAlarmPage{},
		&PromAlarmRealtime{},
		&PromAlarmSuppress{},
		&PromDict{},
		&PromStrategy{},
		&PromStrategyGroup{},

		&SysUser{},
		&SysRole{},
		&SysApi{},
	)
}
