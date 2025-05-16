package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

type DashboardChart interface {
	GetID() uint32
	GetDashboardID() uint32
	GetTitle() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetUrl() string
	GetWidth() string
	GetHeight() string
}

// SaveDashboardChartReq represents a request to save a dashboard chart
type SaveDashboardChartReq struct {
	ID          uint32
	DashboardID uint32
	Title       string
	Remark      string
	Status      vobj.GlobalStatus
	Url         string
	Width       string
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

func (d *SaveDashboardChartReq) GetStatus() vobj.GlobalStatus {
	if d == nil {
		return vobj.GlobalStatusUnknown
	}
	return d.Status
}

func (d *SaveDashboardChartReq) GetUrl() string {
	if d == nil {
		return ""
	}
	return d.Url
}

func (d *SaveDashboardChartReq) GetWidth() string {
	if d == nil {
		return ""
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
