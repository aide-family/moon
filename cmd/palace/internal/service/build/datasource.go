package build

import (
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
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

func ToTeamMetricDatasourceItem(item do.DatasourceMetric) *palacecommon.TeamMetricDatasourceItem {
	if validate.IsNil(item) {
		return nil
	}
	return &palacecommon.TeamMetricDatasourceItem{
		TeamId:         item.GetTeamID(),
		DatasourceId:   item.GetID(),
		CreatedAt:      timex.Format(item.GetCreatedAt()),
		UpdatedAt:      timex.Format(item.GetUpdatedAt()),
		Name:           item.GetName(),
		Remark:         item.GetRemark(),
		Driver:         palacecommon.DatasourceDriverMetric(item.GetDriver()),
		Endpoint:       item.GetEndpoint(),
		ScrapeInterval: durationpb.New(item.GetScrapeInterval()),
		Headers:        item.GetHeaders(),
		QueryMethod:    palacecommon.HTTPMethod(item.GetQueryMethod()),
		Ca:             item.GetCA(),
		Tls:            ToTLSItem(item.GetTLS()),
		BasicAuth:      ToBasicAuthItem(item.GetBasicAuth()),
		Extra:          item.GetExtra(),
		Status:         palacecommon.GlobalStatus(item.GetStatus().GetValue()),
		Creator:        ToUserBaseItem(item.GetCreator()),
	}
}

func ToTeamMetricDatasourceItems(items []do.DatasourceMetric) []*palacecommon.TeamMetricDatasourceItem {
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

func ToMetricDatasourceMetadataItems(datasourceID uint32, items []*common.MetricItem) []*bo.DatasourceMetricMetadata {
	if len(items) == 0 {
		return nil
	}
	return slices.MapFilter(items, func(item *common.MetricItem) (*bo.DatasourceMetricMetadata, bool) {
		if validate.IsNil(item) {
			return nil, false
		}
		return ToMetricDatasourceMetadataItem(datasourceID, item), true
	})
}

func ToMetricDatasourceMetadataItem(datasourceID uint32, item *common.MetricItem) *bo.DatasourceMetricMetadata {
	if validate.IsNil(item) {
		return nil
	}
	labels := make(map[string][]string)
	for _, label := range item.GetLabels() {
		labels[label.GetKey()] = label.GetValues()
	}
	return &bo.DatasourceMetricMetadata{
		Name:         item.GetName(),
		Help:         item.GetHelp(),
		Type:         item.GetType(),
		Labels:       labels,
		Unit:         item.GetUnit(),
		DatasourceID: datasourceID,
	}
}

func ToTeamMetricDatasourceMetadataItem(item do.DatasourceMetricMetadata) *palacecommon.TeamMetricDatasourceMetadataItem {
	if validate.IsNil(item) {
		return nil
	}
	labels := make([]*palacecommon.TeamMetricDatasourceMetadataItem_Label, 0, len(item.GetLabels()))
	for k, v := range item.GetLabels() {
		labels = append(labels, &palacecommon.TeamMetricDatasourceMetadataItem_Label{
			Key:    k,
			Values: v,
		})
	}
	return &palacecommon.TeamMetricDatasourceMetadataItem{
		MetadataId: item.GetID(),
		Name:       item.GetName(),
		Help:       item.GetHelp(),
		Type:       item.GetType(),
		Labels:     labels,
		Unit:       item.GetUnit(),
	}
}

func ToTeamMetricDatasourceMetadataItems(items []do.DatasourceMetricMetadata) []*palacecommon.TeamMetricDatasourceMetadataItem {
	return slices.Map(items, ToTeamMetricDatasourceMetadataItem)
}

func ToListMetricDatasourceMetadataRequest(req *palace.ListMetricDatasourceMetadataRequest) *bo.ListTeamMetricDatasourceMetadata {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.ListTeamMetricDatasourceMetadata{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		DatasourceID:      req.GetDatasourceId(),
		Keyword:           req.GetKeyword(),
		Type:              req.GetType(),
	}
}

func ToUpdateMetricDatasourceMetadataRequest(req *palace.UpdateMetricDatasourceMetadataRequest) *bo.UpdateMetricDatasourceMetadataRequest {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.UpdateMetricDatasourceMetadataRequest{
		DatasourceID: req.GetDatasourceId(),
		MetadataID:   req.GetMetadataId(),
		Help:         req.GetHelp(),
		Unit:         req.GetUnit(),
		Type:         req.GetType(),
	}
}
