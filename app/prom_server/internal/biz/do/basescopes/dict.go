package basescopes

import (
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const (
	DictTableFieldCategory Field = "category"
)

// WhereCategory 根据字典类型查询
func WhereCategory(category vo.Category) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if category.IsUnknown() {
			return db
		}
		return db.Where(DictTableFieldCategory.String(), category)
	}
}
