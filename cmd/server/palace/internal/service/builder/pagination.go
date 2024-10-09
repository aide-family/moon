package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
)

var _ IPaginationModuleBuilder = (*paginationModuleBuilder)(nil)

type (
	IPaginationModuleBuilder interface {
		ToBo(*api.PaginationReq) types.Pagination
		ToAPI(types.Pagination) *api.PaginationReply
	}

	paginationModuleBuilder struct {
		ctx context.Context
	}
)

func (p *paginationModuleBuilder) ToBo(req *api.PaginationReq) types.Pagination {
	return types.NewPagination(req)
}

func (p *paginationModuleBuilder) ToAPI(pagination types.Pagination) *api.PaginationReply {
	if types.IsNil(pagination) {
		return nil
	}
	return &api.PaginationReply{
		PageNum:  pagination.GetPageNum(),
		PageSize: pagination.GetPageSize(),
		Total:    pagination.GetTotal(),
	}
}
