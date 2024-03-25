package do

import (
	"gorm.io/gorm"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
)

const TableNameMyDashboardConfig = "my_dashboard_configs"

const (
	MyDashboardConfigFieldTitle         = "title"
	MyDashboardConfigFieldColor         = "color"
	MyDashboardConfigFieldUserId        = "user_id"
	MyDashboardConfigFieldRemark        = basescopes.BaseFieldRemark
	MyDashboardConfigFieldStatus        = basescopes.BaseFieldStatus
	MyDashboardConfigPreloadFieldCharts = "Charts"
)

func MyDashboardConfigPreloadCharts() basescopes.ScopeMethod {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload(MyDashboardConfigPreloadFieldCharts)
	}
}

// MyDashboardConfig 我的仪表盘配置
type MyDashboardConfig struct {
	BaseModel
	Title  string      `gorm:"column:title;type:varchar(32);not null;comment:大盘名称"`
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:备注"`
	Color  string      `gorm:"column:color;type:varchar(32);not null;comment:颜色"`
	UserId uint32      `gorm:"column:user_id;type:int unsigned;not null;comment:用户ID"`
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:状态"`

	Charts []*MyChart `gorm:"many2many:my_dashboard_config_charts;comment:图表"`
}

func (*MyDashboardConfig) TableName() string {
	return TableNameMyDashboardConfig
}

func (l *MyDashboardConfig) GetCharts() []*MyChart {
	if l == nil {
		return nil
	}
	return l.Charts
}
