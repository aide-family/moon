package collector

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

const (
	// Metric name for datasource status. 1 = healthy, 0 = unhealthy.
	datasourceStatusMetricName = "marksman_datasource_status"
	probeTimeout               = 5 * time.Second
)

// DatasourceItem is a minimal view of a datasource for status probing.
type DatasourceItem struct {
	UID  int64
	Name string
	URL  string
}

// DatasourceLister lists datasources that should be probed (e.g. for metrics).
type DatasourceLister interface {
	ListForProbe(ctx context.Context) ([]DatasourceItem, error)
}

// NewDatasourceCollector returns a Prometheus collector that probes each datasource
// via HTTP and exposes marksman_datasource_status (1 = up, 0 = down).
func NewDatasourceCollector(lister DatasourceLister) prometheus.Collector {
	return &datasourceCollector{
		lister: lister,
		statusDesc: prometheus.NewDesc(
			datasourceStatusMetricName,
			"Datasource health status from HTTP probe. 1 = healthy, 0 = unhealthy.",
			[]string{"uid", "name"},
			nil,
		),
	}
}

type datasourceCollector struct {
	lister     DatasourceLister
	statusDesc *prometheus.Desc
}

// Describe implements [prometheus.Collector].
func (d *datasourceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- d.statusDesc
}

// Collect implements [prometheus.Collector].
func (d *datasourceCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	items, err := d.lister.ListForProbe(ctx)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: probeTimeout}
	for _, item := range items {
		status := float64(0)
		if probeURL(client, item.URL) {
			status = 1
		}
		ch <- prometheus.MustNewConstMetric(
			d.statusDesc,
			prometheus.GaugeValue,
			status,
			strconv.FormatInt(item.UID, 10),
			item.Name,
		)
	}
}

// NewDatasourceListerFromRepo returns a DatasourceLister that uses the repository to list all datasources for probing.
func NewDatasourceListerFromRepo(repo repository.Datasource) DatasourceLister {
	return &repoDatasourceLister{repo: repo}
}

type repoDatasourceLister struct {
	repo repository.Datasource
}

func (r *repoDatasourceLister) ListForProbe(ctx context.Context) ([]DatasourceItem, error) {
	list, err := r.repo.ListAllForProbe(ctx, 500)
	if err != nil {
		return nil, err
	}
	out := make([]DatasourceItem, 0, len(list))
	for _, item := range list {
		out = append(out, fromDatasourceItemBo(item))
	}
	return out, nil
}

func fromDatasourceItemBo(b *bo.DatasourceItemBo) DatasourceItem {
	if b == nil {
		return DatasourceItem{}
	}
	return DatasourceItem{
		UID:  b.UID.Int64(),
		Name: b.Name,
		URL:  b.URL,
	}
}

// probeURL performs a GET request to url. Returns true if response is 2xx.
// GET is used because many metrics endpoints (e.g. Prometheus) do not support HEAD (405).
func probeURL(client *http.Client, url string) bool {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
