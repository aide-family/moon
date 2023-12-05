package systemscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

const (
	ApiAssociationReplaceRoles = "Roles"
)

// ApiInIds id列表
func ApiInIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// ApiLike 模糊查询
func ApiLike(keyword string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("name LIKE ?", keyword+"%")
	}
}
