package basescopes

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/vo"
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
func StatusEQ(status vo.Status) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == vo.StatusUnknown {
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

// UpdateAtDesc 按更新时间倒序
func UpdateAtDesc() query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order("updated_at desc")
	}
}

// Page 分页
func Page(pgInfo query.Pagination) query.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if pgInfo == nil {
			return db
		}
		return db.Offset(int((pgInfo.GetCurr() - 1) * pgInfo.GetSize())).Limit(int(pgInfo.GetSize()))
	}
}

type TxContext struct{}

// WithTx 上下文中设置tx
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, TxContext{}, tx)
}

// GetTx 从上下文中获取tx
func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(TxContext{}).(*gorm.DB)
	if ok {
		return tx
	}
	return db
}
