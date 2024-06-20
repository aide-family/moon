package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	QueryMetricListParams struct {
		Page         types.Pagination `json:"page"`
		Keyword      string           `json:"keyword"`
		DatasourceID uint32           `json:"datasourceID"`
		MetricType   vobj.MetricType  `json:"metricType"`
	}

	GetMetricParams struct {
		ID           uint32 `json:"id"`
		WithRelation bool   `json:"withRelation"`
	}

	UpdateMetricParams struct {
		ID uint32 `json:"id"`
		// 单位
		Unit string `json:"unit"`
		// 描述
		Remark string `json:"remark"`
	}
)
