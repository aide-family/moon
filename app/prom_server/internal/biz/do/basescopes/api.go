package basescopes

import (
	"gorm.io/gorm"
)

const (
	ApiAssociationReplaceRoles = "Roles"
)

// PreloadRoles 预加载角色
func PreloadRoles(db *gorm.DB) *gorm.DB {
	return db.Preload(ApiAssociationReplaceRoles)
}
