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
