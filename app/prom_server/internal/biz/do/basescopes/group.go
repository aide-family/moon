package basescopes

import (
	"strings"

	"gorm.io/gorm"
)

const (
	PromStrategyGroupReplaceCategories     = "Categories"
	PromStrategyGroupReplacePromStrategies = "PromStrategies"
)

// PreloadStrategyGroupCategories 预加载策略组下的分类
func PreloadStrategyGroupCategories() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(PromStrategyGroupReplaceCategories)
	}
}

// PreloadStrategyGroupPromStrategies 预加载策略组下的策略
func PreloadStrategyGroupPromStrategies(childPreload ...string) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range childPreload {
			db = db.Preload(strings.Join([]string{PromStrategyGroupReplacePromStrategies, preload}, "."))
		}
		return db
	}
}
