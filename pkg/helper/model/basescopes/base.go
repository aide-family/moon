package basescopes

import (
	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
	"prometheus-manager/pkg/helper/valueobj"
)

// InIds id列表
func InIds(ids ...uint32) query.ScopeMethod {
	return query.WhereInColumn("id", ids)
}

// NotInIds id列表
func NotInIds(ids ...uint32) query.ScopeMethod {
	if len(ids) == 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Not(InIds(ids...))
	}
}

// StatusEQ 状态
func StatusEQ(status valueobj.Status) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == valueobj.StatusUnknown {
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

// NameEQ 名称相等
func NameEQ(name string) query.ScopeMethod {
	return query.WhereInColumn("name", name)
}

// WithTrashed 查询已删除的记录
func WithTrashed(isDelete bool) query.ScopeMethod {
	if isDelete {
		return query.WithTrashed
	}

	return func(db *gorm.DB) *gorm.DB {
		return db
	}
}

// CreatedAtDesc 按创建时间倒序
func CreatedAtDesc() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at desc")
	}
}
