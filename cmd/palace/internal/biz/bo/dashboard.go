package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
)

type Dashboard interface {
	GetID() uint32
	GetTitle() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetColorHex() string
}

// SaveDashboardReq represents a request to save a dashboard
type SaveDashboardReq struct {
	ID       uint32
	Title    string
	Remark   string
	Status   vobj.GlobalStatus
	ColorHex string
}

func (d *SaveDashboardReq) GetID() uint32 {
	if d == nil {
		return 0
	}
	return d.ID
}

func (d *SaveDashboardReq) GetTitle() string {
	return d.Title
}

func (d *SaveDashboardReq) GetRemark() string {
	return d.Remark
}

func (d *SaveDashboardReq) GetStatus() vobj.GlobalStatus {
	return d.Status
}

func (d *SaveDashboardReq) GetColorHex() string {
	return d.ColorHex
}

// ListDashboardReq represents a request to list dashboards
type ListDashboardReq struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *ListDashboardReq) ToListReply(dashboards []do.Dashboard) *ListDashboardReply {
	return &ListDashboardReply{
		PaginationReply: r.ToReply(),
		Items:           dashboards,
	}
}

// ListDashboardReply represents a reply to list dashboards
type ListDashboardReply = ListReply[do.Dashboard]

type SelectTeamDashboardReq struct {
	*PaginationRequest
	Status  vobj.GlobalStatus
	Keyword string
}

func (r *SelectTeamDashboardReq) ToSelectReply(dashboards []do.Dashboard) *SelectTeamDashboardReply {
	return &SelectTeamDashboardReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(dashboards, func(dashboard do.Dashboard) SelectItem {
			return &selectItem{
				Value:    dashboard.GetID(),
				Label:    dashboard.GetTitle(),
				Disabled: dashboard.GetDeletedAt() > 0 || !dashboard.GetStatus().IsEnable(),
				Extra: &selectItemExtra{
					Remark: dashboard.GetRemark(),
				},
			}
		}),
	}
}

type SelectTeamDashboardReply = ListReply[SelectItem]
