package bo

import (
	"github.com/aide-cloud/moon/pkg/vobj"
)

type (
	GetMetricsParams struct {
		Endpoint string            `json:"endpoint"`
		Config   map[string]string `json:"config"`
		// 存储类型
		StorageType vobj.StorageType `json:"storageType"`
	}

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
)
