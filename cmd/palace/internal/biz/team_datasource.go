package biz

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/util/validate"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
	"github.com/aide-family/moon/pkg/merr"
)

func NewTeamDatasourceBiz(
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

func (t *TeamDatasource) DatasourceSelect(ctx context.Context, req *bo.DatasourceSelect) (*bo.DatasourceSelectReply, error) {
	switch req.Type {
	case vobj.DatasourceTypeMetric:
		return t.teamDatasourceMetricRepo.Select(ctx, req)
	default:
		return nil, merr.ErrorBadRequest("unsupported datasource type")
	}
}

func (t *TeamDatasource) SaveMetricDatasource(ctx context.Context, req *bo.SaveTeamMetricDatasource) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if req.ID <= 0 {
		return t.teamDatasourceMetricRepo.Create(ctx, req)
	}
	metricDatasourceDo, err := t.teamDatasourceMetricRepo.Get(ctx, req.ID)
	if err != nil {
		return err
	}
	if metricDatasourceDo.GetStatus().IsEnable() {
		return merr.ErrorBadRequest("datasource is enabled and cannot be modified")
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

func (t *TeamDatasource) SyncMetricMetadata(ctx context.Context, req *bo.SyncMetricMetadataRequest) (err error) {
	syncClient, ok := t.houyiRepo.Sync()
	if !ok {
		return merr.ErrorBadRequest("sync service is not running")
	}

	datasourceMetricDo, err := t.teamDatasourceMetricRepo.Get(ctx, req.DatasourceID)
	if err != nil {
		return err
	}
	teamDo := datasourceMetricDo.GetTeam()
	key := repository.TeamDatasourceMetricMetadataSyncKey.Key(teamDo.GetID(), req.DatasourceID)
	locked, err := t.cacheRepo.Lock(ctx, key, 5*time.Minute)
	if err != nil {
		t.helper.WithContext(ctx).Warnw("msg", "lock team datasource metric metadata sync key error", "err", err)
		return err
	}
	if !locked {
		return merr.ErrorBadRequest("datasource metadata sync is in progress, please try again later")
	}
	defer func() {
		if validate.IsNotNil(err) {
			_ = t.cacheRepo.Unlock(ctx, key)
		}
	}()
	reply, err := syncClient.SyncMetricMetadata(ctx, &houyiv1.MetricMetadataRequest{
		Item:       bo.ToSyncMetricDatasourceItem(datasourceMetricDo),
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

func (t *TeamDatasource) UpdateMetricDatasourceMetadata(ctx context.Context, req *bo.UpdateMetricDatasourceMetadataRequest) error {
	return t.teamDatasourceMetricMetadataRepo.Update(ctx, req)
}

func (t *TeamDatasource) DeleteMetricDatasourceMetadata(ctx context.Context, metadataID uint32) error {
	return t.teamDatasourceMetricMetadataRepo.Delete(ctx, metadataID)
}
