package bo

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type TeamAuditListRequest struct {
	*PaginationRequest
	Status  []vobj.StatusAudit `json:"status"`
	Keyword string             `json:"keyword"`
	Actions []vobj.AuditAction `json:"actions"`
	UserID  uint32             `json:"userId"`
}

func (r *TeamAuditListRequest) ToTeamAuditListReply(items []*system.TeamAudit) *TeamAuditListReply {
	return &TeamAuditListReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(items, func(item *system.TeamAudit) do.TeamAudit { return item }),
	}
}

type TeamAuditListReply = ListReply[do.TeamAudit]

type UpdateTeamAuditStatus interface {
	GetAuditID() uint32
	GetStatus() vobj.StatusAudit
	GetReason() string
}

type UpdateTeamAuditStatusReq struct {
	AuditID uint32           `json:"auditId"`
	Status  vobj.StatusAudit `json:"status"`
	Reason  string           `json:"reason"`
	auditDo do.TeamAudit
}

func (r *UpdateTeamAuditStatusReq) GetAuditID() uint32 {
	if r == nil {
		return 0
	}
	return r.AuditID
}

func (r *UpdateTeamAuditStatusReq) GetStatus() vobj.StatusAudit {
	if r == nil {
		return vobj.AuditStatusUnknown
	}
	return r.Status
}

func (r *UpdateTeamAuditStatusReq) GetReason() string {
	if r == nil {
		return ""
	}
	return r.Reason
}

func (r *UpdateTeamAuditStatusReq) Validate() error {
	if r.AuditID <= 0 {
		return merr.ErrorParamsError("invalid audit id")
	}
	if r.Status == vobj.AuditStatusUnknown {
		return merr.ErrorParamsError("invalid audit status")
	}
	if validate.IsNil(r.auditDo) {
		return merr.ErrorParamsError("audit is nil")
	}

	if r.auditDo.GetStatus().IsFinal() {
		return merr.ErrorParamsError("audit status is final")
	}
	return nil
}

func (r *UpdateTeamAuditStatusReq) WithTeamAudit(audit do.TeamAudit) *UpdateTeamAuditStatusReq {
	r.auditDo = audit
	return r
}
