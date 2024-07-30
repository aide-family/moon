package bo

import (
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// GetMetricsParams 查询指标请求参数
	GetMetricsParams struct {
		Endpoint string            `json:"endpoint"`
		Config   map[string]string `json:"config"`
		// 存储类型
		StorageType vobj.StorageType `json:"storageType"`
	}

	// MetricDetail 指标详情
	MetricDetail struct {
		// 指标名称
		Name string `json:"name"`
		// 帮助信息
		Help string `json:"help"`
		// 类型
		Type string `json:"type"`
		// 标签集合
		Labels map[string][]string `json:"labels"`
		// 指标单位
		Unit string `json:"unit"`
	}

	// QueryQLParams 查询QL请求参数
	QueryQLParams struct {
		// 查询指标请求参数
		GetMetricsParams
		// 查询QL
		QueryQL string `json:"queryQL"`
		// 时间范围
		TimeRange []string `json:"timeRange"`
		// 步长
		Step uint32 `json:"step"`
	}

	// PushMetricParams 推送指标请求参数
	PushMetricParams struct {
		// 指标明细
		*MetricDetail
		// 数据源ID
		DatasourceID uint32 `json:"datasourceId"`
		// 是否完成
		Done bool `json:"done"`
		// 团队ID
		TeamID uint32 `json:"teamId"`
	}
)
