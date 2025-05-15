package build

import (
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	com "github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/api/palace/common"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/plugin/datasource"
	"github.com/moon-monitor/moon/pkg/plugin/datasource/prometheus"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/timex"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToSaveTeamMetricDatasourceRequest(req *palace.SaveTeamMetricDatasourceRequest) *bo.SaveTeamMetricDatasource {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.SaveTeamMetricDatasource{
		ID:             req.GetDatasourceId(),
		Name:           req.GetName(),
		Status:         vobj.GlobalStatusEnable,
		Remark:         req.GetRemark(),
		Driver:         vobj.DatasourceDriverMetric(req.GetMetricDatasourceDriver()),
		Endpoint:       req.GetEndpoint(),
		ScrapeInterval: req.GetScrapeInterval().AsDuration(),
		Headers:        req.GetHeaders(),
		QueryMethod:    vobj.HTTPMethod(req.GetQueryMethod()),
		CA:             req.GetCa(),
		TLS:            ToTLS(req.GetTls()),
		BasicAuth:      ToBasicAuth(req.GetBasicAuth()),
		Extra:          req.GetExtra(),
	}
}

func ToListTeamMetricDatasourceRequest(req *palace.ListTeamMetricDatasourceRequest) *bo.ListTeamMetricDatasource {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.ListTeamMetricDatasource{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Keyword:           req.GetKeyword(),
	}
}

func ToTeamMetricDatasourceItem(item do.DatasourceMetric) *common.TeamMetricDatasourceItem {
	if validate.IsNil(item) {
		return nil
	}
	return &common.TeamMetricDatasourceItem{
		TeamId:         item.GetTeamID(),
		DatasourceId:   item.GetID(),
		CreatedAt:      timex.Format(item.GetCreatedAt()),
		UpdatedAt:      timex.Format(item.GetUpdatedAt()),
		Name:           item.GetName(),
		Remark:         item.GetRemark(),
		Driver:         common.DatasourceDriverMetric(item.GetDriver()),
		Endpoint:       item.GetEndpoint(),
		ScrapeInterval: durationpb.New(item.GetScrapeInterval()),
		Headers:        item.GetHeaders(),
		QueryMethod:    common.HTTPMethod(item.GetQueryMethod()),
		Ca:             item.GetCA(),
		Tls:            ToTLSItem(item.GetTLS()),
		BasicAuth:      ToBasicAuthItem(item.GetBasicAuth()),
		Extra:          item.GetExtra(),
		Status:         common.GlobalStatus(item.GetStatus().GetValue()),
		Creator:        ToUserBaseItem(item.GetCreator()),
	}
}

func ToTeamMetricDatasourceItems(items []do.DatasourceMetric) []*common.TeamMetricDatasourceItem {
	return slices.Map(items, ToTeamMetricDatasourceItem)
}

func ToBatchSaveTeamMetricDatasourceMetadataRequest(req *palace.SyncMetadataRequest) *bo.BatchSaveTeamMetricDatasourceMetadata {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.BatchSaveTeamMetricDatasourceMetadata{
		TeamID:       req.GetTeamId(),
		DatasourceID: req.GetDatasourceId(),
		Metadata:     ToMetricDatasourceMetadataItems(req.GetDatasourceId(), req.GetItems()),
		IsDone:       req.GetIsDone(),
	}
}

func ToMetricDatasourceMetadataItems(datasourceID uint32, items []*com.MetricItem) []*bo.DatasourceMetricMetadata {
	if len(items) == 0 {
		return nil
	}
	return slices.MapFilter(items, func(item *com.MetricItem) (*bo.DatasourceMetricMetadata, bool) {
		if validate.IsNil(item) {
			return nil, false
		}
		return ToMetricDatasourceMetadataItem(datasourceID, item), true
	})
}

func ToMetricDatasourceMetadataItem(datasourceID uint32, item *com.MetricItem) *bo.DatasourceMetricMetadata {
	if validate.IsNil(item) {
		return nil
	}
	return &bo.DatasourceMetricMetadata{
		Name:         item.GetName(),
		Help:         item.GetHelp(),
		Type:         item.GetType(),
		Labels:       item.GetLabels(),
		Unit:         item.GetUnit(),
		DatasourceID: datasourceID,
	}
}

func ToMetricDatasource(datasourceMetric do.DatasourceMetric, logger log.Logger) (datasource.Metric, error) {
	switch datasourceMetric.GetDriver() {
	case vobj.DatasourceDriverMetricPrometheus:
		return prometheus.New(bo.NewMetricDatasourceConfig(datasourceMetric), logger), nil
	default:
		return nil, merr.ErrorBadRequest("invalid datasource driver: %s", datasourceMetric.GetDriver())
	}
}
