package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewServerBiz(
	serverRepo repository.Server,
	houyiRepo repository.Houyi,
	metricDatasourceRepo repository.TeamDatasourceMetric,
	metricStrategyRepo repository.TeamStrategyMetric,
	teamRepo repository.Team,
	logger log.Logger,
) *Server {
	return &Server{
		serverRepo:           serverRepo,
		houyiRepo:            houyiRepo,
		metricDatasourceRepo: metricDatasourceRepo,
		metricStrategyRepo:   metricStrategyRepo,
		teamRepo:             teamRepo,
		helper:               log.NewHelper(log.With(logger, "module", "biz.server")),
	}
}

type Server struct {
	serverRepo           repository.Server
	houyiRepo            repository.Houyi
	metricDatasourceRepo repository.TeamDatasourceMetric
	metricStrategyRepo   repository.TeamStrategyMetric
	teamRepo             repository.Team
	helper               *log.Helper
}

func (b *Server) Register(ctx context.Context, req *bo.ServerRegisterReq) error {
	if validate.IsNil(req) {
		return merr.ErrorInvalidArgument("invalid request")
	}

	if !req.IsOnline {
		return b.Deregister(ctx, req)
	}

	if err := b.serverRepo.RegisterServer(ctx, req); err != nil {
		return merr.ErrorInternalServer("failed to register server: %v", err)
	}

	b.helper.WithContext(ctx).Debugf("registered server type: %v, uuid: %s", req.ServerType, req.Uuid)
	return nil
}

func (b *Server) Deregister(ctx context.Context, req *bo.ServerRegisterReq) error {
	if validate.IsNil(req) {
		return merr.ErrorInvalidArgument("invalid request")
	}
	if req.IsOnline {
		return b.Register(ctx, req)
	}
	if err := b.serverRepo.DeregisterServer(ctx, req); err != nil {
		return merr.ErrorInternalServer("failed to deregister server: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("deregistered server type: %v, uuid: %s", req.ServerType, req.Uuid)
	return nil
}

func (b *Server) SyncMetricDatasource(ctx context.Context, changedMetricDatasource bo.ChangedMetricDatasource) error {
	houyi, ok := b.houyiRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get houyi client")
	}
	eg := new(errgroup.Group)
	for teamId, rowIds := range changedMetricDatasource {
		if len(rowIds) == 0 || teamId <= 0 {
			continue
		}
		teamIdTmp := teamId
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncMetricDatasource(ctx, houyi, teamIdTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) SyncMetricStrategy(ctx context.Context, changedMetricStrategy bo.ChangedMetricStrategy) error {
	houyi, ok := b.houyiRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get houyi client")
	}
	eg := new(errgroup.Group)
	for teamId, rowIds := range changedMetricStrategy {
		if len(rowIds) == 0 || teamId <= 0 {
			continue
		}
		teamIdTmp := teamId
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncMetricStrategy(ctx, houyi, teamIdTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) syncMetricDatasource(ctx context.Context, houyi repository.HouyiSyncClient, teamId uint32, rowIds []uint32) error {
	teamDo, err := b.teamRepo.FindByID(ctx, teamId)
	if err != nil {
		return merr.ErrorInternalServer("failed to get team: %v", err)
	}
	ctx = permission.WithTeamIDContext(ctx, teamId)
	datasourceDos, err := b.metricDatasourceRepo.FindByIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find metric datasource: %v", err)
	}
	datasourceItems := slices.MapFilter(datasourceDos, func(item do.DatasourceMetric) (*common.MetricDatasourceItem, bool) {
		syncItem := bo.ToSyncMetricDatasourceItem(item, teamDo)
		return syncItem, validate.IsNotNil(syncItem)
	})

	reply, err := houyi.SyncMetricDatasource(ctx, &houyiv1.MetricDatasourceRequest{Items: datasourceItems})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync metric strategy: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync metric strategy: %v", reply)
	return nil
}

func (b *Server) syncMetricStrategy(ctx context.Context, houyi repository.HouyiSyncClient, teamId uint32, rowIds []uint32) error {
	teamDo, err := b.teamRepo.FindByID(ctx, teamId)
	if err != nil {
		return merr.ErrorInternalServer("failed to get team: %v", err)
	}
	ctx = permission.WithTeamIDContext(ctx, teamId)
	strategyMetricDos, err := b.metricStrategyRepo.FindByStrategyIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find metric strategy: %v", err)
	}
	strategyItems := slices.MapFilter(strategyMetricDos, func(item do.StrategyMetric) (*common.MetricStrategyItem, bool) {
		syncItem := bo.ToSyncMetricStrategyItem(item, teamDo)
		return syncItem, validate.IsNotNil(syncItem)
	})
	reply, err := houyi.SyncMetricStrategy(ctx, &houyiv1.MetricStrategyRequest{Strategies: strategyItems})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync metric strategy: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync metric strategy: %v", reply)
	return nil
}

func (b *Server) SyncNoticeGroup(ctx context.Context, changedNoticeGroup bo.ChangedNoticeGroup) error {

	return nil
}

func (b *Server) SyncNoticeUser(ctx context.Context, changedNoticeUser bo.ChangedNoticeUser) error {

	return nil
}

func (b *Server) SyncNoticeSMSConfig(ctx context.Context, changedNoticeSMSConfig bo.ChangedNoticeSMSConfig) error {

	return nil
}

func (b *Server) SyncNoticeEmailConfig(ctx context.Context, changedNoticeEmailConfig bo.ChangedNoticeEmailConfig) error {

	return nil
}

func (b *Server) SyncNoticeHookConfig(ctx context.Context, changedNoticeHookConfig bo.ChangedNoticeHookConfig) error {

	return nil
}
