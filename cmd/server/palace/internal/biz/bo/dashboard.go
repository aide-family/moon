package bo

import (
	"context"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
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
		Status         vobj.Status
		StrategyGroups []uint32
	}

	// DeleteDashboardParams 删除仪表盘请求参数
	DeleteDashboardParams struct {
		ID     uint32
		Status vobj.Status
	}

	// UpdateDashboardParams 更新仪表盘请求参数
	UpdateDashboardParams struct {
		ID        uint32
		Dashboard *AddDashboardParams
	}

	// ListDashboardParams 仪表盘列表请求参数
	ListDashboardParams struct {
		Page    types.Pagination
		Keyword string
		Status  vobj.Status
	}

	// BatchUpdateDashboardStatusParams 批量更新仪表盘状态请求参数
	BatchUpdateDashboardStatusParams struct {
		IDs    []uint32
		Status vobj.Status
	}

	// AddChartParams 添加图表请求参数
	AddChartParams struct {
		DashboardID uint32
		ChartItem   *ChartItem
	}

	// UpdateChartParams 更新图表请求参数
	UpdateChartParams struct {
		DashboardID uint32
		ChartItem   *ChartItem
	}

	// DeleteChartParams 删除图表请求参数
	DeleteChartParams struct {
		DashboardID uint32
		ChartID     uint32
	}

	// ListChartParams 获取图表列表请求参数
	ListChartParams struct {
		DashboardID uint32
		Page        types.Pagination
		Keyword     string
		Status      vobj.Status
		ChartTypes  []vobj.DashboardChartType
	}

	// GetChartParams 获取图表请求参数
	GetChartParams struct {
		DashboardID uint32
		ChartID     uint32
	}
)

// ToModel 转换为模型
func (p *ChartItem) ToModel(ctx context.Context) *bizmodel.DashboardChart {
	m := &bizmodel.DashboardChart{
		AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: p.ID}},
		Name:          p.Name,
		Status:        vobj.Status(p.Status),
		Remark:        p.Remark,
		URL:           p.URL,
		DashboardID:   p.DashboardID,
		ChartType:     vobj.DashboardChartType(p.ChartType),
		Width:         p.Width,
		Height:        p.Height,
	}
	m.WithContext(ctx)
	return m
}

// GetChartTypes ListChartParams 获取图表类型
func (p *ListChartParams) GetChartTypes() []int {
	return types.SliceTo(p.ChartTypes, func(chartType vobj.DashboardChartType) int {
		return int(chartType)
	})
}

// GetStrategyGroupDos 获取策略组列表
func (p *AddDashboardParams) GetStrategyGroupDos() []*bizmodel.StrategyGroup {
	return types.SliceTo(p.StrategyGroups, func(strategyGroupID uint32) *bizmodel.StrategyGroup {
		return &bizmodel.StrategyGroup{
			AllFieldModel: bizmodel.AllFieldModel{AllFieldModel: model.AllFieldModel{ID: strategyGroupID}},
		}
	})
}

// ToModel 转换为模型
func (p *AddDashboardParams) ToModel(ctx context.Context) *bizmodel.Dashboard {
	m := &bizmodel.Dashboard{
		Name:   p.Name,
		Status: vobj.Status(p.Status),
		Remark: p.Remark,
		Color:  p.Color,
	}
	m.WithContext(ctx)
	return m
}

// ToModel 转换为模型
func (p *UpdateDashboardParams) ToModel(ctx context.Context) *bizmodel.Dashboard {
	m := p.Dashboard.ToModel(ctx)
	m.ID = p.ID
	return m
}
