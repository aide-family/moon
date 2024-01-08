package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	EndpointBiz struct {
		log *log.Helper

		endpointRepo repository.EndpointRepo
	}
)

func NewEndpointBiz(endpointRepo repository.EndpointRepo, logger log.Logger) *EndpointBiz {
	return &EndpointBiz{
		log:          log.NewHelper(log.With(logger, "module", "biz.Endpoint")),
		endpointRepo: endpointRepo,
	}
}

// AppendEndpoint 新增
func (b *EndpointBiz) AppendEndpoint(ctx context.Context, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	return b.endpointRepo.Append(ctx, endpoint)
}

// UpdateEndpointById 更新
func (b *EndpointBiz) UpdateEndpointById(ctx context.Context, id uint32, endpoint *bo.EndpointBO) (*bo.EndpointBO, error) {
	updateInfo := endpoint
	updateInfo.Id = id
	return b.endpointRepo.Update(ctx, updateInfo)
}

// UpdateStatusByIds 批量更新状态
func (b *EndpointBiz) UpdateStatusByIds(ctx context.Context, ids []uint32, status vo.Status) error {
	if len(ids) == 0 {
		return nil
	}
	return b.endpointRepo.UpdateStatus(ctx, ids, status)
}

// DetailById 查询详情
func (b *EndpointBiz) DetailById(ctx context.Context, id uint32) (*bo.EndpointBO, error) {
	return b.endpointRepo.Get(ctx, basescopes.InIds(id))
}

// DeleteEndpointById 删除
func (b *EndpointBiz) DeleteEndpointById(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return b.endpointRepo.Delete(ctx, ids)
}

type ListEndpointParams struct {
	Keyword string `form:"keyword"`
	Curr    int32  `form:"curr"`
	Size    int32  `form:"size"`
}

// ListEndpoint 查询
func (b *EndpointBiz) ListEndpoint(ctx context.Context, params *ListEndpointParams) ([]*bo.EndpointBO, basescopes.Pagination, error) {
	pageInfo := basescopes.NewPage(params.Curr, params.Size)
	wheres := []basescopes.ScopeMethod{basescopes.NameLike(params.Keyword)}

	list, err := b.endpointRepo.List(ctx, pageInfo, wheres...)
	if err != nil {
		return nil, nil, err
	}
	return list, pageInfo, nil
}
