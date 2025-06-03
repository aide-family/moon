package victoria

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
	"github.com/go-kratos/kratos/v2/log"
	transporthttp "github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/sync/errgroup"
)

var _ datasource.Metric = (*Victoria)(nil)

const (
	victoriaMetricsQueryAPI      = "/api/v1/query"
	victoriaMetricsQueryRangeAPI = "/api/v1/query_range"
	victoriaMetricsSeriesAPI     = "/api/v1/series"
	victoriaMetricsMetadataAPI   = "/api/v1/label/__name__/values"
)

func New(c datasource.MetricConfig, logger log.Logger) *Victoria {
	return &Victoria{
		c:      c,
		helper: log.NewHelper(log.With(logger, "module", "plugin.datasource.victoria")),
	}
}

type Victoria struct {
	c      datasource.MetricConfig
	helper *log.Helper
}

// GetScrapeInterval implements datasource.Metric.
func (v *Victoria) GetScrapeInterval() time.Duration {
	if interval := v.c.GetScrapeInterval(); interval > 0 {
		return interval
	}
	return 15 * time.Second
}

// Metadata implements datasource.Metric.
func (v *Victoria) Metadata(ctx context.Context) (<-chan *datasource.MetricMetadata, error) {
	metadataInfo, err := v.metadataNames(ctx)
	if err != nil {
		return nil, err
	}

	send := make(chan *datasource.MetricMetadata, 20)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Warnw("method", "victoriametrics.metadata", "err", err)
			}
		}()
		defer close(send)
		v.sendMetadata(send, metadataInfo)
	}()

	return send, nil
}

func (v *Victoria) sendMetadata(send chan<- *datasource.MetricMetadata, metricNames []string) {
	batchNum := 20
	namesLen := len(metricNames)

	eg := new(errgroup.Group)
	eg.SetLimit(10)
	for i := 0; i < namesLen; i += batchNum {
		left := i
		right := left + batchNum
		if right > namesLen {
			right = namesLen
		}
		operationNames := metricNames[left:right]
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
			defer cancel()
			seriesInfo, seriesErr := v.series(ctx, operationNames...)
			if seriesErr != nil {
				log.Warnw("series error", seriesErr)
				return seriesErr
			}

			metricsTmp := make([]*datasource.MetricMetadataItem, 0, right-left)
			for _, metricName := range operationNames {
				item := &datasource.MetricMetadataItem{
					Type:   "",
					Name:   metricName,
					Help:   "",
					Unit:   "",
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
		log.Warnw("series error", err)
	}
}

func (v *Victoria) series(ctx context.Context, metricNames ...string) (map[string]map[string][]string, error) {
	params := httpx.ParseQuery(map[string]any{})
	for _, metricName := range metricNames {
		params.Add("match[]", metricName)
	}
	hx := v.configureHTTPClient(ctx)
	api, err := url.JoinPath(v.c.GetEndpoint(), victoriaMetricsSeriesAPI)
	if err != nil {
		return nil, err
	}
	response, err := hx.Get(api, params)
	if err != nil {
		return nil, err
	}
	var resp Series
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}
	data := make(map[string]map[string][]string, len(resp.Data))
	for _, datum := range resp.Data {
		metricName := datum["__name__"]
		if metricName == "" {
			continue
		}
		if _, ok := data[metricName]; !ok {
			data[metricName] = make(map[string][]string)
		}
		for k, val := range datum {
			if k == "__name__" {
				continue
			}
			if _, ok := data[metricName][k]; !ok {
				data[metricName][k] = make([]string, 0)
			}
			data[metricName][k] = slices.Unique(append(data[metricName][k], val))
		}
	}
	return data, nil
}

func (v *Victoria) metadataNames(ctx context.Context) ([]string, error) {
	hx := v.configureHTTPClient(ctx)
	api, err := url.JoinPath(v.c.GetEndpoint(), victoriaMetricsMetadataAPI)
	if err != nil {
		return nil, err
	}
	response, err := hx.Get(api)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			v.helper.Warnw("method", "victoriametrics.metadata", "err", err)
		}
	}(response.Body)
	var resp Names
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// Proxy implements datasource.Metric.
func (v *Victoria) Proxy(ctx transporthttp.Context, target string) error {
	w := ctx.Response()
	r := ctx.Request()

	// Get query data
	query := r.URL.Query()
	// Bind query to target
	api, err := url.JoinPath(v.c.GetEndpoint(), target)
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
	hx := v.configureHTTPClient(ctx)
	// Initiate a new request and write data back to w
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
			v.helper.Warnw("method", "victoriametrics.proxy", "err", err)
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

// Query implements datasource.Metric.
func (v *Victoria) Query(ctx context.Context, req *datasource.MetricQueryRequest) (*datasource.MetricQueryResponse, error) {
	if req.StartTime > 0 && req.EndTime > 0 {
		return v.queryRange(ctx, req.Expr, req.StartTime, req.EndTime, req.Step)
	}
	return v.query(ctx, req.Expr, req.Time)
}

func (v *Victoria) queryRange(ctx context.Context, expr string, start, end int64, step uint32) (*datasource.MetricQueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"start": start,
		"end":   end,
		"step":  step,
	})

	hx := v.configureHTTPClient(ctx)
	api, err := url.JoinPath(v.c.GetEndpoint(), victoriaMetricsQueryRangeAPI)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.Get(api, params)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			v.helper.Warnw("method", "prometheus.queryRange", "err", err)
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

func (v *Victoria) query(ctx context.Context, expr string, t int64) (*datasource.MetricQueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"time":  t,
	})

	hx := v.configureHTTPClient(ctx)
	api, err := url.JoinPath(v.c.GetEndpoint(), victoriaMetricsQueryAPI)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.Get(api, params)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			v.helper.Errorw("method", "prometheus.query", "err", err)
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

func (v *Victoria) configureHTTPClient(ctx context.Context) httpx.Client {
	hx := httpx.NewClient().WithContext(ctx)
	hx = hx.WithHeader(http.Header{
		"Accept":          []string{"*/*"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	})

	// Configure TLS if available
	if tls := v.c.GetTLS(); validate.IsNotNil(tls) {
		hx = hx.WithTLSClientConfig(tls.GetClientCert(), tls.GetClientKey())
		if serverName := tls.GetServerName(); serverName != "" {
			hx = hx.WithServerName(serverName)
		}
	}

	// Configure CA if available
	if ca := v.c.GetCA(); validate.TextIsNotNull(ca) {
		hx = hx.WithRootCA(ca)
	}

	// Configure authentication if available
	if basicAuth := v.c.GetBasicAuth(); validate.IsNotNil(basicAuth) {
		hx = hx.WithBasicAuth(basicAuth.GetUsername(), basicAuth.GetPassword())
	}

	// Configure custom headers if available
	if headers := v.c.GetHeaders(); len(headers) > 0 {
		for _, keyVal := range headers {
			hx = hx.WithHeader(http.Header{keyVal.Key: []string{keyVal.Value}})
		}
	}

	return hx
}
