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
		ID:          b.GetId(),
	}
}

// MQDatasourceAPIBuilder MQ数据源api构建器
type MQDatasourceAPIBuilder struct {
	list []*api.MQDatasource
}

// NewMQDatasourceAPIBuilder 创建MQ数据源api构建器
func NewMQDatasourceAPIBuilder(datasource ...*api.MQDatasource) *MQDatasourceAPIBuilder {
	return &MQDatasourceAPIBuilder{
		list: datasource,
	}
}

// ToBo 转换为业务对象
func (b *MQDatasourceAPIBuilder) ToBo() *bo.MQDatasource {
	if types.IsNil(b) || len(b.list) == 0 {
		return nil
	}

	item := b.list[0]
	return &bo.MQDatasource{
		TeamID: item.GetTeamID(),
		ID:     item.GetId(),
		Status: vobj.Status(item.GetStatus()),
		Conf:   item.GetMq(),
	}
}

// ToBos 转换为业务对象数组
func (b *MQDatasourceAPIBuilder) ToBos() []*bo.MQDatasource {
	if types.IsNil(b) || len(b.list) == 0 {
		return nil
	}

	return types.SliceTo(b.list, func(item *api.MQDatasource) *bo.MQDatasource {
		return NewMQDatasourceAPIBuilder(item).ToBo()
	})
}
