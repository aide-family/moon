package basescopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// InIds id列表
func InIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids)
}

// StatusEQ 状态
func StatusEQ(status int32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == 0 {
			return db
		}
		return db.Where("status = ?", status)
	}
}

// NameLike 名称
func NameLike(name string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if name == "" {
			return db
		}
		return db.Where("name like?", name+"%")
	}
}
