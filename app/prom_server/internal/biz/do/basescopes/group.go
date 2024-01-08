package basescopes

import (
	"gorm.io/gorm"
)

const (
	PromStrategyGroupReplaceCategories = "Categories"
)

// PreloadStrategyGroupCategories 预加载策略组下的分类
func PreloadStrategyGroupCategories() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyGroupReplaceCategories)
	}
}
