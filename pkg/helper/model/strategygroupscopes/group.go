package strategygroupscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

const (
	PromStrategyGroupReplaceCategories = "Categories"
)

// PreloadCategories 预加载策略组下的分类
func PreloadCategories() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyGroupReplaceCategories)
	}
}
