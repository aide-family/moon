package impl

import (
	"context"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data"
	"github.com/aide-family/jade_tree/internal/data/impl/convert"
	"github.com/aide-family/jade_tree/internal/data/impl/query"
)

func NewCommandAuditRepository(d *data.Data) repository.CommandAudit {
	query.SetDefault(d.DB())
	return &commandAuditRepository{Data: d}
}

type commandAuditRepository struct {
	*data.Data
}

func (r *commandAuditRepository) Create(ctx context.Context, in *bo.CommandAuditCreateRepoBo) (*bo.SSHCommandAuditItemBo, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("audit create input is required")
	}
	row := convert.ToSSHCommandAuditDO(in)
	a := query.SSHCommandAudit
	if err := a.WithContext(ctx).Create(row); err != nil {
		return nil, err
	}
	return convert.ToSSHCommandAuditItemBo(row), nil
}

func (r *commandAuditRepository) Get(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandAuditItemBo, error) {
	a := query.SSHCommandAudit
	row, err := a.WithContext(ctx).Where(a.ID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("ssh command audit not found")
		}
		return nil, err
	}
	return convert.ToSSHCommandAuditItemBo(row), nil
}

func (r *commandAuditRepository) List(ctx context.Context, req *bo.ListSSHCommandAuditsBo) (*bo.PageResponseBo[*bo.SSHCommandAuditItemBo], error) {
	a := query.SSHCommandAudit
	w := a.WithContext(ctx)
	if req.StatusFilter != enum.SSHCommandAuditStatus_SSHCommandAuditStatus_UNKNOWN {
		w = w.Where(a.Status.Eq(int32(req.StatusFilter)))
	}
	if req.PageRequestBo != nil {
		total, err := w.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		w = w.Limit(req.Limit()).Offset(req.Offset())
	}
	rows, err := w.Order(a.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.SSHCommandAuditItemBo, 0, len(rows))
	for _, row := range rows {
		items = append(items, convert.ToSSHCommandAuditItemBo(row))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *commandAuditRepository) Reject(ctx context.Context, in *bo.CommandAuditRejectBo) (*bo.SSHCommandAuditItemBo, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("reject input is required")
	}
	a := query.SSHCommandAudit
	now := time.Now()
	columns := []field.AssignExpr{
		a.Status.Value(int32(enum.SSHCommandAuditStatus_SSHCommandAuditStatus_REJECTED)),
		a.RejectReason.Value(in.Reason),
		a.Reviewer.Value(in.Reviewer.Int64()),
		a.ReviewedAt.Value(now),
	}
	info, err := a.WithContext(ctx).
		Where(a.ID.Eq(in.AuditUID.Int64()), a.Status.Eq(int32(enum.SSHCommandAuditStatus_SSHCommandAuditStatus_PENDING))).
		UpdateColumnSimple(columns...)
	if err != nil {
		return nil, err
	}
	if info.RowsAffected == 0 {
		return nil, merr.ErrorInvalidArgument("audit is not pending or does not exist")
	}
	return r.Get(ctx, in.AuditUID)
}

func (r *commandAuditRepository) Approve(ctx context.Context, uid, reviewer snowflake.ID) (*bo.SSHCommandAuditItemBo, *bo.SSHCommandItemBo, error) {
	var cmdOut *bo.SSHCommandItemBo
	err := r.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := query.Use(tx)
		a := q.SSHCommandAudit
		aud, err := a.WithContext(ctx).Where(a.ID.Eq(uid.Int64())).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return merr.ErrorNotFound("ssh command audit not found")
			}
			return err
		}
		if aud.Status != enum.SSHCommandAuditStatus_SSHCommandAuditStatus_PENDING {
			return merr.ErrorInvalidArgument("audit is not pending")
		}
		now := time.Now()
		sc := q.SSHCommand
		switch aud.Kind {
		case enum.SSHCommandAuditKind_SSHCommandAuditKind_CREATE:
			cnt, err := sc.WithContext(ctx).Where(sc.Name.Eq(aud.Name)).Count()
			if err != nil {
				return err
			}
			if cnt > 0 {
				return merr.ErrorInvalidArgument("command name already exists")
			}
			row := convert.ToSSHCommandDO(aud.Creator, convert.ToSSHCommandFieldsFromAudit(aud))
			if err := sc.WithContext(ctx).Create(row); err != nil {
				return err
			}
			cmdOut = convert.ToSSHCommandItemBo(row)
		case enum.SSHCommandAuditKind_SSHCommandAuditKind_UPDATE:
			if aud.TargetCommandID == 0 {
				return merr.ErrorInvalidArgument("target command is required for update audit")
			}
			dup, err := sc.WithContext(ctx).Where(sc.Name.Eq(aud.Name), sc.ID.Neq(aud.TargetCommandID.Int64())).Count()
			if err != nil {
				return err
			}
			if dup > 0 {
				return merr.ErrorInvalidArgument("command name already exists")
			}
			columns := []field.AssignExpr{
				sc.Name.Value(aud.Name),
				sc.Description.Value(aud.Description),
				sc.Content.Value(aud.Content),
				sc.WorkDir.Value(aud.WorkDir),
				sc.Env.Value(aud.Env),
			}
			info, err := sc.WithContext(ctx).Where(sc.ID.Eq(aud.TargetCommandID.Int64())).UpdateColumnSimple(columns...)
			if err != nil {
				return err
			}
			if info.RowsAffected == 0 {
				return merr.ErrorNotFound("ssh command not found")
			}
			updated, err := sc.WithContext(ctx).Where(sc.ID.Eq(aud.TargetCommandID.Int64())).First()
			if err != nil {
				return err
			}
			cmdOut = convert.ToSSHCommandItemBo(updated)
		default:
			return merr.ErrorInvalidArgument("unsupported audit kind")
		}
		approveColumns := []field.AssignExpr{
			a.Status.Value(int32(enum.SSHCommandAuditStatus_SSHCommandAuditStatus_APPROVED)),
			a.Reviewer.Value(reviewer.Int64()),
			a.ReviewedAt.Value(now),
		}
		_, err = a.WithContext(ctx).
			Where(a.ID.Eq(uid.Int64())).
			UpdateColumnSimple(approveColumns...)
		return err
	})
	if err != nil {
		return nil, nil, err
	}
	auditOut, err := r.Get(ctx, uid)
	if err != nil {
		return nil, nil, err
	}
	return auditOut, cmdOut, nil
}
