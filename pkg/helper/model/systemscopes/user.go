package systemscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

const (
	UserAssociationReplaceRoles query.AssociationKey = "Roles"
)

// UserInIds id列表
func UserInIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// UserLike 模糊查询
func UserLike(keyword string) query.ScopeMethod {
	if keyword == "" {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return query.WhereLikeKeyword(keyword+"%", "username", "nickname", "email", "phone")
}

// UserEqName 等于name
func UserEqName(name string) query.ScopeMethod {
	return query.WhereInColumn("username", name)
}

// UserEqEmail 等于email
func UserEqEmail(email string) query.ScopeMethod {
	return query.WhereInColumn("email", email)
}

// UserEqPhone 等于phone
func UserEqPhone(phone string) query.ScopeMethod {
	return query.WhereInColumn("phone", phone)
}

// UserPreloadRoles 预加载角色
func UserPreloadRoles(roleIds ...uint32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(roleIds) > 0 {
			return db.Preload("Roles", query.WhereInColumn("id", roleIds...))
		}
		return db.Preload("Roles")
	}
}

// CreatedAtDesc 按创建时间倒序
func CreatedAtDesc() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc")
	}
}
