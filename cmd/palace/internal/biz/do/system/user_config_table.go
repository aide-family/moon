package system

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

var _ do.UserConfigTable = (*UserConfigTable)(nil)

const tableNameUserConfigTable = "sys_user_config_tables"

type UserConfigTable struct {
	do.CreatorModel
	TableKey string   `gorm:"column:table_key;type:varchar(64);not null;comment:table name" json:"tableKey"`
	PageSize int      `gorm:"column:page_size;type:int(10);not null;comment:items per page" json:"pageSize"`
	Columns  []string `gorm:"column:columns;type:text;not null;comment:columns" json:"columns"`
}

func (u *UserConfigTable) GetTableKey() string {
	if u == nil {
		return ""
	}
	return u.TableKey
}

func (u *UserConfigTable) GetPageSize() int {
	if u == nil {
		return 0
	}
	return u.PageSize
}

func (u *UserConfigTable) GetColumns() []string {
	if u == nil {
		return nil
	}
	return u.Columns
}

func (u *UserConfigTable) TableName() string {
	return tableNameUserConfigTable
}
