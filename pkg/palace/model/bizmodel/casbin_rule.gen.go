package bizmodel

import (
	"encoding/json"
)

const tableNameCasbinRule = "casbin_rule"

// CasbinRule mapped from table <casbin_rule>
type CasbinRule struct {
	ID    uint32 `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Ptype string `gorm:"column:ptype;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:1" json:"ptype"`
	V0    string `gorm:"column:v0;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:2" json:"v0"`
	V1    string `gorm:"column:v1;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:3" json:"v1"`
	V2    string `gorm:"column:v2;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:4" json:"v2"`
	V3    string `gorm:"column:v3;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:5" json:"v3"`
	V4    string `gorm:"column:v4;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:6" json:"v4"`
	V5    string `gorm:"column:v5;type:varchar(100);uniqueIndex:idx_casbin_rule,priority:7" json:"v5"`
}

// String json string
func (c *CasbinRule) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// TableName CasbinRule's table name
func (*CasbinRule) TableName() string {
	return tableNameCasbinRule
}
