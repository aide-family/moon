package impl

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm/clause"
)

func NewTeamDatasourceMetricMetadataRepo(data *data.Data, logger log.Logger) repository.TeamDatasourceMetricMetadata {
	return &teamDatasourceMetricMetadataImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.impl.team_datasource_metric_metadata")),
	}
}

type teamDatasourceMetricMetadataImpl struct {
	*data.Data
	helper *log.Helper
}

func (t *teamDatasourceMetricMetadataImpl) BatchSave(ctx context.Context, req *bo.BatchSaveTeamMetricDatasourceMetadata) error {
	metadataDos := build.ToDatasourceMetricMetadataList(ctx, req.Metadata)
	datasourceMetricMetadataMutation := getTeamBizQuery(ctx, t.Data).DatasourceMetricMetadata
	return datasourceMetricMetadataMutation.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: datasourceMetricMetadataMutation.DatasourceMetricID.ColumnName().String()},
			{Name: datasourceMetricMetadataMutation.Name.ColumnName().String()},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			datasourceMetricMetadataMutation.Type.ColumnName().String(),
			datasourceMetricMetadataMutation.Labels.ColumnName().String(),
		}),
	}).CreateInBatches(metadataDos, len(metadataDos))
}

func (t *teamDatasourceMetricMetadataImpl) List(ctx context.Context, req *bo.ListTeamMetricDatasourceMetadata) (*bo.ListTeamMetricDatasourceMetadataReply, error) {
	return nil, nil
}

func (t *teamDatasourceMetricMetadataImpl) UpdateRemark(ctx context.Context, req *bo.UpdateTeamMetricDatasourceMetadataRemarkRequest) error {
	return nil
}

func (t *teamDatasourceMetricMetadataImpl) Get(ctx context.Context, metadataID uint32) (do.DatasourceMetricMetadata, error) {
	return nil, nil
}

func (t *teamDatasourceMetricMetadataImpl) Delete(ctx context.Context, metadataID uint32) error {
	return nil
}
