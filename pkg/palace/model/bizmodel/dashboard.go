package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameDashboard = "dashboard"

// Dashboard mapped from table <dashboard>
type Dashboard struct {
	model.AllFieldModel
	Name   string      `gorm:"column:name;type:varchar(64);not null;comment:仪表盘名称" json:"name"`     // 仪表盘名称
	Status vobj.Status `gorm:"column:status;type:int;not null;comment:仪表盘状态" json:"status"`         // 仪表盘状态
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"` // 描述信息
	Color  string      `gorm:"column:color;type:varchar(64);not null;comment:颜色" json:"color"`
	// 仪表盘图表
	Charts []*DashboardChart `gorm:"foreignKey:DashboardID" json:"charts"`
	// 仪表盘策略组
	StrategyGroups []*StrategyGroup `gorm:"many2many:dashboard_strategy_groups" json:"strategy_groups"`
}

// String json string
func (c *Dashboard) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *Dashboard) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *Dashboard) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName Dashboard's table name
func (*Dashboard) TableName() string {
	return tableNameDashboard
}
