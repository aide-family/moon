package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// DatasourceAPIBuilder 数据源api构建器
type DatasourceAPIBuilder struct {
	*api.Datasource
}

// NewDatasourceAPIBuilder 创建数据源api构建器
func NewDatasourceAPIBuilder(datasource *api.Datasource) *DatasourceAPIBuilder {
	return &DatasourceAPIBuilder{
		Datasource: datasource,
	}
}

// ToBo 转换为业务对象
func (b *DatasourceAPIBuilder) ToBo() *bo.Datasource {
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
