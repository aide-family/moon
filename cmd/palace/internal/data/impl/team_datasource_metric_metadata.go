package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gorm/clause"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/cmd/palace/internal/data/impl/build"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
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
	bizQuery, teamId := getTeamBizQueryWithTeamID(ctx, t.Data)
	datasourceMetricMetadataMutation := bizQuery.DatasourceMetricMetadata
	wrapper := datasourceMetricMetadataMutation.WithContext(ctx)
	wrapper = wrapper.Where(datasourceMetricMetadataMutation.DatasourceMetricID.Eq(req.DatasourceID))
	wrapper = wrapper.Where(datasourceMetricMetadataMutation.TeamID.Eq(teamId))
	if validate.TextIsNotNull(req.Keyword) {
		or := []gen.Condition{
			datasourceMetricMetadataMutation.Name.Like(req.Keyword),
			datasourceMetricMetadataMutation.Help.Like(req.Keyword),
		}
		wrapper = wrapper.Where(datasourceMetricMetadataMutation.Or(or...))
	}
	if validate.TextIsNotNull(req.Type) {
		wrapper = wrapper.Where(datasourceMetricMetadataMutation.Type.Eq(req.Type))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.WithContext(ctx).Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		wrapper = wrapper.Limit(int(req.PaginationRequest.Limit)).Offset(req.PaginationRequest.Offset())
	}
	wrapper = wrapper.Order(datasourceMetricMetadataMutation.CreatedAt.Desc())
	items, err := wrapper.Find()
	if err != nil {
		return nil, err
	}

	rows := slices.Map(items, func(item *team.DatasourceMetricMetadata) do.DatasourceMetricMetadata {
		return item
	})
	return req.ToListReply(rows), nil
}

func (t *teamDatasourceMetricMetadataImpl) UpdateRemark(ctx context.Context, req *bo.UpdateTeamMetricDatasourceMetadataRemarkRequest) error {
	return nil
}

func (t *teamDatasourceMetricMetadataImpl) Get(ctx context.Context, metadataID uint32) (do.DatasourceMetricMetadata, error) {
	datasourceMetricMetadataMutation := getTeamBizQuery(ctx, t.Data).DatasourceMetricMetadata
	wrapper := datasourceMetricMetadataMutation.WithContext(ctx)
	wrapper = wrapper.Where(datasourceMetricMetadataMutation.ID.Eq(metadataID))
	item, err := wrapper.First()
	if err != nil {
		return nil, teamDatasourceMetricMetadataNotFound(err)
	}
	return item, nil
}

func (t *teamDatasourceMetricMetadataImpl) Delete(ctx context.Context, metadataID uint32) error {
	datasourceMetricMetadataMutation := getTeamBizQuery(ctx, t.Data).DatasourceMetricMetadata
	wrapper := datasourceMetricMetadataMutation.WithContext(ctx)
	wrapper = wrapper.Where(datasourceMetricMetadataMutation.ID.Eq(metadataID))
	_, err := wrapper.Delete()
	return err
}
