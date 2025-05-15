package system

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

var _ do.UserConfigTable = (*UserConfigTable)(nil)

const tableNameUserConfigTable = "sys_user_config_tables"

type UserConfigTable struct {
	do.CreatorModel
	TableKey string   `gorm:"column:table_key;type:varchar(64);not null;comment:表名" json:"tableKey"`
	PageSize int      `gorm:"column:page_size;type:int(10);not null;comment:每页条数" json:"pageSize"`
	Columns  []string `gorm:"column:columns;type:text;not null;comment:列" json:"columns"`
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
