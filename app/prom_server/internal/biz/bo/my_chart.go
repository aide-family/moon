package bo

import (
	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
)

type (
	MyChartBO struct {
		Id     uint32    `json:"id"`
		UserId uint32    `json:"userId"`
		Title  string    `json:"title"`
		Remark string    `json:"remark"`
		Url    string    `json:"url"`
		Status vo.Status `json:"status"`
	}
)

// ToApiSelectV1 转换为api数据
func (b *MyChartBO) ToApiSelectV1() *api.MyChart {
	return &api.MyChart{
		Title:  b.Title,
		Remark: b.Remark,
		Url:    b.Url,
		Id:     b.Id,
	}
}

// ToApi 转换为api数据
func (b *MyChartBO) ToApi() *api.MyChart {
	return &api.MyChart{
		Title:  b.Title,
		Remark: b.Remark,
		Url:    b.Url,
		Id:     b.Id,
	}
}

// ToModel 转换为model数据
func (b *MyChartBO) ToModel() *do.MyChart {
	return &do.MyChart{
		BaseModel: do.BaseModel{ID: b.Id},
		UserId:    b.UserId,
		Title:     b.Title,
		Remark:    b.Remark,
		Url:       b.Url,
		Status:    b.Status,
	}
}

// MyChartModelToBO 转换为bo数据
func MyChartModelToBO(m *do.MyChart) *MyChartBO {
	return &MyChartBO{
		Id:     m.ID,
		UserId: m.UserId,
		Title:  m.Title,
		Remark: m.Remark,
		Url:    m.Url,
		Status: m.Status,
	}
}
