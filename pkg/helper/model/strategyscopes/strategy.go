package strategyscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// LikeStrategy 策略
func LikeStrategy(keyword string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("name LIKE?", keyword+"%")
	}
}

// StatusEQ 状态
func StatusEQ(status int32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == 0 {
			return db
		}
		return db.Where("status = ?", status)
	}
}

// GroupIdsEQ 策略组ID
func GroupIdsEQ(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("group_id", ids)
}

// InIds id列表
func InIds(ids []uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids)
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
