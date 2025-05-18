package prometheus

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/sync/errgroup"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

var _ datasource.Metric = (*Prometheus)(nil)

const (
	// prometheusAPIV1Query query api v1
	prometheusAPIV1Query = "/api/v1/query"
	// prometheusAPIV1QueryRange query range api v1
	prometheusAPIV1QueryRange = "/api/v1/query_range"
	// prometheusAPIV1Metadata metadata api
	prometheusAPIV1Metadata = "/api/v1/metadata"
	// prometheusAPIV1Series series api
	prometheusAPIV1Series = "/api/v1/series"
)

func New(c datasource.MetricConfig, logger log.Logger) *Prometheus {
	return &Prometheus{
		c:      c,
		helper: log.NewHelper(log.With(logger, "module", "plugin.datasource.prometheus")),
	}
}

type Prometheus struct {
	c      datasource.MetricConfig
	helper *log.Helper
}

// Proxy implements datasource.Metric.
func (p *Prometheus) Proxy(ctx transporthttp.Context, target string) error {
	w := ctx.Response()
	r := ctx.Request()

	// 获取query data
	query := r.URL.Query()
	// 绑定query到to
	api, err := url.JoinPath(p.c.GetEndpoint(), target)
	if !validate.IsNil(err) {
		return err
	}
	toURL, err := url.Parse(api)
	if !validate.IsNil(err) {
		return err
	}
	toURL.RawQuery = query.Encode()
	// body
	body := r.Body
	hx := p.configureHTTPClient(ctx)
	// 发起一个新请求， 把数据写回w
	proxyReq, err := http.NewRequestWithContext(ctx, r.Method, toURL.String(), body)
	if !validate.IsNil(err) {
		return err
	}
	proxyReq.Header = r.Header
	proxyReq.Form = r.Form
	proxyReq.Body = r.Body
	resp, err := hx.Do(proxyReq)
	if !validate.IsNil(err) {
		return err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			p.helper.Warnw("method", "prometheus.proxy", "err", err)
		}
	}(resp.Body)
	for k, v := range resp.Header {
		if len(v) == 0 {
			continue
		}
		w.Header().Set(k, v[0])
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func (p *Prometheus) GetScrapeInterval() time.Duration {
	if p.c.GetScrapeInterval() > 0 {
		return p.c.GetScrapeInterval()
	}
	return 15 * time.Second
}

func (p *Prometheus) Query(ctx context.Context, req *datasource.MetricQueryRequest) (*datasource.MetricQueryResponse, error) {
	if req.StartTime > 0 && req.EndTime > 0 {
		return p.queryRange(ctx, req.Expr, req.StartTime, req.EndTime, req.Step)
	}
	return p.query(ctx, req.Expr, req.Time)
}

func (p *Prometheus) Metadata(ctx context.Context) (<-chan *datasource.MetricMetadata, error) {
	metadataInfo, err := p.metadata(ctx)
	if err != nil {
		return nil, err
	}

	send := make(chan *datasource.MetricMetadata, 20)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				p.helper.Errorw("method", "metadata", "panic", err)
			}
		}()
		defer close(send)
		p.sendMetadata(send, metadataInfo)
	}()

	return send, nil
}

func (p *Prometheus) sendMetadata(send chan<- *datasource.MetricMetadata, metrics map[string][]PromMetricInfo) {
	metricNameMap := make(map[string]PromMetricInfo)
	metricNames := make([]string, 0, len(metrics))
	for metricName := range metrics {
		metricNames = append(metricNames, metricName)
		if len(metrics[metricName]) == 0 {
			continue
		}
		metricNameMap[metricName] = metrics[metricName][0]
	}

	now := timex.Now()
	batchNum := 20
	namesLen := len(metricNames)
	eg := new(errgroup.Group)
	for i := 0; i < namesLen; i += batchNum {
		left := i
		right := left + batchNum
		if right > namesLen {
			right = namesLen
		}
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			seriesInfo, seriesErr := p.series(ctx, now, metricNames[left:right]...)
			if seriesErr != nil {
				log.Warnw("series error", seriesErr)
				return seriesErr
			}

			metricsTmp := make([]*datasource.MetricMetadataItem, 0, right-left)
			for _, metricName := range metricNames[left:right] {
				metricInfo := metricNameMap[metricName]
				item := &datasource.MetricMetadataItem{
					Type:   metricInfo.Type,
					Name:   metricName,
					Help:   metricInfo.Help,
					Unit:   metricInfo.Unit,
					Labels: seriesInfo[metricName],
				}
				metricsTmp = append(metricsTmp, item)
			}
			send <- &datasource.MetricMetadata{
				Metric:    metricsTmp,
				Timestamp: timex.Now().Unix(),
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		p.helper.Errorw("method", "metadata", "err", err)
	}
}

func (p *Prometheus) configureHTTPClient(ctx context.Context) httpx.Client {
	hx := httpx.NewClient().WithContext(ctx)
	hx = hx.WithHeader(http.Header{
		"Accept":          []string{"*/*"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	})

	// Configure TLS if available
	if tls := p.c.GetTLS(); validate.IsNotNil(tls) {
		hx = hx.WithTLSClientConfig(tls.GetClientCert(), tls.GetClientKey())
		if serverName := tls.GetServerName(); serverName != "" {
			hx = hx.WithServerName(serverName)
		}
	}

	// Configure CA if available
	if ca := p.c.GetCA(); validate.TextIsNotNull(ca) {
		hx = hx.WithRootCA(ca)
	}

	// Configure authentication if available
	if basicAuth := p.c.GetBasicAuth(); validate.IsNotNil(basicAuth) {
		hx = hx.WithBasicAuth(basicAuth.GetUsername(), basicAuth.GetPassword())
	}

	// Configure custom headers if available
	if headers := p.c.GetHeaders(); len(headers) > 0 {
		for _, keyVal := range headers {
			hx = hx.WithHeader(http.Header{keyVal.Key: []string{keyVal.Value}})
		}
	}

	return hx
}

func (p *Prometheus) query(ctx context.Context, expr string, t int64) (*datasource.MetricQueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"time":  t,
	})

	hx := p.configureHTTPClient(ctx)
	api, err := url.JoinPath(p.c.GetEndpoint(), prometheusAPIV1Query)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.Get(api, params)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			p.helper.Errorw("method", "prometheus.query", "err", err)
		}
	}(getResponse.Body)
	if getResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResponse.Body)
		return nil, merr.ErrorBadRequest("status code: %d => %s", getResponse.StatusCode, string(body))
	}
	var allResp datasource.MetricQueryResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	return &allResp, nil
}

func (p *Prometheus) queryRange(ctx context.Context, expr string, start, end int64, step uint32) (*datasource.MetricQueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"start": start,
		"end":   end,
		"step":  step,
	})

	hx := p.configureHTTPClient(ctx)
	api, err := url.JoinPath(p.c.GetEndpoint(), prometheusAPIV1QueryRange)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.Get(api, params)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			p.helper.Warnw("method", "prometheus.queryRange", "err", err)
		}
	}(getResponse.Body)
	if getResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResponse.Body)
		return nil, merr.ErrorBadRequest("status code: %d => %s", getResponse.StatusCode, string(body))
	}
	var allResp datasource.MetricQueryResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}

	return &allResp, nil
}

func (p *Prometheus) series(ctx context.Context, now time.Time, metricNames ...string) (map[string]map[string][]string, error) {
	start := now.Add(-time.Hour * 24).Format("2006-01-02T15:04:05.000Z")
	end := now.Format("2006-01-02T15:04:05.000Z")

	params := httpx.ParseQuery(map[string]any{
		"start": start,
		"end":   end,
	})

	for _, metricName := range metricNames {
		params.Set("match[]", metricName)
	}

	hx := p.configureHTTPClient(ctx)
	api, err := url.JoinPath(p.c.GetEndpoint(), prometheusAPIV1Series)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.Get(api, params)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			p.helper.Warnw("method", "prometheus.series", "err", err)
		}
	}(getResponse.Body)
	if getResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResponse.Body)
		return nil, merr.ErrorBadRequest("status code: %d => %s", getResponse.StatusCode, string(body))
	}
	var allResp PromMetricSeriesResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}

	res := make(map[string]map[string][]string)
	for _, v := range allResp.Data {
		metricName := v["__name__"]
		if metricName == "" {
			continue
		}
		if _, ok := res[metricName]; !ok {
			res[metricName] = make(map[string][]string)
		}
		for k, val := range v {
			if k == "__name__" {
				continue
			}
			if _, ok := res[metricName][k]; !ok {
				res[metricName][k] = make([]string, 0)
			}
			res[metricName][k] = append(res[metricName][k], val)
		}
	}

	return res, nil
}

func (p *Prometheus) metadata(ctx context.Context) (map[string][]PromMetricInfo, error) {
	hx := p.configureHTTPClient(ctx)
	api, err := url.JoinPath(p.c.GetEndpoint(), prometheusAPIV1Metadata)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.Get(api)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			p.helper.Warnw("method", "prometheus.metadata", "err", err)
		}
	}(getResponse.Body)
	if getResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(getResponse.Body)
		return nil, merr.ErrorBadRequest("status code: %d => %s", getResponse.StatusCode, string(body))
	}
	var allResp PromMetadataResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	return allResp.Data, nil
}
