package alarmscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	RealtimeAssociationReplaceIntervenes = "AlarmIntervenes"
	RealtimeAssociationUpgradeInfo       = "AlarmUpgradeInfo"
	RealtimeAssociationSuppressInfo      = "AlarmSuppressInfo"
	RealtimeAssociationBeenNotifyMembers = "BeenNotifyMembers"
	RealtimeAssociationBeenChatGroups    = "BeenChatGroups"
	RealtimeAssociationLevel             = "Level"
)

// RealtimeLike 查询关键字
func RealtimeLike(keyword string) query.ScopeMethod {
	if keyword == "" {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return query.WhereLikeKeyword(keyword+"%", "note", "instance")
}

// RealtimeEventAtDesc 事件时间倒序
func RealtimeEventAtDesc() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("event_at DESC")
	}
}

// InHistoryIds 查询历史ID列表
func InHistoryIds(historyIds ...uint32) query.ScopeMethod {
	return query.WhereInColumn("history_id", historyIds...)
}

// ClauseOnConflict 冲突处理
func ClauseOnConflict() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}, {Name: "history_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"status"}),
		})
	}
}

// PreloadLevel 预加载级别
func PreloadLevel() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(RealtimeAssociationLevel)
	}
}
