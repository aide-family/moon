package pagescopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
)

// LikePageName 页面名称模糊查询
func LikePageName(keyword string) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		return db.Where("`name` Like ?", keyword+"%")
	}
}

// StatusEQ 页面状态查询
func StatusEQ(status int32) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == 0 {
			return db
		}
		return db.Where("`status` = ?", status)
	}
}
