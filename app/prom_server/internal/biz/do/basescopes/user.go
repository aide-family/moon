package basescopes

import (
	"gorm.io/gorm"
)

const (
	UserAssociationReplaceRoles = "Roles"
)

const (
	UserTableFieldUsername Field = "username"
	UserTableFieldEmail    Field = "email"
	UserTableFieldPhone    Field = "phone"
	UserTableFieldNickname Field = "nickname"
)

// UserLike 模糊查询
func UserLike(keyword string) ScopeMethod {
	return WhereLikePrefixKeyword(keyword, UserTableFieldUsername, UserTableFieldEmail, UserTableFieldPhone, UserTableFieldNickname)
}

// UserEqName 等于name
func UserEqName(name string) ScopeMethod {
	return WhereInColumn(UserTableFieldUsername, name)
}

// UserEqEmail 等于email
func UserEqEmail(email string) ScopeMethod {
	return WhereInColumn(UserTableFieldEmail, email)
}

// UserEqPhone 等于phone
func UserEqPhone(phone string) ScopeMethod {
	return WhereInColumn(UserTableFieldPhone, phone)
}

// UserPreloadRoles 预加载角色
func UserPreloadRoles(roleIds ...uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(roleIds) > 0 {
			return db.Preload(UserAssociationReplaceRoles, WhereInColumn(BaseFieldID, roleIds...))
		}
		return db.Preload(UserAssociationReplaceRoles)
	}
}
