package build

import (
	"github.com/aide-family/moon/api"
	types2 "github.com/aide-family/moon/pkg/util/types"
)

type PageBuild struct {
	types2.Pagination
}

func NewPageBuild(page types2.Pagination) *PageBuild {
	return &PageBuild{
		Pagination: page,
	}
}

// ToApi 转换为api对象
func (b *PageBuild) ToApi() *api.PaginationReply {
	if types2.IsNil(b) || types2.IsNil(b.Pagination) {
		return nil
	}
	return &api.PaginationReply{
		PageNum:  int32(b.GetPageNum()),
		PageSize: int32(b.GetPageSize()),
		Total:    int64(b.GetTotal()),
	}
}
