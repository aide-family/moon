package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/pkg/util/crypto"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToDatasourceMetric(ctx context.Context, datasource do.DatasourceMetric) *team.DatasourceMetric {
	if validate.IsNil(datasource) {
		return nil
	}
	if datasource, ok := datasource.(*team.DatasourceMetric); ok {
		datasource.WithContext(ctx)
		return datasource
	}
	return &team.DatasourceMetric{
		TeamModel:      ToTeamModel(ctx, datasource),
		Name:           datasource.GetName(),
		Status:         datasource.GetStatus(),
		Remark:         datasource.GetRemark(),
		Driver:         datasource.GetDriver(),
		Endpoint:       datasource.GetEndpoint(),
		ScrapeInterval: int64(datasource.GetScrapeInterval().Seconds()),
		Headers:        crypto.NewObject(datasource.GetHeaders()),
		QueryMethod:    datasource.GetQueryMethod(),
		CA:             crypto.String(datasource.GetCA()),
		TLS:            crypto.NewObject(datasource.GetTLS()),
		BasicAuth:      crypto.NewObject(datasource.GetBasicAuth()),
		Extra:          crypto.NewObject(datasource.GetExtra()),
		Metrics:        []*team.StrategyMetric{},
	}
}
func ToDatasourceMetrics(ctx context.Context, datasourceList []do.DatasourceMetric) []*team.DatasourceMetric {
	return slices.MapFilter(datasourceList, func(v do.DatasourceMetric) (*team.DatasourceMetric, bool) {
		if validate.IsNil(v) {
			return nil, false
		}
		return ToDatasourceMetric(ctx, v), true
	})
}

func ToDatasourceMetricMetadata(ctx context.Context, metadata *bo.DatasourceMetricMetadata) *team.DatasourceMetricMetadata {
	if validate.IsNil(metadata) {
		return nil
	}
	item := &team.DatasourceMetricMetadata{
		Name:               metadata.Name,
		Help:               metadata.Help,
		Type:               metadata.Type,
		Labels:             metadata.Labels,
		Unit:               metadata.Unit,
		DatasourceMetricID: metadata.DatasourceID,
	}
	item.WithContext(ctx)
	return item
}

func ToDatasourceMetricMetadataList(ctx context.Context, metadataList []*bo.DatasourceMetricMetadata) []*team.DatasourceMetricMetadata {
	return slices.MapFilter(metadataList, func(v *bo.DatasourceMetricMetadata) (*team.DatasourceMetricMetadata, bool) {
		if validate.IsNil(v) {
			return nil, false
		}
		return ToDatasourceMetricMetadata(ctx, v), true
	})
}
