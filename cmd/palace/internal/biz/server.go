package biz

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"golang.org/x/sync/errgroup"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
	rabbitv1 "github.com/aide-family/moon/pkg/api/rabbit/v1"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/middler/permission"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewServerBiz(
	serverRepo repository.Server,
	houyiRepo repository.Houyi,
	rabbitRepo repository.Rabbit,
	metricDatasourceRepo repository.TeamDatasourceMetric,
	metricStrategyRepo repository.TeamStrategyMetric,
	noticeGroupRepo repository.TeamNotice,
	teamRepo repository.Team,
	teamSMSConfigRepo repository.TeamSMSConfig,
	teamEmailConfigRepo repository.TeamEmailConfig,
	teamHookConfigRepo repository.TeamHook,
	logger log.Logger,
) *Server {
	return &Server{
		serverRepo:           serverRepo,
		houyiRepo:            houyiRepo,
		rabbitRepo:           rabbitRepo,
		metricDatasourceRepo: metricDatasourceRepo,
		metricStrategyRepo:   metricStrategyRepo,
		noticeGroupRepo:      noticeGroupRepo,
		teamRepo:             teamRepo,
		teamSMSConfigRepo:    teamSMSConfigRepo,
		teamEmailConfigRepo:  teamEmailConfigRepo,
		teamHookConfigRepo:   teamHookConfigRepo,
		helper:               log.NewHelper(log.With(logger, "module", "biz.server")),
	}
}

type Server struct {
	serverRepo           repository.Server
	houyiRepo            repository.Houyi
	rabbitRepo           repository.Rabbit
	metricDatasourceRepo repository.TeamDatasourceMetric
	metricStrategyRepo   repository.TeamStrategyMetric
	noticeGroupRepo      repository.TeamNotice
	teamRepo             repository.Team
	teamSMSConfigRepo    repository.TeamSMSConfig
	teamEmailConfigRepo  repository.TeamEmailConfig
	teamHookConfigRepo   repository.TeamHook
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

	b.helper.WithContext(ctx).Debugf("registered server type: %v, uuid: %s", req.ServerType, req.UUID)
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
	b.helper.WithContext(ctx).Debugf("deregistered server type: %v, uuid: %s", req.ServerType, req.UUID)
	return nil
}

func (b *Server) SyncMetricDatasource(ctx context.Context, changedMetricDatasource bo.ChangedMetricDatasource) error {
	houyi, ok := b.houyiRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get houyi client")
	}
	eg := new(errgroup.Group)
	for teamID, rowIds := range changedMetricDatasource {
		if len(rowIds) == 0 || teamID <= 0 {
			continue
		}
		teamIDTmp := teamID
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncMetricDatasource(ctx, houyi, teamIDTmp, rowIdsTmp)
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
	for teamID, rowIds := range changedMetricStrategy {
		if len(rowIds) == 0 || teamID <= 0 {
			continue
		}
		teamIDTmp := teamID
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncMetricStrategy(ctx, houyi, teamIDTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) syncMetricDatasource(ctx context.Context, houyi repository.HouyiSyncClient, teamID uint32, rowIds []uint32) error {
	ctx = permission.WithTeamIDContext(ctx, teamID)
	ctx = metadata.AppendToClientContext(ctx, cnst.MetadataGlobalKeyTeamID, strconv.FormatUint(uint64(teamID), 10))
	datasourceDos, err := b.metricDatasourceRepo.FindByIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find metric datasource: %v", err)
	}
	datasourceItems := slices.MapFilter(datasourceDos, func(item do.DatasourceMetric) (*common.MetricDatasourceItem, bool) {
		syncItem := bo.ToSyncMetricDatasourceItem(item)
		return syncItem, validate.IsNotNil(syncItem)
	})

	reply, err := houyi.SyncMetricDatasource(ctx, &houyiv1.MetricDatasourceRequest{Items: datasourceItems})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync metric strategy: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync metric strategy: %v", reply)
	return nil
}

func (b *Server) syncMetricStrategy(ctx context.Context, houyi repository.HouyiSyncClient, teamID uint32, rowIds []uint32) error {
	ctx = metadata.AppendToClientContext(ctx, cnst.MetadataGlobalKeyTeamID, strconv.FormatUint(uint64(teamID), 10))
	ctx = permission.WithTeamIDContext(ctx, teamID)
	strategyMetricDos, err := b.metricStrategyRepo.FindByStrategyIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find metric strategy: %v", err)
	}
	strategyItems := slices.MapFilter(strategyMetricDos, func(item do.StrategyMetric) (*common.MetricStrategyItem, bool) {
		syncItem := bo.ToSyncMetricStrategyItem(item)
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
	rabbit, ok := b.rabbitRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get rabbit client")
	}
	eg := new(errgroup.Group)
	for teamID, rowIds := range changedNoticeGroup {
		if len(rowIds) == 0 || teamID <= 0 {
			continue
		}
		teamIDTmp := teamID
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncNoticeGroup(ctx, rabbit, teamIDTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) syncNoticeGroup(ctx context.Context, rabbit repository.RabbitSyncClient, teamID uint32, rowIds []uint32) error {
	ctx = permission.WithTeamIDContext(ctx, teamID)
	ctx = metadata.AppendToClientContext(ctx, cnst.MetadataGlobalKeyTeamID, strconv.FormatUint(uint64(teamID), 10))
	groupDos, err := b.noticeGroupRepo.FindByIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find notice group: %v", err)
	}
	if len(groupDos) == 0 {
		return nil
	}
	reply, err := rabbit.NoticeGroup(ctx, &rabbitv1.SyncNoticeGroupRequest{
		NoticeGroups: bo.ToSyncNoticeGroupItems(groupDos),
	})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync notice group: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync notice group: %v", reply)
	return nil
}

func (b *Server) SyncNoticeSMSConfig(ctx context.Context, changedNoticeSMSConfig bo.ChangedNoticeSMSConfig) error {
	rabbit, ok := b.rabbitRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get rabbit client")
	}
	eg := new(errgroup.Group)
	for teamID, rowIds := range changedNoticeSMSConfig {
		if len(rowIds) == 0 || teamID <= 0 {
			continue
		}
		teamIDTmp := teamID
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncNoticeSMSConfig(ctx, rabbit, teamIDTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) syncNoticeSMSConfig(ctx context.Context, rabbit repository.RabbitSyncClient, teamID uint32, rowIds []uint32) error {
	ctx = permission.WithTeamIDContext(ctx, teamID)
	ctx = metadata.AppendToClientContext(ctx, cnst.MetadataGlobalKeyTeamID, strconv.FormatUint(uint64(teamID), 10))
	smsDos, err := b.teamSMSConfigRepo.FindByIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find notice sms config: %v", err)
	}
	if len(smsDos) == 0 {
		return nil
	}
	reply, err := rabbit.Sms(ctx, &rabbitv1.SyncSmsRequest{
		Smss: bo.ToSyncSMSConfigItems(smsDos),
	})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync notice sms config: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync notice sms config: %v", reply)
	return nil
}

func (b *Server) SyncNoticeEmailConfig(ctx context.Context, changedNoticeEmailConfig bo.ChangedNoticeEmailConfig) error {
	rabbit, ok := b.rabbitRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get rabbit client")
	}
	eg := new(errgroup.Group)
	for teamID, rowIds := range changedNoticeEmailConfig {
		if len(rowIds) == 0 || teamID <= 0 {
			continue
		}
		teamIDTmp := teamID
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncNoticeEmailConfig(ctx, rabbit, teamIDTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) syncNoticeEmailConfig(ctx context.Context, rabbit repository.RabbitSyncClient, teamID uint32, rowIds []uint32) error {
	ctx = permission.WithTeamIDContext(ctx, teamID)
	ctx = metadata.AppendToClientContext(ctx, cnst.MetadataGlobalKeyTeamID, strconv.FormatUint(uint64(teamID), 10))
	emailDos, err := b.teamEmailConfigRepo.FindByIds(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find notice email config: %v", err)
	}
	if len(emailDos) == 0 {
		return nil
	}
	reply, err := rabbit.Email(ctx, &rabbitv1.SyncEmailRequest{
		Emails: bo.ToSyncEmailConfigItems(emailDos),
	})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync notice email config: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync notice email config: %v", reply)
	return nil
}

func (b *Server) SyncNoticeHookConfig(ctx context.Context, changedNoticeHookConfig bo.ChangedNoticeHookConfig) error {
	rabbit, ok := b.rabbitRepo.Sync()
	if !ok {
		return merr.ErrorInternalServer("failed to get rabbit client")
	}
	eg := new(errgroup.Group)
	for teamID, rowIds := range changedNoticeHookConfig {
		if len(rowIds) == 0 || teamID <= 0 {
			continue
		}
		teamIDTmp := teamID
		rowIdsTmp := rowIds
		eg.Go(func() error {
			return b.syncNoticeHookConfig(ctx, rabbit, teamIDTmp, rowIdsTmp)
		})
	}
	return eg.Wait()
}

func (b *Server) syncNoticeHookConfig(ctx context.Context, rabbit repository.RabbitSyncClient, teamID uint32, rowIds []uint32) error {
	ctx = permission.WithTeamIDContext(ctx, teamID)
	ctx = metadata.AppendToClientContext(ctx, cnst.MetadataGlobalKeyTeamID, strconv.FormatUint(uint64(teamID), 10))
	hookDos, err := b.teamHookConfigRepo.Find(ctx, rowIds)
	if err != nil {
		return merr.ErrorInternalServer("failed to find notice hook config: %v", err)
	}
	if len(hookDos) == 0 {
		return nil
	}
	reply, err := rabbit.Hook(ctx, &rabbitv1.SyncHookRequest{
		Hooks: bo.ToSyncHookConfigItems(hookDos),
	})
	if err != nil {
		return merr.ErrorInternalServer("failed to sync notice hook config: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("sync notice hook config: %v", reply)
	return nil
}
