package system

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/util/types"
)

// UserInIds id列表
func UserInIds[T types.Int](ids ...T) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// UserLike 模糊查询
func UserLike(keyword string) query.ScopeMethod {
	return query.WhereLikeKeyword(keyword+"%", "name")
}

// UserEqName 等于name
func UserEqName(name string) query.ScopeMethod {
	return query.WhereInColumn("name", name)
}
