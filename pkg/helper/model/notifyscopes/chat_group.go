package notifyscopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
	"prometheus-manager/pkg/helper/valueobj"
)

// ChatGroupInIds 根据群组id列表查询
func ChatGroupInIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// ChatGroupLike 根据群组名称模糊查询
func ChatGroupLike(keyword string) query.ScopeMethod {
	if keyword == "" {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return query.WhereLikeKeyword(keyword+"%", "name")
}

// ChatGroupStatusEq 根据群组状态查询
func ChatGroupStatusEq(status valueobj.Status) query.ScopeMethod {
	if status.IsUnknown() {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return query.WhereInColumn("status", status)
}
