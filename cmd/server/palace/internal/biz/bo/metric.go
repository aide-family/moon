package bo

import (
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// QueryMetricListParams 查询指标列表请求参数
	QueryMetricListParams struct {
		Page         types.Pagination `json:"page"`
		Keyword      string           `json:"keyword"`
		DatasourceID uint32           `json:"datasourceID"`
		MetricType   vobj.MetricType  `json:"metricType"`
	}

	// GetMetricParams 获取指标请求参数
	GetMetricParams struct {
		ID           uint32 `json:"id"`
		WithRelation bool   `json:"withRelation"`
	}

	// UpdateMetricParams 更新指标请求参数
	UpdateMetricParams struct {
		ID uint32 `json:"id"`
		// 单位
		Unit string `json:"unit"`
		// 描述
		Remark string `json:"remark"`
	}

	// MetricLabel 指标标签
	MetricLabel struct {
		Name   string   `json:"name"`
		Values []string `json:"values"`
	}

	// MetricBo 指标明细
	MetricBo struct {
		Name   string          `json:"name"`
		Help   string          `json:"help"`
		Type   vobj.MetricType `json:"type"`
		Unit   string          `json:"unit"`
		Labels []*MetricLabel  `json:"labels"`
	}

	// CreateMetricParams 创建指标请求参数
	CreateMetricParams struct {
		Metric       *MetricBo `json:"metric"`
		Done         bool      `json:"done"`
		DatasourceID uint32    `json:"datasourceID"`
		TeamID       uint32    `json:"teamId"`
	}
)

// ToModel 转换成数据库模型
func (c *CreateMetricParams) ToModel() *bizmodel.DatasourceMetric {
	if types.IsNil(c) || types.IsNil(c.Metric) {
		return nil
	}
	return &bizmodel.DatasourceMetric{
		Name:         c.Metric.Name,
		Category:     c.Metric.Type,
		Unit:         c.Metric.Unit,
		Remark:       c.Metric.Help,
		DatasourceID: c.DatasourceID,
		Labels: types.SliceTo(c.Metric.Labels, func(label *MetricLabel) *bizmodel.MetricLabel {
			return &bizmodel.MetricLabel{
				Name: label.Name,
				LabelValues: types.SliceTo(label.Values, func(value string) *bizmodel.MetricLabelValue {
					return &bizmodel.MetricLabelValue{
						Name: value,
					}
				}),
			}
		}),
	}
}
