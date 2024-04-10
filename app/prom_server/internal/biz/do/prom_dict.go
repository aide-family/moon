package do

import (
	"gorm.io/gorm"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
)

const TableNamePromDict = "prom_dict"

const (
	PromDictFieldName     = "name"
	PromDictFieldStatus   = "status"
	PromDictFieldRemark   = "remark"
	PromDictFieldColor    = "color"
	PromDictFieldCategory = "category"

	PromAlarmPageReloadFieldPromStrategies = "PromStrategies"
)

// SysDictWhereCategory 根据字典类型查询
func SysDictWhereCategory(category vobj.Category) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if category.IsUnknown() {
			return db
		}
		return db.Where(PromDictFieldCategory, category)
	}
}

// SysDict 系统的字典管理
type SysDict struct {
	BaseModel
	Name           string          `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__p__name__category,priority:1;comment:字典名称"`
	Category       vobj.Category   `gorm:"column:category;type:tinyint;not null;uniqueIndex:idx__p__name__category,priority:2;index:idx__category,priority:1;comment:字典类型"`
	Color          string          `gorm:"column:color;type:varchar(32);not null;default:#165DFF;comment:字典tag颜色"`
	Status         vobj.Status     `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	Remark         string          `gorm:"column:remark;type:varchar(255);not null;comment:字典备注"`
	PromStrategies []*PromStrategy `gorm:"many2many:prom_strategy_alarm_pages"`
}

// TableName SysDict table name
func (*SysDict) TableName() string {
	return TableNamePromDict
}
