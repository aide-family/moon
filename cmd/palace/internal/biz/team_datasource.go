package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
	"github.com/aide-family/moon/pkg/merr"
)

func NewTeamDatasource(
	teamDatasourceMetricRepo repository.TeamDatasourceMetric,
	teamDatasourceMetricMetadataRepo repository.TeamDatasourceMetricMetadata,
	cacheRepo repository.Cache,
	houyiRepo repository.Houyi,
	logger log.Logger,
) *TeamDatasource {
	return &TeamDatasource{
		teamDatasourceMetricRepo:         teamDatasourceMetricRepo,
		teamDatasourceMetricMetadataRepo: teamDatasourceMetricMetadataRepo,
		cacheRepo:                        cacheRepo,
		houyiRepo:                        houyiRepo,
		helper:                           log.NewHelper(log.With(logger, "module", "biz.team_datasource")),
	}
}

type TeamDatasource struct {
	teamDatasourceMetricRepo         repository.TeamDatasourceMetric
	teamDatasourceMetricMetadataRepo repository.TeamDatasourceMetricMetadata
	cacheRepo                        repository.Cache
	houyiRepo                        repository.Houyi
	helper                           *log.Helper
}

func (t *TeamDatasource) SaveMetricDatasource(ctx context.Context, req *bo.SaveTeamMetricDatasource) error {
	if req.ID <= 0 {
		return t.teamDatasourceMetricRepo.Create(ctx, req)
	}
	metricDatasourceDo, err := t.teamDatasourceMetricRepo.Get(ctx, req.ID)
	if err != nil {
		return err
	}
	if metricDatasourceDo.GetStatus().IsEnable() {
		return merr.ErrorBadRequest("数据源已启用，不能修改")
	}

	return t.teamDatasourceMetricRepo.Update(ctx, req)
}

func (t *TeamDatasource) UpdateMetricDatasourceStatus(ctx context.Context, req *bo.UpdateTeamMetricDatasourceStatusRequest) error {
	return t.teamDatasourceMetricRepo.UpdateStatus(ctx, req)
}

func (t *TeamDatasource) DeleteMetricDatasource(ctx context.Context, datasourceID uint32) error {
	return t.teamDatasourceMetricRepo.Delete(ctx, datasourceID)
}

func (t *TeamDatasource) GetMetricDatasource(ctx context.Context, datasourceID uint32) (do.DatasourceMetric, error) {
	return t.teamDatasourceMetricRepo.Get(ctx, datasourceID)
}

func (t *TeamDatasource) ListMetricDatasource(ctx context.Context, req *bo.ListTeamMetricDatasource) (*bo.ListTeamMetricDatasourceReply, error) {
	return t.teamDatasourceMetricRepo.List(ctx, req)
}

func (t *TeamDatasource) BatchSaveMetricDatasourceMetadata(ctx context.Context, req *bo.BatchSaveTeamMetricDatasourceMetadata) error {
	if req.IsDone {
		err := t.cacheRepo.Unlock(ctx, repository.TeamDatasourceMetricMetadataSyncKey.Key(req.TeamID, req.DatasourceID))
		if err != nil {
			t.helper.WithContext(ctx).Warnw("msg", "unlock team datasource metric metadata sync key error", "err", err)
		}
	}
	if len(req.Metadata) == 0 {
		return nil
	}
	return t.teamDatasourceMetricMetadataRepo.BatchSave(ctx, req)
}

func (t *TeamDatasource) SyncMetricMetadata(ctx context.Context, req *bo.SyncMetricMetadataRequest) error {
	syncClient, ok := t.houyiRepo.Sync()
	if !ok {
		return merr.ErrorBadRequest("同步服务未启动")
	}

	datasourceMetricDo, err := t.teamDatasourceMetricRepo.Get(ctx, req.DatasourceID)
	if err != nil {
		return err
	}
	teamDo := datasourceMetricDo.GetTeam()
	locked, err := t.cacheRepo.Lock(ctx, repository.TeamDatasourceMetricMetadataSyncKey.Key(teamDo.GetID(), req.DatasourceID), 5*time.Minute)
	if err != nil {
		t.helper.WithContext(ctx).Warnw("msg", "lock team datasource metric metadata sync key error", "err", err)
		return err
	}
	if !locked {
		return merr.ErrorBadRequest("数据源元数据同步正在执行中，请稍后再试")
	}
	reply, err := syncClient.MetricMetadata(ctx, &houyiv1.MetricMetadataRequest{
		Item:       NewMetricDatasourceItem(datasourceMetricDo),
		OperatorId: permission.GetUserIDByContextWithDefault(ctx, teamDo.GetCreatorID()),
	})
	if err != nil {
		return err
	}
	t.helper.WithContext(ctx).Infof("sync metric metadata reply: %v", reply)

	return nil
}

func (t *TeamDatasource) ListMetricDatasourceMetadata(ctx context.Context, req *bo.ListTeamMetricDatasourceMetadata) (*bo.ListTeamMetricDatasourceMetadataReply, error) {
	return t.teamDatasourceMetricMetadataRepo.List(ctx, req)
}

func (t *TeamDatasource) GetMetricDatasourceMetadata(ctx context.Context, req *bo.GetMetricDatasourceMetadataRequest) (do.DatasourceMetricMetadata, error) {
	return t.teamDatasourceMetricMetadataRepo.Get(ctx, req.ID)
}

func (t *TeamDatasource) UpdateMetricDatasourceMetadataRemark(ctx context.Context, req *bo.UpdateTeamMetricDatasourceMetadataRemarkRequest) error {
	return t.teamDatasourceMetricMetadataRepo.UpdateRemark(ctx, req)
}

func (t *TeamDatasource) DeleteMetricDatasourceMetadata(ctx context.Context, metadataID uint32) error {
	return t.teamDatasourceMetricMetadataRepo.Delete(ctx, metadataID)
}
