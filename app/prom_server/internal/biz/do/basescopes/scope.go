package basescopes

import (
	"gorm.io/gorm"
)

type ScopeMethod = func(db *gorm.DB) *gorm.DB

// WhereInColumn 通过字段名和值列表进行查询
func WhereInColumn[T any](column Field, values ...T) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		idsLen := len(values)
		switch idsLen {
		case 0:
			return db
		case 1:
			return db.Where(column.Format("=", "?").String(), values[0])
		default:
			return db.Where(column.Format("IN", "(?)").String(), values)
		}
	}
}
