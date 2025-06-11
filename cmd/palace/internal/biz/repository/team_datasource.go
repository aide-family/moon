package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
)

type TeamDatasourceMetric interface {
	Create(ctx context.Context, req *bo.SaveTeamMetricDatasource) error
	Update(ctx context.Context, req *bo.SaveTeamMetricDatasource) error
	UpdateStatus(ctx context.Context, req *bo.UpdateTeamMetricDatasourceStatusRequest) error
	Delete(ctx context.Context, datasourceID uint32) error
	Get(ctx context.Context, datasourceID uint32) (do.DatasourceMetric, error)
	List(ctx context.Context, req *bo.ListTeamMetricDatasource) (*bo.ListTeamMetricDatasourceReply, error)
	FindByIds(ctx context.Context, datasourceIds []uint32) ([]do.DatasourceMetric, error)
}

type TeamDatasourceMetricMetadata interface {
	BatchSave(ctx context.Context, req *bo.BatchSaveTeamMetricDatasourceMetadata) error
	List(ctx context.Context, req *bo.ListTeamMetricDatasourceMetadata) (*bo.ListTeamMetricDatasourceMetadataReply, error)
	UpdateRemark(ctx context.Context, req *bo.UpdateTeamMetricDatasourceMetadataRemarkRequest) error
	Get(ctx context.Context, metadataID uint32) (do.DatasourceMetricMetadata, error)
	Delete(ctx context.Context, metadataID uint32) error
}
