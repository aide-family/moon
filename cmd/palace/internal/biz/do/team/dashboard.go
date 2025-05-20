package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

var _ do.Dashboard = (*Dashboard)(nil)

const tableNameDashboard = "team_dashboards"

type Dashboard struct {
	do.TeamModel
	Title    string            `gorm:"column:title;type:varchar(255);not null;comment:title" json:"title"`
	Remark   string            `gorm:"column:remark;type:text;comment:remark" json:"remark"`
	Status   vobj.GlobalStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:status" json:"status"`
	ColorHex string            `gorm:"column:color_hex;type:varchar(20);not null;comment:color hex" json:"colorHex"`
	Charts   []*DashboardChart `gorm:"foreignKey:DashboardID;references:ID" json:"charts"`
}

func (d *Dashboard) GetTitle() string {
	if d == nil {
		return ""
	}
	return d.Title
}

func (d *Dashboard) GetRemark() string {
	if d == nil {
		return ""
	}
	return d.Remark
}

func (d *Dashboard) GetStatus() vobj.GlobalStatus {
	if d == nil {
		return vobj.GlobalStatusUnknown
	}
	return d.Status
}

func (d *Dashboard) GetColorHex() string {
	if d == nil {
		return ""
	}
	return d.ColorHex
}

func (d *Dashboard) GetCharts() []do.DashboardChart {
	if d == nil || d.Charts == nil {
		return nil
	}
	return slices.Map(d.Charts, func(chart *DashboardChart) do.DashboardChart { return chart })
}

func (d *Dashboard) TableName() string {
	return tableNameDashboard
}
