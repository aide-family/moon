package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
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
func (b *EndpointBiz) AppendEndpoint(ctx context.Context, endpoints []*bo.EndpointBO) error {
	return b.endpointRepo.Append(ctx, endpoints)
}

// DeleteEndpoint 删除
func (b *EndpointBiz) DeleteEndpoint(ctx context.Context, endpoints []*bo.EndpointBO) error {
	return b.endpointRepo.Delete(ctx, endpoints)
}

// ListEndpoint 查询
func (b *EndpointBiz) ListEndpoint(ctx context.Context) ([]*bo.EndpointBO, error) {
	list, err := b.endpointRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}
