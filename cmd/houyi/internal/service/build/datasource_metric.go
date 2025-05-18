package build

import (
	"github.com/aide-family/moon/cmd/houyi/internal/biz/do"
	"github.com/aide-family/moon/pkg/api/houyi/common"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/slices"
)

func ToMetricDatasourceConfig(metricItem *common.MetricDatasourceItem) (*do.DatasourceMetricConfig, error) {
	config := metricItem.GetConfig()
	if config == nil {
		return nil, merr.ErrorParamsError("config is nil")
	}
	return &do.DatasourceMetricConfig{
		TeamId:   metricItem.GetTeam().GetTeamId(),
		ID:       metricItem.GetId(),
		Name:     metricItem.GetName(),
		Driver:   metricItem.GetDriver(),
		Endpoint: config.GetEndpoint(),
		Headers: slices.Map(config.GetHeaders(), func(header *common.KeyValueItem) *kv.KV {
			return &kv.KV{
				Key:   header.GetKey(),
				Value: header.GetValue(),
			}
		}),
		Method:         config.GetMethod(),
		CA:             config.GetCa(),
		BasicAuth:      ToBasicAuth(config.GetBasicAuth()),
		TLS:            ToTLS(config.GetTls()),
		Enable:         metricItem.GetEnable(),
		ScrapeInterval: metricItem.GetScrapeInterval().AsDuration(),
	}, nil
}
