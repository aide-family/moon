package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/util/types"
)

var _ IPaginationModuleBuilder = (*paginationModuleBuilder)(nil)

type (
	// IPaginationModuleBuilder 分页模块构造器
	IPaginationModuleBuilder interface {
		// ToBo 转换为业务对象
		ToBo(*api.PaginationReq) types.Pagination
		// ToAPI 转换为API对象
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
