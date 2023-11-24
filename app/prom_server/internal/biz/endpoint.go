package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
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
func (b *EndpointBiz) AppendEndpoint(ctx context.Context, endpoints []*dobo.EndpointBO) error {
	return b.endpointRepo.Append(ctx, dobo.NewEndpointBO(endpoints...).DO().List())
}

// DeleteEndpoint 删除
func (b *EndpointBiz) DeleteEndpoint(ctx context.Context, endpoints []*dobo.EndpointBO) error {
	return b.endpointRepo.Delete(ctx, dobo.NewEndpointBO(endpoints...).DO().List())
}

// ListEndpoint 查询
func (b *EndpointBiz) ListEndpoint(ctx context.Context) ([]*dobo.EndpointBO, error) {
	list, err := b.endpointRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	return dobo.NewEndpointDO(list...).BO().List(), nil
}
