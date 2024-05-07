package do

import (
	"encoding/json"

	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg"
)

const TableNameMyChart = "my_charts"

const (
	MyChartFieldUserId = "user_id"
	MyChartFieldTitle  = "title"
	MyChartFieldUrl    = "url"
	MyChartFieldStatus = basescopes.BaseFieldStatus
	MyChartFieldRemark = basescopes.BaseFieldRemark
)

// MyChart 我的仪表盘
type MyChart struct {
	BaseModel
	UserId    uint32         `gorm:"column:user_id;type:int unsigned;not null;comment:用户ID"`
	Title     string         `gorm:"column:title;type:varchar(32);not null;comment:标题"`
	Remark    string         `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Url       string         `gorm:"column:url;type:text;not null;comment:地址"`
	Status    vobj.Status    `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
	ChartType vobj.ChartType `gorm:"column:category;type:tinyint;not null;default:0;comment:图表类型"`
	Width     string         `gorm:"column:width;type:varchar(12);not null;default:540px;comment:宽度"`
	Height    string         `gorm:"column:height;type:varchar(12);not null;default:320px;comment:高度"`
}

func (m *MyChart) TableName() string {
	return TableNameMyChart
}

// String MyChart
func (m *MyChart) String() string {
	if pkg.IsNil(m) {
		return "{}"
	}
	bs, _ := json.Marshal(m)
	return string(bs)
}
