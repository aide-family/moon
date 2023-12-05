package notifyscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// NotifyInIds id列表条件
func NotifyInIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// NotifyNotInIds id列表条件
func NotifyNotInIds(ids ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(ids) > 0 {
			return db.Not(query.WhereInColumn("id", ids...))
		}
		return db
	}
}

// NotifyLike 模糊查询
func NotifyLike(keyword string) query.ScopeMethod {
	return query.WhereLikeKeyword(keyword+"%", "name")
}

// NotifyEqName 等于name
func NotifyEqName(name string) query.ScopeMethod {
	return query.WhereInColumn("name", name)
}

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
