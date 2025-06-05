package data

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/plugin/server/cron_server"
	"github.com/aide-family/moon/pkg/util/safety"
)

// ProviderSetData is a set of data providers.
var ProviderSetData = wire.NewSet(New)

func New(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	var err error

	data := &Data{
		counterMetrics:          safety.NewMap[string, *prometheus.CounterVec](),
		gaugeMetrics:            safety.NewMap[string, *prometheus.GaugeVec](),
		histogramMetrics:        safety.NewMap[string, *prometheus.HistogramVec](),
		summaryMetrics:          safety.NewMap[string, *prometheus.SummaryVec](),
		scriptJobEventBus:       make(chan cron_server.CronJob, 100),
		removeScriptJobEventBus: make(chan cron_server.CronJob, 100),
		metricEventBus:          make(chan []byte, 100),
		helper:                  log.NewHelper(log.With(logger, "module", "data")),
	}
	data.cache, err = cache.NewCache(c.GetCache())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if err = data.cache.Close(); err != nil {
			log.NewHelper(logger).Errorw("method", "close cache", "err", err)
		}
	}
	return data, cleanup, nil
}

type Data struct {
	cache                   cache.Cache
	counterMetrics          *safety.Map[string, *prometheus.CounterVec]
	gaugeMetrics            *safety.Map[string, *prometheus.GaugeVec]
	histogramMetrics        *safety.Map[string, *prometheus.HistogramVec]
	summaryMetrics          *safety.Map[string, *prometheus.SummaryVec]
	scriptJobEventBus       chan cron_server.CronJob
	removeScriptJobEventBus chan cron_server.CronJob
	metricEventBus          chan []byte
	helper                  *log.Helper
}

func (d *Data) GetCache() cache.Cache {
	return d.cache
}

func (d *Data) SetCounterMetric(name string, metrics *prometheus.CounterVec) bool {
	if _, ok := d.counterMetrics.Get(name); ok {
		return false
	}
	d.counterMetrics.Set(name, metrics)
	return true
}

func (d *Data) SetGaugeMetric(name string, metrics *prometheus.GaugeVec) bool {
	if _, ok := d.gaugeMetrics.Get(name); ok {
		return false
	}
	d.gaugeMetrics.Set(name, metrics)
	return true
}

func (d *Data) SetHistogramMetric(name string, metrics *prometheus.HistogramVec) bool {
	if _, ok := d.histogramMetrics.Get(name); ok {
		return false
	}
	d.histogramMetrics.Set(name, metrics)
	return true
}

func (d *Data) SetSummaryMetric(name string, metrics *prometheus.SummaryVec) bool {
	if _, ok := d.summaryMetrics.Get(name); ok {
		return false
	}
	d.summaryMetrics.Set(name, metrics)
	return true
}

func (d *Data) GetCounterMetric(name string) (*prometheus.CounterVec, bool) {
	return d.counterMetrics.Get(name)
}

func (d *Data) GetGaugeMetric(name string) (*prometheus.GaugeVec, bool) {
	return d.gaugeMetrics.Get(name)
}

func (d *Data) GetHistogramMetric(name string) (*prometheus.HistogramVec, bool) {
	return d.histogramMetrics.Get(name)
}

func (d *Data) GetSummaryMetric(name string) (*prometheus.SummaryVec, bool) {
	return d.summaryMetrics.Get(name)
}

func (d *Data) OutScriptJobEventBus() <-chan cron_server.CronJob {
	return d.scriptJobEventBus
}

func (d *Data) InScriptJobEventBus(job cron_server.CronJob) {
	d.scriptJobEventBus <- job
}

func (d *Data) OutRemoveScriptJobEventBus() <-chan cron_server.CronJob {
	return d.removeScriptJobEventBus
}

func (d *Data) InRemoveScriptJobEventBus(job cron_server.CronJob) {
	d.removeScriptJobEventBus <- job
}

func (d *Data) OutMetricEventBus() <-chan []byte {
	return d.metricEventBus
}

func (d *Data) InMetricEventBus(event []byte) {
	d.metricEventBus <- event
}
