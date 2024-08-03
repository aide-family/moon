package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	// ChartItem 仪表盘图表明细
	ChartItem struct {
		ID          uint32
		Name        string
		Remark      string
		URL         string
		Status      vobj.Status
		Height      string
		Width       string
		ChartType   vobj.DashboardChartType
		DashboardID uint32
	}

	// AddDashboardParams 添加仪表盘请求参数
	AddDashboardParams struct {
		Name           string
		Remark         string
		Color          string
		Charts         []*ChartItem
		StrategyGroups []uint32
	}

	// DeleteDashboardParams 删除仪表盘请求参数
	DeleteDashboardParams struct {
		ID     uint32
		Status vobj.Status
	}

	// UpdateDashboardParams 更新仪表盘请求参数
	UpdateDashboardParams struct {
		ID             uint32
		Name           string
		Remark         string
		Status         vobj.Status
		Color          string
		Charts         []*ChartItem
		StrategyGroups []uint32
	}

	// ListDashboardParams 仪表盘列表请求参数
	ListDashboardParams struct {
		Page    types.Pagination
		Keyword string
		Status  vobj.Status
	}
)
