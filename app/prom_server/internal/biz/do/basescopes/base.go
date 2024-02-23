package basescopes

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/util/slices"
)

type Field string

const (
	BaseFieldID        Field = "id"
	BaseFieldCreatedAt Field = "created_at"
	BaseFieldUpdatedAt Field = "updated_at"
	BaseFieldDeletedAt Field = "deleted_at"
	BaseFieldStatus    Field = "status"
	BaseFieldName      Field = "name"
	BaseFieldCreateBy  Field = "create_by"
	BaseFieldTitle     Field = "title"
	BaseFieldUserId    Field = "user_id"
)

// String string
func (f Field) String() string {
	return string(f)
}

// Format string
func (f Field) Format(str ...string) Field {
	return Field(fmt.Sprintf("`%s` %s", f, strings.Join(str, " ")))
}

// InIds id列表
func InIds(ids ...uint32) ScopeMethod {
	newIds := slices.Filter(ids, func(id uint32) bool { return id > 0 })
	return WhereInColumn(BaseFieldID, newIds...)
}

// NameIn 名称列表
func NameIn(names ...string) ScopeMethod {
	newNames := slices.Filter(names, func(name string) bool { return name != "" })
	return WhereInColumn(BaseFieldName, newNames...)
}

// NotInIds id列表
func NotInIds(ids ...uint32) ScopeMethod {
	if len(ids) == 0 {
		return func(db *gorm.DB) *gorm.DB {
			return db
		}
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Not("id", ids)
	}
}

// IdGT idGT
func IdGT(id uint32) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(BaseFieldID.Format(">", "?").String(), id)
	}
}

// StatusEQ 状态
func StatusEQ(status vo.Status) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == vo.StatusUnknown {
			return db
		}
		return db.Where(BaseFieldStatus.String(), status)
	}
}

// StatusNotEQ 状态
func StatusNotEQ(status vo.Status) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if status == vo.StatusUnknown {
			return db
		}
		return db.Where(BaseFieldStatus.Format("!=", "?").String(), status)
	}
}

// NameLike 名称
func NameLike(name string) ScopeMethod {
	return WhereLikePrefixKeyword(name, BaseFieldName)
}

// TitleLike 标题
func TitleLike(title string) ScopeMethod {
	return WhereLikePrefixKeyword(title, BaseFieldTitle)
}

// NameEQ 名称相等
func NameEQ(name string) ScopeMethod {
	return WhereInColumn(BaseFieldName, name)
}

// WithTrashed 查询已删除的记录
func WithTrashed(isDelete bool) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if isDelete {
			return db.Unscoped()
		}
		return db
	}
}

// CreatedAtDesc 按创建时间倒序
func CreatedAtDesc() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(BaseFieldCreatedAt.Format(DESC).String())
	}
}

// UpdateAtDesc 按更新时间倒序
func UpdateAtDesc() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(BaseFieldUpdatedAt.Format(DESC).String())
	}
}

// DeletedAtDesc 按删除时间倒序
func DeletedAtDesc() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Order(BaseFieldDeletedAt.Format(DESC).String())
	}
}

// DeleteAtGT0 删除时间大于0
func DeleteAtGT0() ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(BaseFieldDeletedAt.Format(">= 0").String())
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

// WhereLikePrefixKeyword 前缀模糊查询
func WhereLikePrefixKeyword(keyword string, columns ...Field) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" || len(columns) == 0 {
			return db
		}

		dbTmp := db
		for _, column := range columns {
			dbTmp = dbTmp.Or(column.Format(LIKE, "?").String(), keyword+"%")
		}
		return db.Where(dbTmp)
	}
}

// WhereLikeSuffixKeyword 后缀模糊查询
func WhereLikeSuffixKeyword(keyword string, columns ...Field) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" || len(columns) == 0 {
			return db
		}

		dbTmp := db
		for _, column := range columns {
			dbTmp = dbTmp.Or(column.Format(LIKE, "?").String(), "%"+keyword)
		}
		return db.Where(dbTmp)
	}
}

// WhereLikeKeyword 前后缀模糊查询
func WhereLikeKeyword(keyword string, columns ...Field) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" || len(columns) == 0 {
			return db
		}

		dbTmp := db
		for _, column := range columns {
			dbTmp = dbTmp.Or(column.Format(LIKE, "?").String(), "%"+keyword+"%")
		}
		return db.Where(dbTmp)
	}
}

// BetweenColumn 通过字段名和值列表进行查询
func BetweenColumn(column Field, min, max any) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(column.Format(BETWEEN, "?", AND, "?").String(), min, max)
	}
}

// WithCreateBy 创建人查询
func WithCreateBy(ctx context.Context) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		roleId := middler.GetRoleId(ctx)
		if roleId == "1" || roleId == "" || roleId == "0" {
			return db
		}
		userId := middler.GetUserId(ctx)
		return db.Where(BaseFieldCreateBy.String(), userId)
	}
}

// WithUserId 创建人查询
func WithUserId(ctx context.Context) ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		roleId := middler.GetRoleId(ctx)
		if roleId == "1" || roleId == "" || roleId == "0" {
			return db
		}
		userId := middler.GetUserId(ctx)
		return db.Where(BaseFieldUserId.String(), userId)
	}
}
