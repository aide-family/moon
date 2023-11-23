package system

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/util/types"
)

// RoleInIds id列表
func RoleInIds[T types.Int](ids ...T) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// RoleLike 模糊查询
func RoleLike(keyword string) query.ScopeMethod {
	return query.WhereLikeKeyword(keyword+"%", "name")
}
