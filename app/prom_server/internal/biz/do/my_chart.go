package do

import (
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNameMyChart = "my_charts"

const (
	MyChartFieldUserId = "user_id"
	MyChartFieldTitle  = "title"
	MyChartFieldUrl    = "url"
	MyChartFieldStatus = "status"
	MyChartFieldRemark = "remark"
)

// MyChart 我的仪表盘
type MyChart struct {
	BaseModel
	UserId uint32      `gorm:"column:user_id;type:int unsigned;not null;comment:用户ID"`
	Title  string      `gorm:"column:title;type:varchar(32);not null;comment:标题"`
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Url    string      `gorm:"column:url;type:text;not null;comment:地址"`
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`
}

func (MyChart) TableName() string {
	return TableNameMyChart
}
