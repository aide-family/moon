package do

import (
	"strings"

	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

const TableNamePromGroup = "prom_strategy_groups"

const (
	PromGroupFieldName                  = "name"
	PromGroupFieldStatus                = "status"
	PromGroupFieldRemark                = "remark"
	PromGroupFieldStrategyCount         = "strategy_count"
	PromGroupFieldEnableStrategyCount   = "enable_strategy_count"
	PromGroupPreloadFieldPromStrategies = "PromStrategies"
	PromGroupPreloadFieldCategories     = "Categories"
)

// StrategyGroupPreloadCategories 预加载策略组下的分类
func StrategyGroupPreloadCategories(preload bool) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if !preload {
			return db
		}
		return db.Preload(PromGroupPreloadFieldCategories)
	}
}

// StrategyGroupPreloadPromStrategies 预加载策略组下的策略
func StrategyGroupPreloadPromStrategies(childPreload ...string) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if len(childPreload) == 0 {
			return db.Preload(PromGroupPreloadFieldPromStrategies)
		}
		for _, preload := range childPreload {
			db = db.Preload(strings.Join([]string{PromGroupPreloadFieldPromStrategies, preload}, "."))
		}
		return db
	}
}

// PromStrategyGroup 策略组
type PromStrategyGroup struct {
	BaseModel
	Name                string          `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:规则组名称"`
	StrategyCount       int64           `gorm:"column:strategy_count;type:bigint;not null;default:0;comment:规则数量"`
	EnableStrategyCount int64           `gorm:"column:enable_strategy_count;type:bigint;not null;default:0;comment:启用策略数量"`
	Status              vo.Status       `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用"`
	Remark              string          `gorm:"column:remark;type:varchar(255);not null;comment:描述信息"`
	PromStrategies      []*PromStrategy `gorm:"foreignKey:GroupID"`
	Categories          []*SysDict      `gorm:"many2many:prom_group_categories"`
}

// TableName PromStrategyGroup's table name
func (*PromStrategyGroup) TableName() string {
	return TableNamePromGroup
}

// GetPromStrategies 获取策略组下的策略
func (p *PromStrategyGroup) GetPromStrategies() []*PromStrategy {
	if p == nil {
		return nil
	}
	return p.PromStrategies
}

// GetCategories 获取策略组的分类
func (p *PromStrategyGroup) GetCategories() []*SysDict {
	if p == nil {
		return nil
	}
	return p.Categories
}
