package bo

import (
	"fmt"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*Strategy)(nil)

type (
	// Strategy 策略明细
	Strategy struct {
		// 策略ID
		ID uint32 `json:"id,omitempty"`
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
		Labels *vobj.Labels `protobuf_val:"bytes,2,opt,name=value,proto3"`
		// 策略注解
		Annotations vobj.Annotations `protobuf_val:"bytes,2,opt,name=value,proto3"`
		// 执行频率
		Interval *types.Duration `json:"interval,omitempty"`
		// 数据源
		Datasource []*Datasource `json:"datasource,omitempty"`
		// 策略状态
		Status vobj.Status `json:"status,omitempty"`
		// 策略采样率
		Step uint32 `json:"step,omitempty"`
		// 判断条件
		Condition vobj.Condition `json:"condition,omitempty"`
		// 阈值
		Threshold float64 `json:"threshold,omitempty"`
	}

	// Datasource 数据源明细
	Datasource struct {
		// 数据源类型
		Category vobj.DatasourceType `json:"category,omitempty"`
		// 存储器类型
		StorageType vobj.StorageType `json:"storage_type,omitempty"`
		// 数据源配置 json
		Config map[string]string `json:"config,omitempty"`
		// 数据源地址
		Endpoint string `json:"endpoint,omitempty"`
	}
)

// Index 策略唯一索引
func (s *Strategy) Index() string {
	if types.IsNil(s) {
		return "-"
	}
	return fmt.Sprintf("%d", s.ID)
}

// Message 策略转消息
func (s *Strategy) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicStrategy)
}
