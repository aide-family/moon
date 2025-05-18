package build

import (
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/api/palace"
	palacecommon "github.com/aide-family/moon/pkg/api/palace/common"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

func ToKV(keyVal *palacecommon.KeyValueItem) *kv.KV {
	if validate.IsNil(keyVal) {
		return nil
	}
	return &kv.KV{
		Key:   keyVal.Key,
		Value: keyVal.Value,
	}
}

func ToKVs(kvs []*palacecommon.KeyValueItem) []*kv.KV {
	if len(kvs) == 0 {
		return nil
	}
	return slices.Map(kvs, ToKV)
}

func ToKVsItems(kvs []*kv.KV) []*palacecommon.KeyValueItem {
	if len(kvs) == 0 {
		return nil
	}
	return slices.Map(kvs, ToKVItem)
}

func ToKVItem(kv *kv.KV) *palacecommon.KeyValueItem {
	if validate.IsNil(kv) {
		return nil
	}
	return &palacecommon.KeyValueItem{
		Key:   kv.Key,
		Value: kv.Value,
	}
}

func ToSaveTeamMetricDatasourceRequest(req *palace.SaveTeamMetricDatasourceRequest) *bo.SaveTeamMetricDatasource {
	if validate.IsNil(req) {
		return nil
	}

	return &bo.SaveTeamMetricDatasource{
		ID:             req.GetDatasourceId(),
		Name:           req.GetName(),
		Remark:         req.GetRemark(),
		Driver:         vobj.DatasourceDriverMetric(req.GetDriver()),
		Endpoint:       req.GetEndpoint(),
		ScrapeInterval: time.Duration(req.GetScrapeInterval()) * time.Second,
		Headers:        ToKVs(req.GetHeaders()),
		QueryMethod:    vobj.HTTPMethod(req.GetQueryMethod()),
		CA:             req.GetCa(),
		TLS:            ToTLS(req.GetTls()),
		BasicAuth:      ToBasicAuth(req.GetBasicAuth()),
		Extra:          ToKVs(req.GetExtra()),
	}
}

func ToDatasourceSelectRequest(req *palace.DatasourceSelectRequest) *bo.DatasourceSelect {
	if validate.IsNil(req) {
		return nil
	}
	return &bo.DatasourceSelect{
		PaginationRequest: ToPaginationRequest(req.GetPagination()),
		Status:            vobj.GlobalStatus(req.GetStatus()),
		Type:              vobj.DatasourceType(req.GetType()),
		Keyword:           req.GetKeyword(),
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
		ScrapeInterval: int64(item.GetScrapeInterval().Seconds()),
		Headers:        ToKVsItems(item.GetHeaders()),
		QueryMethod:    palacecommon.HTTPMethod(item.GetQueryMethod()),
		Ca:             item.GetCA(),
		Tls:            ToTLSItem(item.GetTLS()),
		BasicAuth:      ToBasicAuthItem(item.GetBasicAuth()),
		Extra:          ToKVsItems(item.GetExtra()),
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
