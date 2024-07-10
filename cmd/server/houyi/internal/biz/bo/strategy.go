package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	Strategy struct {
		// 策略名称
		Alert string `json:"alert,omitempty"`
		// 策略语句
		Expr string `json:"expr,omitempty"`
		// 策略持续时间
		For *types.Duration `json:"for,omitempty"`
		// 持续次数
		Count uint32 `json:"count,omitempty"`
		// 持续的类型
		SustainType vobj.Sustain `json:"sustainType,omitempty"`
		// 多数据源持续类型
		MultiDatasourceSustainType vobj.MultiDatasourceSustain `json:"multiDatasourceSustainType,omitempty"`
		// 策略标签
		Labels map[string]string `protobuf_val:"bytes,2,opt,name=value,proto3"`
		// 策略注解
		Annotations map[string]string `protobuf_val:"bytes,2,opt,name=value,proto3"`
		// 执行频率
		Interval *types.Duration `json:"interval,omitempty"`
		// 数据源
		Datasource []*Datasource `json:"datasource,omitempty"`
	}

	Datasource struct {
		// 数据源类型
		Category vobj.DatasourceType `json:"category,omitempty"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storage_type,omitempty"`
		// 数据源配置 json
		Config string `json:"config,omitempty"`
		// 数据源地址
		Endpoint string `json:"endpoint,omitempty"`
	}
)
