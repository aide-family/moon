package build

import (
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

type DatasourceBuilder struct {
	*bizmodel.Datasource
}

func NewDatasourceBuilder(datasource *bizmodel.Datasource) *DatasourceBuilder {
	return &DatasourceBuilder{
		Datasource: datasource,
	}
}

func (b *DatasourceBuilder) ToApi() *admin.Datasource {
	if types.IsNil(b) || types.IsNil(b.Datasource) {
		return nil
	}
	configMap := make(map[string]string)
	if err := json.Unmarshal([]byte(b.Config), &configMap); !types.IsNil(err) {
		log.Warnw("error", err)
	}
	return &admin.Datasource{
		Id:          b.ID,
		Name:        b.Name,
		Type:        api.DatasourceType(b.Category),
		Endpoint:    b.Endpoint,
		Status:      api.Status(b.Status),
		CreatedAt:   b.CreatedAt.String(),
		UpdatedAt:   b.UpdatedAt.String(),
		Config:      configMap,
		Remark:      b.Remark,
		StorageType: api.StorageType(b.StorageType),
	}
}

type DatasourceQueryDataBuilder struct {
	*bo.DatasourceQueryData
}

func NewDatasourceQueryDataBuilder(data *bo.DatasourceQueryData) *DatasourceQueryDataBuilder {
	return &DatasourceQueryDataBuilder{
		DatasourceQueryData: data,
	}
}

// ToApi 转换为api
func (b *DatasourceQueryDataBuilder) ToApi() *api.MetricQueryResult {
	if types.IsNil(b) || types.IsNil(b.DatasourceQueryData) {
		return nil
	}
	var value *api.MetricQueryValue
	if !types.IsNil(b.Value) {
		value = &api.MetricQueryValue{
			Value:     b.Value.Value,
			Timestamp: b.Value.Timestamp,
		}
	}
	return &api.MetricQueryResult{
		Labels:     b.Labels,
		ResultType: b.ResultType,
		Values: types.SliceTo(b.Values, func(item *bo.DatasourceQueryValue) *api.MetricQueryValue {
			return &api.MetricQueryValue{
				Value:     item.Value,
				Timestamp: item.Timestamp,
			}
		}),
		Value: value,
	}
}
