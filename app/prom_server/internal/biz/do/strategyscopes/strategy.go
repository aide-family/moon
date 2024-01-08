package strategyscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

const (
	PreloadKeyAlarmPages        = "AlarmPages"
	PreloadKeyCategories        = "Categories"
	PreloadKeyEndpoint          = "Endpoint"
	PreloadKeyAlertLevel        = "AlertLevel"
	PreloadKeyPromNotifies      = "PromNotifies"
	PreloadKeyPromNotifyUpgrade = "PromNotifyUpgrade"
	PreloadKeyGroupInfo         = "GroupInfo"
)

// GroupIdsEQ 策略组ID
func GroupIdsEQ(ids ...uint32) query.ScopeMethod {
	tmpIds := make([]uint32, 0, len(ids))
	for _, id := range ids {
		if id > 0 {
			tmpIds = append(tmpIds, id)
		}
	}
	if len(tmpIds) == 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}

	return query.WhereInColumn("group_id", ids)
}

// AlertLike 策略名称匹配
func AlertLike(keyword string) query.ScopeMethod {
	if keyword == "" {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return query.WhereLikeKeyword(keyword+"%", "alert")
}

// PreloadEndpoint 预加载endpoint
func PreloadEndpoint(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyEndpoint)
}

// PreloadAlarmPages 预加载alarm_pages
func PreloadAlarmPages(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyAlarmPages)
}

// PreloadCategories 预加载categories
func PreloadCategories(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyCategories)
}

// PreloadAlertLevel 预加载alert_level
func PreloadAlertLevel(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyAlertLevel)
}

// PreloadPromNotifies 预加载prom_notifies
func PreloadPromNotifies(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyPromNotifies)
}

// PreloadPromNotifyUpgrade 预加载prom_notify_upgrade
func PreloadPromNotifyUpgrade(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyPromNotifyUpgrade)
}

// PreloadGroupInfo 预加载group_info
func PreloadGroupInfo(db *gorm.DB) *gorm.DB {
	return db.Preload(PreloadKeyGroupInfo)
}
