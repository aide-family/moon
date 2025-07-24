package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

type DashboardChart interface {
	GetID() uint32
	GetDashboardID() uint32
	GetTitle() string
	GetRemark() string
	GetURL() string
	GetWidth() uint32
	GetHeight() string
}

// SaveDashboardChartReq represents a request to save a dashboard chart
type SaveDashboardChartReq struct {
	ID          uint32
	DashboardID uint32
	Title       string
	Remark      string
	URL         string
	Width       uint32
	Height      string
}

func (d *SaveDashboardChartReq) GetID() uint32 {
	if d == nil {
		return 0
	}
	return d.ID
}

func (d *SaveDashboardChartReq) GetDashboardID() uint32 {
	if d == nil {
		return 0
	}
	return d.DashboardID
}

func (d *SaveDashboardChartReq) GetTitle() string {
	if d == nil {
		return ""
	}
	return d.Title
}

func (d *SaveDashboardChartReq) GetRemark() string {
	if d == nil {
		return ""
	}
	return d.Remark
}

func (d *SaveDashboardChartReq) GetURL() string {
	if d == nil {
		return ""
	}
	return d.URL
}

func (d *SaveDashboardChartReq) GetWidth() uint32 {
	if d == nil {
		return 6
	}
	return d.Width
}

func (d *SaveDashboardChartReq) GetHeight() string {
	if d == nil {
		return ""
	}
	return d.Height
}

// ListDashboardChartReq represents a request to list dashboard charts
type ListDashboardChartReq struct {
	*PaginationRequest
	Status      vobj.GlobalStatus
	DashboardID uint32
	Keyword     string
}

func (r *ListDashboardChartReq) ToListReply(charts []do.DashboardChart) *ListDashboardChartReply {
	return &ListDashboardChartReply{
		PaginationReply: r.ToReply(),
		Items:           charts,
	}
}

// ListDashboardChartReply represents a reply to list dashboard charts
type ListDashboardChartReply = ListReply[do.DashboardChart]

type SelectTeamDashboardChartReq struct {
	*PaginationRequest
	Status      vobj.GlobalStatus
	DashboardID uint32
	Keyword     string
}

func (r *SelectTeamDashboardChartReq) ToSelectReply(charts []do.DashboardChart) *SelectTeamDashboardChartReply {
	return &SelectTeamDashboardChartReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(charts, func(chart do.DashboardChart) SelectItem {
			return &selectItem{
				Value:    chart.GetID(),
				Label:    chart.GetTitle(),
				Disabled: chart.GetDeletedAt() > 0 || !chart.GetStatus().IsEnable(),
				Extra: &selectItemExtra{
					Remark: chart.GetRemark(),
				},
			}
		}),
	}
}

type SelectTeamDashboardChartReply = ListReply[SelectItem]

// BatchUpdateDashboardStatusReq represents a request to batch update dashboard status
type BatchUpdateDashboardStatusReq struct {
	Ids    []uint32
	Status vobj.GlobalStatus
}

// BatchUpdateDashboardChartStatusReq represents a request to batch update dashboard chart status
type BatchUpdateDashboardChartStatusReq struct {
	Ids         []uint32
	DashboardID uint32
	Status      vobj.GlobalStatus
}

type OperateOneDashboardChartReq struct {
	ID          uint32 `json:"id"`
	DashboardID uint32 `json:"dashboardID"`
}
