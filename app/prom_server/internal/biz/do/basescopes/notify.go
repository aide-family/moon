package basescopes

import (
	"gorm.io/gorm"
)

const (
	NotifyTablePreloadKeyChatGroups      = "ChatGroups"
	NotifyTablePreloadKeyBeNotifyMembers = "BeNotifyMembers"
)

// NotifyPreloadChatGroups 预加载通知组
func NotifyPreloadChatGroups(chatGroupIds ...uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(chatGroupIds) > 0 {
			return db.Preload(NotifyTablePreloadKeyChatGroups, WhereInColumn(BaseFieldID, chatGroupIds...))
		}
		return db.Preload(NotifyTablePreloadKeyChatGroups)
	}
}

// NotifyPreloadBeNotifyMembers 预加载被通知成员
func NotifyPreloadBeNotifyMembers(beNotifyMemberIds ...uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(beNotifyMemberIds) > 0 {
			return db.Preload(NotifyTablePreloadKeyBeNotifyMembers, WhereInColumn(BaseFieldID, beNotifyMemberIds...))
		}
		return db.Preload(NotifyTablePreloadKeyBeNotifyMembers)
	}
}
