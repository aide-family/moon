package system

import (
	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/pkg/util/types"
)

// ApiInIds id列表
func ApiInIds[T types.Int](ids ...T) query.ScopeMethod {
	return query.WhereInColumn("id", ids...)
}

// ApiLike 模糊查询
func ApiLike(keyword string) query.ScopeMethod {
	return query.WhereLikeKeyword(keyword+"%", "name")
}
