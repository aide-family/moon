package build

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

func NewBuilder() *builder {
	return &builder{}
}

type (
	builder struct {
		ctx context.Context
	}

	Builder interface {
		WithContext(ctx context.Context) Builder

		// TODO 注册新的数据转换方法写在这里

		WithDoDatasource(d *bizmodel.Datasource) DatasourceBuilder
		WithBoDatasourceQueryData(d *bo.DatasourceQueryData) DatasourceQueryDataBuilder
	}
)

func (b *builder) WithBoDatasourceQueryData(d *bo.DatasourceQueryData) DatasourceQueryDataBuilder {
	return &datasourceQueryDataBuilder{
		DatasourceQueryData: d,
		ctx:                 b.ctx,
	}
}

func (b *builder) WithDoDatasource(d *bizmodel.Datasource) DatasourceBuilder {
	return &datasourceBuilder{
		Datasource: d,
		ctx:        b.ctx,
	}
}

func (b *builder) WithContext(ctx context.Context) Builder {
	b.ctx = ctx
	return b
}
