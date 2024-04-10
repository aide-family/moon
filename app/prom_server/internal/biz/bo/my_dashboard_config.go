package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

type (
	ListDashboardReq struct {
		Page    Pagination  `json:"page"`
		Keyword string      `json:"keyword"`
		Status  vobj.Status `json:"status"`
	}
	MyDashboardConfigBO struct {
		Id     uint32       `json:"id"`
		Status vobj.Status  `json:"status"`
		Remark string       `json:"remark"`
		Title  string       `json:"title"`
		Color  string       `json:"color"`
		UserId uint32       `json:"userId"`
		Charts []*MyChartBO `json:"charts"`

		CreatedAt int64 `json:"createdAt"`
		UpdatedAt int64 `json:"updatedAt"`
		DeletedAt int64 `json:"deletedAt"`
	}
)

// String json string
func (l *MyDashboardConfigBO) String() string {
	if l == nil {
		return "{}"
	}
	marshal, err := json.Marshal(l)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetCharts 获取图表列表
func (l *MyDashboardConfigBO) GetCharts() []*MyChartBO {
	if l == nil {
		return nil
	}
	return l.Charts
}

// ToApi 转换为API查询对象
func (l *MyDashboardConfigBO) ToApi() *api.MyDashboardConfig {
	if l == nil {
		return nil
	}
	return &api.MyDashboardConfig{
		Id:        l.Id,
		Title:     l.Title,
		Color:     l.Color,
		Charts:    slices.To(l.GetCharts(), func(i *MyChartBO) *api.MyChart { return i.ToApiSelectV1() }),
		Status:    l.Status.Value(),
		Remark:    l.Remark,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		DeletedAt: l.DeletedAt,
	}
}

// ToApiSelectV1 转换为API查询对象
func (l *MyDashboardConfigBO) ToApiSelectV1() *api.MyDashboardConfigOption {
	if l == nil {
		return nil
	}
	return &api.MyDashboardConfigOption{
		Value: l.Id,
		Label: l.Title,
		Color: l.Color,
	}
}

// ToModel 转换为实体
func (l *MyDashboardConfigBO) ToModel() *do.MyDashboardConfig {
	if l == nil {
		return nil
	}
	return &do.MyDashboardConfig{
		BaseModel: do.BaseModel{ID: l.Id},
		Title:     l.Title,
		Remark:    l.Remark,
		Color:     l.Color,
		UserId:    l.UserId,
		Status:    l.Status,
		Charts:    slices.To(l.GetCharts(), func(i *MyChartBO) *do.MyChart { return i.ToModel() }),
	}
}

// MyDashboardConfigModelToBO 实体转换业务对象
func MyDashboardConfigModelToBO(m *do.MyDashboardConfig) *MyDashboardConfigBO {
	if m == nil {
		return nil
	}
	return &MyDashboardConfigBO{
		Id:        m.ID,
		Status:    m.Status,
		Remark:    m.Remark,
		Title:     m.Title,
		Color:     m.Color,
		UserId:    m.UserId,
		Charts:    slices.To(m.GetCharts(), func(i *do.MyChart) *MyChartBO { return MyChartModelToBO(i) }),
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
		DeletedAt: int64(m.DeletedAt),
	}
}
