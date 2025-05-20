package team

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

var _ do.DashboardChart = (*DashboardChart)(nil)

const tableNameDashboardChart = "team_dashboard_charts"

type DashboardChart struct {
	do.TeamModel
	DashboardID uint32            `gorm:"column:dashboard_id;type:int;not null;comment:dashboard ID" json:"dashboardID"`
	Title       string            `gorm:"column:title;type:varchar(255);not null;comment:title" json:"title"`
	Remark      string            `gorm:"column:remark;type:text;comment:remark" json:"remark"`
	Status      vobj.GlobalStatus `gorm:"column:status;type:tinyint;not null;default:0;comment:status" json:"status"`
	Url         string            `gorm:"column:url;type:varchar(255);not null;comment:URL" json:"url"`
	Width       string            `gorm:"column:width;type:varchar(255);not null;comment:width" json:"width"`
	Height      string            `gorm:"column:height;type:varchar(255);not null;comment:height" json:"height"`
	Dashboard   *Dashboard        `gorm:"foreignKey:DashboardID;references:ID" json:"dashboard"`
}

func (c *DashboardChart) GetDashboardID() uint32 {
	if c == nil {
		return 0
	}
	return c.DashboardID
}

func (c *DashboardChart) GetTitle() string {
	if c == nil {
		return ""
	}
	return c.Title
}

func (c *DashboardChart) GetRemark() string {
	if c == nil {
		return ""
	}
	return c.Remark
}

func (c *DashboardChart) GetStatus() vobj.GlobalStatus {
	if c == nil {
		return vobj.GlobalStatusUnknown
	}
	return c.Status
}

func (c *DashboardChart) GetDashboard() do.Dashboard {
	if c == nil || c.Dashboard == nil {
		return nil
	}
	return c.Dashboard
}

func (c *DashboardChart) GetUrl() string {
	if c == nil {
		return ""
	}
	return c.Url
}

func (c *DashboardChart) GetWidth() string {
	if c == nil {
		return ""
	}
	return c.Width
}

func (c *DashboardChart) GetHeight() string {
	if c == nil {
		return ""
	}
	return c.Height
}

func (c *DashboardChart) TableName() string {
	return tableNameDashboardChart
}
