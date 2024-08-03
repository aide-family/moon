package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameDashboardChart = "dashboard_charts"

// DashboardChart mapped from table <dashboard_charts>
type DashboardChart struct {
	model.AllFieldModel
	Name        string                  `gorm:"column:name;type:varchar(64);not null;comment:仪表盘名称" json:"name"`     // 仪表盘名称
	Status      vobj.Status             `gorm:"column:status;type:int;not null;comment:仪表盘状态" json:"status"`         // 仪表盘状态
	Remark      string                  `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"` // 描述信息
	URL         string                  `gorm:"column:url;type:text;not null;comment:图表地址" json:"url"`
	DashboardID uint32                  `gorm:"column:dashboard_id;type:int unsigned;not null;comment:仪表盘ID" json:"dashboard_id"`
	ChartType   vobj.DashboardChartType `gorm:"column:chart_type;type:int;not null;comment:图表类型" json:"chart_type"`
	Width       string                  `gorm:"column:width;type:varchar(64);not null;comment:图表宽度" json:"width"`
	Height      string                  `gorm:"column:height;type:varchar(64);not null;comment:图表高度" json:"height"`
}

// String json string
func (c *DashboardChart) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *DashboardChart) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *DashboardChart) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName DashboardChart's table name
func (*DashboardChart) TableName() string {
	return tableNameDashboardChart
}
