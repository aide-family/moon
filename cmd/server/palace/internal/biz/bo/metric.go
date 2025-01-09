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
		Name      string                           `json:"name"`
		Help      string                           `json:"help"`
		Type      vobj.MetricType                  `json:"type"`
		Unit      string                           `json:"unit"`
		Labels    []*MetricLabel                   `json:"labels"`
		MapLabels map[string]*bizmodel.MetricLabel `json:"mapLabels"`
	}

	// CreateMetricParams 创建指标请求参数
	CreateMetricParams struct {
		Metric       *MetricBo `json:"metric"`
		Done         bool      `json:"done"`
		DatasourceID uint32    `json:"datasourceID"`
		TeamID       uint32    `json:"teamId"`
	}
)

// GetMapLabels 获取标签map
func (m *MetricBo) GetMapLabels() map[string]*bizmodel.MetricLabel {
	if types.IsNil(m) || types.IsNil(m.MapLabels) {
		return map[string]*bizmodel.MetricLabel{}
	}
	return m.MapLabels
}

// ToMetricModel 转换成数据库Metric模型
func (c *CreateMetricParams) ToMetricModel() *bizmodel.DatasourceMetric {
	if types.IsNil(c) || types.IsNil(c.Metric) {
		return nil
	}

	return &bizmodel.DatasourceMetric{
		Name:         c.Metric.Name,
		Category:     c.Metric.Type,
		Unit:         c.Metric.Unit,
		Remark:       c.Metric.Help,
		DatasourceID: c.DatasourceID,
	}
}
