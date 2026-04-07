package impl

import (
	"context"
	"strings"

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
	"github.com/aide-family/jade_tree/internal/data/impl/do"
	"github.com/aide-family/jade_tree/internal/data/impl/query"
)

func NewProbeTaskRepository(d *data.Data) repository.ProbeTask {
	query.SetDefault(d.DB())
	return &probeTaskRepository{Data: d}
}

type probeTaskRepository struct {
	*data.Data
}

func (r *probeTaskRepository) Create(ctx context.Context, in *bo.CreateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	row := &do.ProbeTask{
		BaseModel:      *new(do.BaseModel).WithCreator(in.Creator),
		Type:           in.Fields.Type,
		Host:           in.Fields.Host,
		Port:           in.Fields.Port,
		URL:            in.Fields.URL,
		Name:           in.Fields.Name,
		Status:         in.Fields.Status,
		TimeoutSeconds: in.Fields.TimeoutSeconds,
	}
	p := query.ProbeTask
	if err := p.WithContext(ctx).Create(row); err != nil {
		return nil, err
	}
	return convert.ToProbeTaskItemBo(row), nil
}

func (r *probeTaskRepository) Update(ctx context.Context, in *bo.UpdateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	p := query.ProbeTask
	columns := []field.AssignExpr{
		p.Type.Value(in.Fields.Type),
		p.Host.Value(in.Fields.Host),
		p.Port.Value(in.Fields.Port),
		p.URL.Value(in.Fields.URL),
		p.Name.Value(in.Fields.Name),
		p.TimeoutSeconds.Value(in.Fields.TimeoutSeconds),
	}
	info, err := p.WithContext(ctx).Where(p.ID.Eq(in.UID.Int64())).UpdateColumnSimple(columns...)
	if err != nil {
		return nil, err
	}
	if info.RowsAffected == 0 {
		return nil, merr.ErrorNotFound("probe task not found")
	}
	row, err := p.WithContext(ctx).Where(p.ID.Eq(in.UID.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("probe task not found")
		}
		return nil, err
	}
	return convert.ToProbeTaskItemBo(row), nil
}

func (r *probeTaskRepository) UpdateStatus(ctx context.Context, in *bo.UpdateProbeTaskStatusBo) (*bo.ProbeTaskItemBo, error) {
	p := query.ProbeTask
	info, err := p.WithContext(ctx).
		Where(p.ID.Eq(in.UID.Int64())).
		UpdateColumnSimple(p.Status.Value(int32(in.Status)))
	if err != nil {
		return nil, err
	}
	if info.RowsAffected == 0 {
		return nil, merr.ErrorNotFound("probe task not found")
	}
	row, err := p.WithContext(ctx).Where(p.ID.Eq(in.UID.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("probe task not found")
		}
		return nil, err
	}
	return convert.ToProbeTaskItemBo(row), nil
}

func (r *probeTaskRepository) Delete(ctx context.Context, uid snowflake.ID) error {
	p := query.ProbeTask
	info, err := p.WithContext(ctx).Where(p.ID.Eq(uid.Int64())).Delete()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("probe task not found")
	}
	return nil
}

func (r *probeTaskRepository) Get(ctx context.Context, uid snowflake.ID) (*bo.ProbeTaskItemBo, error) {
	p := query.ProbeTask
	row, err := p.WithContext(ctx).Where(p.ID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("probe task not found")
		}
		return nil, err
	}
	return convert.ToProbeTaskItemBo(row), nil
}

func (r *probeTaskRepository) List(ctx context.Context, req *bo.ListProbeTasksBo) (*bo.PageResponseBo[*bo.ProbeTaskItemBo], error) {
	p := query.ProbeTask
	w := p.WithContext(ctx)
	if req.Type != "" {
		w = w.Where(p.Type.Eq(req.Type))
	}
	if req.Keyword != "" {
		keyword := "%" + strings.TrimSpace(req.Keyword) + "%"
		w = w.Or(p.Name.Like(keyword), p.Host.Like(keyword), p.Port.Like(keyword), p.URL.Like(keyword))
	}
	if req.Status != enum.GlobalStatus_GlobalStatus_UNKNOWN {
		w = w.Where(p.Status.Eq(int32(req.Status)))
	}
	total, err := w.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)
	rows, err := w.Order(p.CreatedAt.Desc()).Limit(req.Limit()).Offset(req.Offset()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.ProbeTaskItemBo, 0, len(rows))
	for _, row := range rows {
		items = append(items, convert.ToProbeTaskItemBo(row))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *probeTaskRepository) ListEnabled(ctx context.Context) ([]*bo.ProbeTaskItemBo, error) {
	p := query.ProbeTask
	rows, err := p.WithContext(ctx).Where(p.Status.Eq(int32(enum.GlobalStatus_ENABLED))).Order(p.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.ProbeTaskItemBo, 0, len(rows))
	for _, row := range rows {
		items = append(items, convert.ToProbeTaskItemBo(row))
	}
	return items, nil
}
