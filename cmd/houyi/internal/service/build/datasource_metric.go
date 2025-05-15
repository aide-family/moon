package build

import (
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/do"
	"github.com/moon-monitor/moon/pkg/api/houyi/common"
	"github.com/moon-monitor/moon/pkg/merr"
)

func ToMetricDatasourceConfig(metricItem *common.MetricDatasourceItem) (*do.DatasourceMetricConfig, error) {
	switch metricItem.GetDriver() {
	case common.MetricDatasourceDriver_PROMETHEUS:
		return ToMetricDatasourceConfigWithPrometheus(metricItem)
	case common.MetricDatasourceDriver_VICTORIAMETRICS:
		return ToMetricDatasourceConfigWithVictoriaMetrics(metricItem)
	default:
		return nil, merr.ErrorParamsError("invalid metric datasource driver: %s", metricItem.GetDriver())
	}
}

func ToMetricDatasourceConfigWithPrometheus(metricItem *common.MetricDatasourceItem) (*do.DatasourceMetricConfig, error) {
	prometheusConfig := metricItem.GetPrometheus()
	if prometheusConfig == nil {
		return nil, merr.ErrorParamsError("prometheus config is nil")
	}
	return &do.DatasourceMetricConfig{
		TeamId:         metricItem.GetTeam().GetTeamId(),
		ID:             metricItem.GetId(),
		Name:           metricItem.GetName(),
		Driver:         common.MetricDatasourceDriver_PROMETHEUS,
		Endpoint:       prometheusConfig.GetEndpoint(),
		Headers:        prometheusConfig.GetHeaders(),
		Method:         prometheusConfig.GetMethod(),
		CA:             prometheusConfig.GetCa(),
		BasicAuth:      ToBasicAuth(prometheusConfig.GetBasicAuth()),
		TLS:            ToTLS(prometheusConfig.GetTls()),
		Enable:         metricItem.GetEnable(),
		ScrapeInterval: metricItem.GetScrapeInterval().AsDuration(),
	}, nil
}

func ToMetricDatasourceConfigWithVictoriaMetrics(metricItem *common.MetricDatasourceItem) (*do.DatasourceMetricConfig, error) {
	victoriaMetricsConfig := metricItem.GetVictoriaMetrics()
	if victoriaMetricsConfig == nil {
		return nil, merr.ErrorParamsError("victoria metrics config is nil")
	}
	return &do.DatasourceMetricConfig{
		TeamId:         metricItem.GetTeam().GetTeamId(),
		ID:             metricItem.GetId(),
		Name:           metricItem.GetName(),
		Driver:         common.MetricDatasourceDriver_VICTORIAMETRICS,
		Endpoint:       victoriaMetricsConfig.GetEndpoint(),
		Headers:        victoriaMetricsConfig.GetHeaders(),
		Method:         victoriaMetricsConfig.GetMethod(),
		CA:             victoriaMetricsConfig.GetCa(),
		BasicAuth:      ToBasicAuth(victoriaMetricsConfig.GetBasicAuth()),
		TLS:            ToTLS(victoriaMetricsConfig.GetTls()),
		Enable:         metricItem.GetEnable(),
		ScrapeInterval: metricItem.GetScrapeInterval().AsDuration(),
	}, nil
}
