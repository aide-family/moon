package build

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
)

// ToDashboardItem converts a business object to a proto object
func ToDashboardItem(dashboard do.Dashboard) *common.TeamDashboardItem {
	if dashboard == nil {
		return nil
	}

	return &common.TeamDashboardItem{
		TeamDashboardId: dashboard.GetID(),
		Name:            dashboard.GetTitle(),
		Remark:          dashboard.GetRemark(),
		Status:          common.GlobalStatus(dashboard.GetStatus().GetValue()),
		ColorHex:        dashboard.GetColorHex(),
		CreatedAt:       timex.Format(dashboard.GetCreatedAt()),
		UpdatedAt:       timex.Format(dashboard.GetUpdatedAt()),
	}
}

// ToDashboardItems converts multiple business objects to proto objects
func ToDashboardItems(dashboards []do.Dashboard) []*common.TeamDashboardItem {
	return slices.Map(dashboards, ToDashboardItem)
}

// ToDashboardChartItem converts a business object to a proto object
func ToDashboardChartItem(chart do.DashboardChart) *common.TeamDashboardChartItem {
	if chart == nil {
		return nil
	}

	return &common.TeamDashboardChartItem{
		TeamDashboardChartId: chart.GetID(),
		DashboardId:          chart.GetDashboardID(),
		Title:                chart.GetTitle(),
		Remark:               chart.GetRemark(),
		Status:               common.GlobalStatus(chart.GetStatus().GetValue()),
		Url:                  chart.GetUrl(),
		Width:                chart.GetWidth(),
		Height:               chart.GetHeight(),
		CreatedAt:            timex.Format(chart.GetCreatedAt()),
		UpdatedAt:            timex.Format(chart.GetUpdatedAt()),
	}
}

// ToDashboardChartItems converts multiple business objects to proto objects
func ToDashboardChartItems(charts []do.DashboardChart) []*common.TeamDashboardChartItem {
	return slices.Map(charts, ToDashboardChartItem)
}
