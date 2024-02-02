package basescopes

import (
	"gorm.io/gorm"
)

const (
	MyDashboardConfigAssociationReplaceCharts = "Charts"
)

func MyDashboardConfigPreloadCharts(db *gorm.DB) *gorm.DB {
	return db.Preload(MyDashboardConfigAssociationReplaceCharts)
}
