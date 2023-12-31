package strategyscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
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
	return db.Preload("Endpoint")
}

// PreloadAlarmPages 预加载alarm_pages
func PreloadAlarmPages(db *gorm.DB) *gorm.DB {
	return db.Preload("AlarmPages")
}

// PreloadCategories 预加载categories
func PreloadCategories(db *gorm.DB) *gorm.DB {
	return db.Preload("Categories")
}

// PreloadAlertLevel 预加载alert_level
func PreloadAlertLevel(db *gorm.DB) *gorm.DB {
	return db.Preload("AlertLevel")
}

// PreloadPromNotifies 预加载prom_notifies
func PreloadPromNotifies(db *gorm.DB) *gorm.DB {
	return db.Preload("PromNotifies")
}

// PreloadPromNotifyUpgrade 预加载prom_notify_upgrade
func PreloadPromNotifyUpgrade(db *gorm.DB) *gorm.DB {
	return db.Preload("PromNotifyUpgrade")
}

// PreloadGroupInfo 预加载group_info
func PreloadGroupInfo(db *gorm.DB) *gorm.DB {
	return db.Preload("GroupInfo")
}
