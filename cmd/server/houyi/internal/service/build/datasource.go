package build

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// DatasourceAPIBuilder 数据源api构建器
type DatasourceAPIBuilder struct {
	*api.DatasourceItem
}

// NewDatasourceAPIBuilder 创建数据源api构建器
func NewDatasourceAPIBuilder(datasource *api.DatasourceItem) *DatasourceAPIBuilder {
	return &DatasourceAPIBuilder{
		DatasourceItem: datasource,
	}
}

// ToMetricBo 转换为业务对象
func (b *DatasourceAPIBuilder) ToMetricBo() *bo.Datasource {
	if types.IsNil(b) || types.IsNil(b.DatasourceItem) {
		return nil
	}

	return &bo.Datasource{
		Category:    vobj.DatasourceType(b.GetCategory()),
		StorageType: vobj.StorageType(b.GetStorageType()),
		Config:      b.GetConfig(),
		Endpoint:    b.GetEndpoint(),
		ID:          b.GetId(),
		Status:      vobj.Status(b.GetStatus()),
		TeamID:      b.GetTeamId(),
	}
}

// ToEventBo 转换为事件对象
func (b *DatasourceAPIBuilder) ToEventBo() *bo.EventDatasource {
	if types.IsNil(b) || types.IsNil(b.DatasourceItem) {
		return nil
	}

	storageType := vobj.StorageType(b.GetStorageType())

	eventDatasource := &bo.EventDatasource{
		TeamID: b.GetTeamId(),
		ID:     b.GetId(),
		Status: vobj.Status(b.GetStatus()),
		Conf: &conf.Event{
			Type:     storageType.String(),
			RocketMQ: &conf.RocketMQ{},
			Mqtt:     &conf.MQTT{},
			Kafka:    &conf.Kafka{},
		},
	}

	switch storageType {
	case vobj.StorageTypeKafka:
		kafka := conf.Kafka{Brokers: b.GetEndpoint()}
		_ = types.Unmarshal([]byte(b.GetConfig()), &kafka)
		eventDatasource.Conf.Kafka = &kafka
	case vobj.StorageTypeRocketMQ:
		rocketMQ := conf.RocketMQ{Endpoint: b.GetEndpoint()}
		_ = types.Unmarshal([]byte(b.GetConfig()), &rocketMQ)
		eventDatasource.Conf.RocketMQ = &rocketMQ
	case vobj.StorageTypeMQTT:
		mqtt := conf.MQTT{Broker: b.GetEndpoint()}
		_ = types.Unmarshal([]byte(b.GetConfig()), &mqtt)
		eventDatasource.Conf.Mqtt = &mqtt
	case vobj.StorageTypeRabbitMQ:
		// TODO: 未实现
	}

	return eventDatasource
}

func (b *DatasourceAPIBuilder) ToLogBo() *bo.LogDatasource {
	if types.IsNil(b) || types.IsNil(b.DatasourceItem) {
		return nil
	}
	storageType := vobj.StorageType(b.GetStorageType())

	datasource := &bo.LogDatasource{
		TeamID: b.GetTeamId(),
		ID:     b.GetId(),
		Status: vobj.Status(b.GetStatus()),
		Conf: &conf.LogQuery{
			Type: storageType.String(),
		},
	}

	switch storageType {
	case vobj.StorageTypeElasticsearch:
		elasticsearch := conf.Elasticsearch{Endpoint: b.GetEndpoint()}
		_ = types.Unmarshal([]byte(b.GetConfig()), &elasticsearch)
		datasource.Conf.Es = &elasticsearch
	case vobj.StorageTypeLoki:
		loki := conf.Loki{Endpoint: b.GetEndpoint()}
		_ = types.Unmarshal([]byte(b.GetConfig()), &loki)
		datasource.Conf.Loki = &loki
	case vobj.StorageAliYunSLS:
		aliYunLogConfig := conf.AliYunLogConfig{Endpoint: b.GetEndpoint()}
		_ = types.Unmarshal([]byte(b.GetConfig()), &aliYunLogConfig)
		datasource.Conf.AliYun = &aliYunLogConfig
	default:
		return nil
	}
	return datasource
}
