package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewAuditRepo(d *data.Data, logger log.Logger) repository.Audit {
	return &auditImpl{
		Data:   d,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.audit")),
	}
}

type auditImpl struct {
	*data.Data
	helper *log.Helper
}

func (a *auditImpl) Get(ctx context.Context, id uint32) (do.TeamAudit, error) {
	auditQuery := getMainQuery(ctx, a).TeamAudit
	audit, err := auditQuery.WithContext(ctx).Where(auditQuery.ID.Eq(id)).First()
	if err != nil {
		return nil, auditNotFound(err)
	}
	return audit, nil
}

func (a *auditImpl) TeamAuditList(ctx context.Context, req *bo.TeamAuditListRequest) (*bo.TeamAuditListReply, error) {
	auditQuery := getMainQuery(ctx, a).TeamAudit
	wrapper := auditQuery.WithContext(ctx)

	if len(req.Status) > 0 {
		status := slices.Map(req.Status, func(statusItem vobj.StatusAudit) int8 { return statusItem.GetValue() })
		wrapper = wrapper.Where(auditQuery.Status.In(status...))
	}
	if len(req.Actions) > 0 {
		actions := slices.Map(req.Actions, func(actionItem vobj.AuditAction) int8 { return actionItem.GetValue() })
		wrapper = wrapper.Where(auditQuery.Action.In(actions...))
	}
	if !validate.TextIsNull(req.Keyword) {
		wrapper = wrapper.Where(auditQuery.Reason.Like(req.Keyword))
	}
	if req.UserID > 0 {
		wrapper = wrapper.Where(auditQuery.CreatorID.Eq(req.UserID))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
		req.WithTotal(total)
	}
	audits, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(audits, func(audit *system.TeamAudit) do.TeamAudit { return audit })
	return req.ToListReply(rows), nil
}

func (a *auditImpl) UpdateTeamAuditStatus(ctx context.Context, req bo.UpdateTeamAuditStatus) error {
	auditMutation := getMainQuery(ctx, a).TeamAudit
	_, err := auditMutation.WithContext(ctx).
		Where(auditMutation.ID.Eq(req.GetAuditID())).
		UpdateSimple(
			auditMutation.Status.Value(req.GetStatus().GetValue()),
			auditMutation.Reason.Value(req.GetReason()),
		)
	return err
}
