package impl

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data"
	"github.com/aide-family/jade_tree/internal/data/impl/do"
)

func NewProbeTaskRepository(d *data.Data) repository.ProbeTask {
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
		Enabled:        in.Fields.Enabled,
		TimeoutSeconds: in.Fields.TimeoutSeconds,
	}
	if err := r.DB().WithContext(ctx).Create(row).Error; err != nil {
		return nil, err
	}
	return toProbeTaskItemBo(row), nil
}

func (r *probeTaskRepository) Update(ctx context.Context, in *bo.UpdateProbeTaskBo) (*bo.ProbeTaskItemBo, error) {
	var row do.ProbeTask
	if err := r.DB().WithContext(ctx).Where(&do.ProbeTask{BaseModel: do.BaseModel{ID: in.UID}}).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("probe task not found")
		}
		return nil, err
	}
	row.Type = in.Fields.Type
	row.Host = in.Fields.Host
	row.Port = in.Fields.Port
	row.URL = in.Fields.URL
	row.Name = in.Fields.Name
	row.Enabled = in.Fields.Enabled
	row.TimeoutSeconds = in.Fields.TimeoutSeconds
	if err := r.DB().WithContext(ctx).Save(&row).Error; err != nil {
		return nil, err
	}
	return toProbeTaskItemBo(&row), nil
}

func (r *probeTaskRepository) Delete(ctx context.Context, uid snowflake.ID) error {
	res := r.DB().WithContext(ctx).Where(&do.ProbeTask{BaseModel: do.BaseModel{ID: uid}}).Delete(&do.ProbeTask{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return merr.ErrorNotFound("probe task not found")
	}
	return nil
}

func (r *probeTaskRepository) Get(ctx context.Context, uid snowflake.ID) (*bo.ProbeTaskItemBo, error) {
	var row do.ProbeTask
	if err := r.DB().WithContext(ctx).Where(&do.ProbeTask{BaseModel: do.BaseModel{ID: uid}}).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("probe task not found")
		}
		return nil, err
	}
	return toProbeTaskItemBo(&row), nil
}

func (r *probeTaskRepository) List(ctx context.Context, req *bo.ListProbeTasksBo) (*bo.PageResponseBo[*bo.ProbeTaskItemBo], error) {
	w := r.DB().WithContext(ctx).Model(&do.ProbeTask{})
	var total int64
	if err := w.Count(&total).Error; err != nil {
		return nil, err
	}
	req.WithTotal(total)
	var rows []*do.ProbeTask
	if err := w.Limit(req.Limit()).Offset(req.Offset()).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.ProbeTaskItemBo, 0, len(rows))
	for _, row := range rows {
		items = append(items, toProbeTaskItemBo(row))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *probeTaskRepository) ListEnabled(ctx context.Context) ([]*bo.ProbeTaskItemBo, error) {
	var rows []*do.ProbeTask
	if err := r.DB().WithContext(ctx).Where(&do.ProbeTask{Enabled: true}).Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]*bo.ProbeTaskItemBo, 0, len(rows))
	for _, row := range rows {
		items = append(items, toProbeTaskItemBo(row))
	}
	return items, nil
}

func toProbeTaskItemBo(row *do.ProbeTask) *bo.ProbeTaskItemBo {
	return &bo.ProbeTaskItemBo{
		UID:            row.ID,
		Type:           row.Type,
		Host:           row.Host,
		Port:           row.Port,
		URL:            row.URL,
		Name:           row.Name,
		Enabled:        row.Enabled,
		TimeoutSeconds: row.TimeoutSeconds,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}
