package datasource

import (
	"context"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"
	"testing"
	"time"
)

func Test_prometheusDatasource_series(t *testing.T) {
	prom := NewPrometheusDatasource(
		WithPrometheusEndpoint("https://prometheus.aide-cloud.cn/"),
		WithPrometheusStep(14),
		WithPrometheusID(1),
	)
	p := prom.(*prometheusDatasource)

	seriesInfo, seriesErr := p.series(context.Background(), time.Now(), "node_load1", "node_load15")
	if seriesErr != nil {
		log.Warnw("series error", seriesErr)
		return
	}
	bs, _ := types.Marshal(seriesInfo)
	t.Log(string(bs))
}

func Test_prometheusDatasource_metadata(t *testing.T) {
	prom := NewPrometheusDatasource(
		WithPrometheusEndpoint("https://prometheus.aide-cloud.cn/"),
		WithPrometheusStep(14),
		WithPrometheusID(1),
	)
	p := prom.(*prometheusDatasource)
	metadataInfo, err := p.metadata(context.Background())
	if err != nil {
		log.Warnw("metadata error", err)
		return
	}
	bs, _ := types.Marshal(metadataInfo)
	t.Log(string(bs))
}

func Test_prometheusDatasource_Metadata(t *testing.T) {
	prom := NewPrometheusDatasource(
		WithPrometheusEndpoint("https://prometheus.aide-cloud.cn/"),
		WithPrometheusStep(14),
		WithPrometheusID(1),
	)
	p := prom.(*prometheusDatasource)
	metricInfo, err := p.Metadata(context.Background())
	if err != nil {
		log.Warnw("metadata error", err)
		return
	}
	bs, _ := types.Marshal(metricInfo)
	t.Log(string(bs))
}
