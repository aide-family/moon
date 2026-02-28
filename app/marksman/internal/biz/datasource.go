package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewDatasource(
	datasourceRepo repository.Datasource,
	helper *klog.Helper,
) *DatasourceBiz {
	return &DatasourceBiz{
		datasourceRepo: datasourceRepo,
		helper:         klog.NewHelper(klog.With(helper.Logger(), "biz", "datasource")),
	}
}

type DatasourceBiz struct {
	helper         *klog.Helper
	datasourceRepo repository.Datasource
}

func (d *DatasourceBiz) CreateDatasource(ctx context.Context, req *bo.CreateDatasourceBo) error {
	if err := d.datasourceRepo.CreateDatasource(ctx, req); err != nil {
		d.helper.Errorw("msg", "create datasource failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create datasource failed").WithCause(err)
	}
	return nil
}

func (d *DatasourceBiz) UpdateDatasource(ctx context.Context, req *bo.UpdateDatasourceBo) error {
	if err := d.datasourceRepo.UpdateDatasource(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("datasource %d not found", req.UID.Int64())
		}
		d.helper.Errorw("msg", "update datasource failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update datasource failed").WithCause(err)
	}
	return nil
}

func (d *DatasourceBiz) DeleteDatasource(ctx context.Context, uid snowflake.ID) error {
	if err := d.datasourceRepo.DeleteDatasource(ctx, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		d.helper.Errorw("msg", "delete datasource failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete datasource failed").WithCause(err)
	}
	return nil
}

func (d *DatasourceBiz) GetDatasource(ctx context.Context, uid snowflake.ID) (*bo.DatasourceItemBo, error) {
	item, err := d.datasourceRepo.GetDatasource(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		d.helper.Errorw("msg", "get datasource failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get datasource failed").WithCause(err)
	}
	return item, nil
}

func (d *DatasourceBiz) ListDatasource(ctx context.Context, req *bo.ListDatasourceBo) (*bo.PageResponseBo[*bo.DatasourceItemBo], error) {
	result, err := d.datasourceRepo.ListDatasource(ctx, req)
	if err != nil {
		d.helper.Errorw("msg", "list datasource failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list datasource failed").WithCause(err)
	}
	return result, nil
}
