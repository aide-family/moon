package systemscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

const (
	RoleAssociationReplaceUsers = "Users"
	RoleAssociationReplaceApis  = "Apis"
)

// RolePreloadUsers 预加载用户
func RolePreloadUsers(userIds ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) > 0 {
			return db.Preload(RoleAssociationReplaceUsers, query.WhereInColumn("id", userIds...))
		}
		return db.Preload(RoleAssociationReplaceUsers)
	}
}

// RolePreloadApis 预加载api
func RolePreloadApis(apiIds ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(apiIds) > 0 {
			return db.Preload(RoleAssociationReplaceApis, query.WhereInColumn("id", apiIds...))
		}
		return db.Preload(RoleAssociationReplaceApis)
	}
}
