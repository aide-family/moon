package do

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type Dashboard interface {
	TeamBase
	GetTitle() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetColorHex() string
	GetCharts() []DashboardChart
}

type DashboardChart interface {
	TeamBase
	GetDashboardID() uint32
	GetTitle() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetDashboard() Dashboard
	GetUrl() string
	GetWidth() string
	GetHeight() string
}
