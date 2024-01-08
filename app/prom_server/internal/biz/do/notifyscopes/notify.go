package notifyscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// NotifyPreloadChatGroups 预加载通知组
func NotifyPreloadChatGroups(chatGroupIds ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(chatGroupIds) > 0 {
			return db.Preload("ChatGroups", query.WhereInColumn("id", chatGroupIds...))
		}
		return db.Preload("ChatGroups")
	}
}

// NotifyPreloadBeNotifyMembers 预加载被通知成员
func NotifyPreloadBeNotifyMembers(beNotifyMemberIds ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(beNotifyMemberIds) > 0 {
			return db.Preload("BeNotifyMembers", query.WhereInColumn("id", beNotifyMemberIds...))
		}
		return db.Preload("BeNotifyMembers")
	}
}
