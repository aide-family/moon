package model

const TableNameCasbinRule = "casbin_rule"

type CasbinRule struct {
	ID    uint32 `gorm:"primary_key" json:"id"`
	PType string `gorm:"column:ptype;type:varchar(100);not null;index:idx_p_type,priority:1;comment:权限类型"`
	V0    string `gorm:"column:v0;type:varchar(100);not null;index:idx_v0,priority:1;comment:权限参数0"`
	V1    string `gorm:"column:v1;type:varchar(100);not null;index:idx_v1,priority:1;comment:权限参数1"`
	V2    string `gorm:"column:v2;type:varchar(100);not null;index:idx_v2,priority:1;comment:权限参数2"`
	V3    string `gorm:"column:v3;type:varchar(100);not null;index:idx_v3,priority:1;comment:权限参数3"`
	V4    string `gorm:"column:v4;type:varchar(100);not null;index:idx_v4,priority:1;comment:权限参数4"`
	V5    string `gorm:"column:v5;type:varchar(100);not null;index:idx_v5,priority:1;comment:权限参数5"`
}

func (*CasbinRule) TableName() string {
	return TableNameCasbinRule
}
