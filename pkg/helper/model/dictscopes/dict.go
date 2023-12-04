package dictscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// WhereCategory 根据字典类型查询
func WhereCategory(category int32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if category == 0 {
			return db
		}
		return db.Where("category =?", category)
	}
}

// LikeName 根据字典名称模糊查询
func LikeName(keyword string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("name LIKE ?", keyword+"%")
	}
}

// WithTrashed 查询已删除的记录
func WithTrashed(isDelete bool) query.ScopeMethod {
	if isDelete {
		return query.WithTrashed
	}

	return func(db *gorm.DB) *gorm.DB {
		return db
	}
}
