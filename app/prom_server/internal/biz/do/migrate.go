package do

import (
	_ "embed"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) (err error) {
	err = db.AutoMigrate(
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
		&PromAlarmBeenNotifyMember{},
		&PromAlarmBeenNotifyChatGroup{},
		&Endpoint{},
		&ExternalNotifyObj{},
		&ExternalCustomerHook{},
		&ExternalCustomer{},
		&SysUser{},
		&SysRole{},
		&SysAPI{},
		&CasbinRule{},
		&MyChart{},
		&MyDashboardConfig{},
	)
	if err != nil {
		return err
	}
	return initSysApi(db)
}

//go:embed init.sql
var sql string

// InitSysApi 初始化系统接口权限列表
func initSysApi(db *gorm.DB) (err error) {
	if err = db.Model(&SysAPI{}).Unscoped().Where("id > 0").Delete(&SysAPI{}).Error; err != nil {
		return err
	}
	return db.Exec(sql).Error
}
