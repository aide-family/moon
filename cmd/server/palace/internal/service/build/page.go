package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
)

type PageBuilder struct {
	types.Pagination
}

func NewPageBuilder(page types.Pagination) *PageBuilder {
	return &PageBuilder{
		Pagination: page,
	}
}

// ToApi 转换为api对象
func (b *PageBuilder) ToApi() *api.PaginationReply {
	if types.IsNil(b) || types.IsNil(b.Pagination) {
		return nil
	}
	return &api.PaginationReply{
		PageNum:  int32(b.GetPageNum()),
		PageSize: int32(b.GetPageSize()),
		Total:    int64(b.GetTotal()),
	}
}
