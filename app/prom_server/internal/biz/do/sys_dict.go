package do

import (
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/field"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"gorm.io/gorm"
)

const TableNameSysDict = "prom_dict"

type SysDictField string
type SysDictWithField string

const (
	SysDictFieldID        field.NumberField = "id"
	SysDictFieldCreatedAt field.TimeField   = "created_at"
	SysDictFieldUpdatedAt field.TimeField   = "updated_at"
	SysDictFieldName      field.StringField = "name"
	SysDictFieldStatus    field.NumberField = "status"
	SysDictFieldRemark    field.StringField = "remark"
	SysDictFieldColor     field.StringField = "color"
	SysDictFieldCategory  field.NumberField = "category"

	SysDictWithPromStrategies SysDictWithField = "PromStrategies"
)

// SysDictWhereCategory 根据字典类型查询
func SysDictWhereCategory(category vobj.Category) basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if category.IsUnknown() {
			return db
		}
		return db.Where(SysDictFieldCategory, category)
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
	return TableNameSysDict
}

func (l *SysDict) GetName() string {
	if l == nil {
		return ""
	}
	return l.Name
}

func (l *SysDict) GetCategory() vobj.Category {
	if l == nil {
		return vobj.CategoryUnknown
	}
	return l.Category
}

func (l *SysDict) GetColor() string {
	if l == nil {
		return ""
	}
	return l.Color
}

func (l *SysDict) GetStatus() vobj.Status {
	if l == nil {
		return vobj.StatusUnknown
	}
	return l.Status
}

func (l *SysDict) GetRemark() string {
	if l == nil {
		return ""
	}
	return l.Remark
}

func (l *SysDict) GetPromStrategies() []*PromStrategy {
	if l == nil {
		return nil
	}
	return l.PromStrategies
}
