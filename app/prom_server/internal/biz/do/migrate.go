package do

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
	"prometheus-manager/pkg/util/cache"
	"prometheus-manager/pkg/util/hash"
)

func Migrate(db *gorm.DB, cache cache.GlobalCache) (err error) {
	err = db.AutoMigrate(
		&PromAlarmChatGroup{},
		&PromAlarmHistory{},
		&PromAlarmIntervene{},
		&PromAlarmNotify{},
		&PromAlarmNotifyMember{},
		&PromAlarmRealtime{},
		&PromAlarmSuppress{},
		&SysDict{},
		&PromStrategy{},
		&PromStrategyNotifyTemplate{},
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
		&DataUserOp{},
		&DataRoleOp{},
		&SysLog{},
	)
	if err != nil {
		return err
	}
	return initSysApi(db, cache)
}

//go:embed init.sql
var sql string

//go:embed dict.sql
var dictSql string

const syncSysApiFlag = "sync_sys_api"

// InitSysApi 初始化系统接口权限列表
func initSysApi(db *gorm.DB, cache cache.GlobalCache) (err error) {
	ctx := context.Background()
	sqlHash := hash.MD5(sql)
	flagBytes, _ := cache.Get(ctx, syncSysApiFlag)
	flag := string(flagBytes)
	if flag == sqlHash {
		return nil
	}

	if err = cache.Set(ctx, syncSysApiFlag, []byte(sqlHash), 0); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			cache.Del(ctx, syncSysApiFlag)
		}
	}()
	err = db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(&SysAPI{}).Unscoped().Where("id > 0").Delete(&SysAPI{}).Error; err != nil {
			return err
		}
		if err = tx.Model(&SysDict{}).Unscoped().Where("id > 0").Delete(&SysDict{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err = db.Exec(sql).Error; err != nil {
		return err
	}
	if err = db.Exec(dictSql).Error; err != nil {
		return err
	}

	return nil
}
