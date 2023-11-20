package strategygroup

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// Like keyword
func Like(keyword string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("name LIKE?", keyword+"%")
	}
}

// InIds id列表
func InIds(ids []uint) query.ScopeMethod {
	return query.WhereInColumn("id", ids)
}
