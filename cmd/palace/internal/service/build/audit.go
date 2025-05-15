package build

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToTeamAuditListRequest(req *palace.GetTeamAuditListRequest) *bo.TeamAuditListRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.TeamAuditListRequest{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status:            slices.Map(req.GetStatus(), func(status common.TeamAuditStatus) vobj.StatusAudit { return vobj.StatusAudit(status) }),
		Keyword:           req.GetKeyword(),
		Actions:           slices.Map(req.GetActions(), func(action common.TeamAuditAction) vobj.AuditAction { return vobj.AuditAction(action) }),
		UserID:            req.GetUserId(),
	}
}

func ToTeamAuditItem(audit do.TeamAudit) *common.TeamAuditItem {
	if validate.IsNil(audit) {
		return nil
	}
	return &common.TeamAuditItem{
		TeamAuditId: audit.GetID(),
		User:        ToUserBaseItem(audit.GetCreator()),
		Status:      common.TeamAuditStatus(audit.GetStatus().GetValue()),
		Reason:      audit.GetReason(),
		CreatedAt:   timex.Format(audit.GetCreatedAt()),
		Team:        ToTeamBaseItem(audit.GetTeam()),
		Action:      common.TeamAuditAction(audit.GetAction().GetValue()),
	}
}

func ToTeamAuditItems(audits []do.TeamAudit) []*common.TeamAuditItem {
	return slices.Map(audits, ToTeamAuditItem)
}
