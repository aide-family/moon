package biz

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
	"github.com/aide-family/marksman/internal/conf"
)

const (
	metricName = "marksman_datasource_status"
)

func NewDatasource(
	bc *conf.Bootstrap,
	datasourceRepo repository.Datasource,
	levelRepo repository.Level,
	metricDatasourceQuerier repository.MetricDatasourceQuerier,
	evaluateBiz *Evaluate,
	helper *klog.Helper,
) *DatasourceBiz {
	mainTsdbConf := bc.GetMainTsdb()
	mainTsdb := &bo.DatasourceItemBo{
		URL:    mainTsdbConf.Url,
		Driver: mainTsdbConf.Driver,
		Type:   enum.DatasourceType_METRICS,
	}
	return &DatasourceBiz{
		mainTsdb:                mainTsdb,
		datasourceRepo:          datasourceRepo,
		levelRepo:               levelRepo,
		metricDatasourceQuerier: metricDatasourceQuerier,
		evaluateBiz:             evaluateBiz,
		helper:                  klog.NewHelper(klog.With(helper.Logger(), "biz", "datasource")),
	}
}

type DatasourceBiz struct {
	helper                  *klog.Helper
	mainTsdb                *bo.DatasourceItemBo
	datasourceRepo          repository.Datasource
	levelRepo               repository.Level
	metricDatasourceQuerier repository.MetricDatasourceQuerier
	evaluateBiz             *Evaluate
}

func (d *DatasourceBiz) CreateDatasource(ctx context.Context, req *bo.CreateDatasourceBo) (snowflake.ID, error) {
	if err := d.datasourceRepo.CheckDatasourceNameExist(ctx, req.Name); err != nil {
		d.helper.Errorw("msg", "check datasource name exist failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("check datasource name exist failed").WithCause(err)
	}
	if err := d.validateDatasourceLevel(ctx, req.LevelUID); err != nil {
		return 0, err
	}
	uid, err := d.datasourceRepo.CreateDatasource(ctx, req)
	if err != nil {
		d.helper.Errorw("msg", "create datasource failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("create datasource failed").WithCause(err)
	}
	return uid, nil
}

func (d *DatasourceBiz) UpdateDatasource(ctx context.Context, req *bo.UpdateDatasourceBo) error {
	if err := d.datasourceRepo.CheckDatasourceNameExist(ctx, req.Name, req.UID); err != nil {
		d.helper.Errorw("msg", "check datasource name exist failed", "error", err, "req", req)
		return merr.ErrorInternalServer("check datasource name exist failed").WithCause(err)
	}
	if err := d.validateDatasourceLevel(ctx, req.LevelUID); err != nil {
		return err
	}
	if err := d.datasourceRepo.UpdateDatasource(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("datasource %d not found", req.UID.Int64())
		}
		d.helper.Errorw("msg", "update datasource failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update datasource failed").WithCause(err)
	}
	d.evaluateBiz.SyncByDatasourceUID(ctx, req.UID)
	return nil
}

func (d *DatasourceBiz) validateDatasourceLevel(ctx context.Context, levelUID snowflake.ID) error {
	if levelUID == 0 {
		return merr.ErrorParams("datasource levelUid is required")
	}
	level, err := d.levelRepo.GetLevel(ctx, levelUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", levelUID.Int64())
		}
		d.helper.Errorw("msg", "get level failed", "error", err, "levelUID", levelUID)
		return merr.ErrorInternalServer("get level failed").WithCause(err)
	}
	if level.Type != enum.LevelType_DATASOURCE {
		return merr.ErrorParams("selected level is not a DATASOURCE level")
	}
	if level.Status != enum.GlobalStatus_ENABLED {
		return merr.ErrorParams("the selected level has been disabled, please select a new one")
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
	d.evaluateBiz.RemoveByDatasourceUID(ctx, uid)
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

func (d *DatasourceBiz) SelectDatasource(ctx context.Context, req *bo.SelectDatasourceBo) (*bo.SelectDatasourceReplyBo, error) {
	result, err := d.datasourceRepo.SelectDatasource(ctx, req)
	if err != nil {
		d.helper.Errorw("msg", "select datasource failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select datasource failed").WithCause(err)
	}
	return result, nil
}

func (d *DatasourceBiz) GetDatasourceStatus(ctx context.Context, req *bo.GetDatasourceStatusRequest) ([]*bo.DatasourceStatusSeriesBo, error) {
	query := fmt.Sprintf(`%s{uid="%s",name="%s"}`, metricName, strconv.FormatInt(req.GetUID(), 10), req.GetName())
	rangeParams := prometheusv1.Range{
		Start: time.Unix(req.GetStartTime(), 0),
		End:   time.Unix(req.GetEndTime(), 0),
		Step:  req.GetStep(),
	}
	matrix, err := d.metricDatasourceQuerier.QueryRange(ctx, d.mainTsdb, query, rangeParams)
	if err != nil {
		d.helper.Errorw("msg", "query datasource status failed", "error", err, "uid", req.UID)
		return nil, merr.ErrorInternalServer("query datasource status failed").WithCause(err)
	}
	out := make([]*bo.DatasourceStatusSeriesBo, 0, len(matrix))
	for _, stream := range matrix {
		points := make([]bo.DatasourceStatusPointBo, 0, len(stream.Values))
		for _, p := range stream.Values {
			points = append(points, bo.DatasourceStatusPointBo{
				Timestamp: int64(p.Timestamp) / 1000, // Prometheus model.Time is milliseconds
				Value:     float64(p.Value),
			})
		}
		out = append(out, &bo.DatasourceStatusSeriesBo{
			UID:    req.GetUID(),
			Name:   req.GetName(),
			Points: points,
		})
	}
	return out, nil
}

func (d *DatasourceBiz) ListMetrics(ctx context.Context, uid snowflake.ID) ([]*bo.MetricSummaryItemBo, error) {
	ds, err := d.datasourceRepo.GetDatasource(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		d.helper.Errorw("msg", "get datasource failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get datasource failed").WithCause(err)
	}
	return d.metricDatasourceQuerier.ListMetrics(ctx, ds)
}

func (d *DatasourceBiz) GetMetricDetail(ctx context.Context, uid snowflake.ID, metric string) (*bo.MetricDetailItemBo, error) {
	ds, err := d.datasourceRepo.GetDatasource(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("datasource %d not found", uid.Int64())
		}
		d.helper.Errorw("msg", "get datasource failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get datasource failed").WithCause(err)
	}
	return d.metricDatasourceQuerier.GetMetricDetail(ctx, ds, metric)
}
