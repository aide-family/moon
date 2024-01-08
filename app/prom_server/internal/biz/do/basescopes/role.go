package basescopes

import (
	"gorm.io/gorm"
)

const (
	RoleAssociationReplaceUsers = "Users"
	RoleAssociationReplaceApis  = "Apis"
)

// RolePreloadUsers 预加载用户
func RolePreloadUsers(userIds ...uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) > 0 {
			return db.Preload(RoleAssociationReplaceUsers, WhereInColumn(BaseFieldID, userIds...))
		}
		return db.Preload(RoleAssociationReplaceUsers)
	}
}

// RolePreloadApis 预加载api
func RolePreloadApis(apiIds ...uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(apiIds) > 0 {
			return db.Preload(RoleAssociationReplaceApis, WhereInColumn(BaseFieldID, apiIds...))
		}
		return db.Preload(RoleAssociationReplaceApis)
	}
}
