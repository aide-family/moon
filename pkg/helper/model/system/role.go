package system

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
	"prometheus-manager/pkg/util/types"
)

const (
	RoleAssociationReplaceUsers query.AssociationKey = "Users"
)

// RoleInIds id列表
func RoleInIds[T types.Int](ids ...T) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// RoleLike 模糊查询
func RoleLike(keyword string) query.ScopeMethod {
	return query.WhereLikeKeyword(keyword+"%", "name")
}

// RoleEqName 等于name
func RoleEqName(name string) query.ScopeMethod {
	return query.WhereInColumn("name", name)
}

// RolePreloadUsers 预加载用户
func RolePreloadUsers[T types.Int](userIds ...T) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(userIds) > 0 {
			return db.Preload("Users", query.WhereInColumn("id", userIds...))
		}
		return db.Preload("Users")
	}
}
