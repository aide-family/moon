package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type DatasourceApiBuilder struct {
	*api.Datasource
}

func NewDatasourceApiBuilder(datasource *api.Datasource) *DatasourceApiBuilder {
	return &DatasourceApiBuilder{
		Datasource: datasource,
	}
}

func (b *DatasourceApiBuilder) ToBo() *bo.Datasource {
	if types.IsNil(b) || types.IsNil(b.Datasource) {
		return nil
	}

	return &bo.Datasource{
		Category:    vobj.DatasourceType(b.GetCategory()),
		StorageType: vobj.StorageType(b.GetStorageType()),
		Config:      b.GetConfig(),
		Endpoint:    b.GetEndpoint(),
	}
}
