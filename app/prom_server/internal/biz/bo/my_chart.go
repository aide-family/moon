package bo

import (
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg"
)

type (
	MyChartBO struct {
		Id        uint32         `json:"id"`
		UserId    uint32         `json:"userId"`
		Title     string         `json:"title"`
		Remark    string         `json:"remark"`
		Url       string         `json:"url"`
		Status    vobj.Status    `json:"status"`
		ChartType vobj.ChartType `json:"chartType"`
		Width     string         `json:"width"`
		Height    string         `json:"height"`
	}
)

// MyChartModelToDO MyChartBO 转换为 MyChart
func MyChartModelToDO(m *MyChartBO) *do.MyChart {
	if pkg.IsNil(m) {
		return nil
	}
	return &do.MyChart{
		BaseModel: do.BaseModel{
			ID: m.Id,
		},
		Url:       m.Url,
		Status:    m.Status,
		ChartType: m.ChartType,
		Width:     m.Width,
		Height:    m.Height,
		UserId:    m.UserId,
		Title:     m.Title,
		Remark:    m.Remark,
	}
}

// MyChartModelToBO MyChart 转换为 MyChartBO
func MyChartModelToBO(m *do.MyChart) *MyChartBO {
	if pkg.IsNil(m) {
		return nil
	}
	return &MyChartBO{
		Id:        m.ID,
		UserId:    m.UserId,
		Title:     m.Title,
		Remark:    m.Remark,
		Url:       m.Url,
		Status:    m.Status,
		ChartType: m.ChartType,
		Width:     m.Width,
		Height:    m.Height,
	}
}
