package systemscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

const (
	RoleAssociationReplaceUsers = "Users"
	RoleAssociationReplaceApis  = "Apis"
)

// RoleInIds id列表
func RoleInIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// RoleLike 模糊查询
func RoleLike(keyword string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("name LIKE ?", keyword+"%")
	}
}

// RoleEqName 等于name
func RoleEqName(name string) query.ScopeMethod {
	return query.WhereInColumn("name", name)
}

// RolePreloadUsers 预加载用户
func RolePreloadUsers(userIds ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) > 0 {
			return db.Preload("Users", query.WhereInColumn("id", userIds...))
		}
		return db.Preload("Users")
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
