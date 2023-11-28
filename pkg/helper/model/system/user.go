package system

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
	"prometheus-manager/pkg/util/types"
)

const (
	UserAssociationReplaceRoles query.AssociationKey = "Roles"
)

// UserInIds id列表
func UserInIds[T types.Int](ids ...T) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// UserLike 模糊查询
func UserLike(keyword string) query.ScopeMethod {
	return query.WhereLikeKeyword(keyword+"%", "username", "nickname", "email", "phone")
}

// UserEqName 等于name
func UserEqName(name string) query.ScopeMethod {
	return query.WhereInColumn("username", name)
}

// UserPreloadRoles 预加载角色
func UserPreloadRoles[T types.Int](roleIds ...T) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(roleIds) > 0 {
			return db.Preload("Roles", query.WhereInColumn("id", roleIds...))
		}
		return db.Preload("Roles")
	}
}
