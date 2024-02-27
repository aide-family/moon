package basescopes

import (
	"gorm.io/gorm"
)

const (
	ApiAssociationReplaceRoles = "Roles"
)

const (
	SysApiFiledName = "name"
	SysApiFiledPath = "path"
)

// PreloadRoles 预加载角色
func PreloadRoles(db *gorm.DB) *gorm.DB {
	return db.Preload(ApiAssociationReplaceRoles)
}

func LikeSysApi(keyword string) ScopeMethod {
	return WhereLikePrefixKeyword(keyword, SysApiFiledName, SysApiFiledPath)
}
