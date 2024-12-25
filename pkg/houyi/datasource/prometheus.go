package datasource

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
)

const (
	// prometheusAPIV1Query 查询接口
	prometheusAPIV1Query = "/api/v1/query"
	// prometheusAPIV1QueryRange 查询接口
	prometheusAPIV1QueryRange = "/api/v1/query_range"
	// prometheusAPIV1Metadata 元数据查询接口
	prometheusAPIV1Metadata = "/api/v1/metadata"
	// prometheusAPIV1Series /api/v1/series
	prometheusAPIV1Series = "/api/v1/series"
)

// NewPrometheusDatasource 创建 prometheus 数据源
func NewPrometheusDatasource(opts ...PrometheusOption) MetricDatasource {
	p := &prometheusDatasource{
		prometheusAPIV1Query:      prometheusAPIV1Query,
		prometheusAPIV1QueryRange: prometheusAPIV1QueryRange,
		prometheusAPIV1Metadata:   prometheusAPIV1Metadata,
		prometheusAPIV1Series:     prometheusAPIV1Series,
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

type (
	prometheusDatasource struct {
		// prom ql 查询接口
		prometheusAPIV1Query string
		// prom ql 查询接口
		prometheusAPIV1QueryRange string
		// prom 元数据查询接口
		prometheusAPIV1Metadata string
		// prom series 查询接口
		prometheusAPIV1Series string

		// basicAuth 数据源基础认证
		basicAuth *BasicAuth

		// endpoint 数据源地址
		endpoint string
		// 查询步长
		step uint32

		id uint32
	}

	// PrometheusOption 数据源配置
	PrometheusOption func(p *prometheusDatasource)

	// 响应参数处理

	// PromQueryResult 查询响应数据
	PromQueryResult struct {
		Metric map[string]string `json:"metric"`
		Value  [2]any            `json:"value"`
		Values [][2]any          `json:"values"`
	}

	// PromQueryData 查询数据
	PromQueryData struct {
		ResultType string             `json:"resultType"`
		Result     []*PromQueryResult `json:"result"`
	}

	// PromQueryResponse 查询响应
	// {"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"node_cpu_seconds_total","cpu":"0","instance":"10.0.0.1:9100","job":"node-exporter","mode":"idle"},"value":[1629874800.0,"0.01"]},{"metric":{"__name__":"node_cpu_seconds_total","cpu":"0","instance":"10.0.0.2:9100","job":"node-exporter","mode":"idle"},"value":[1629874800.0,"0.01"]}]}}
	PromQueryResponse struct {
		Status    string         `json:"status"`
		Data      *PromQueryData `json:"data"`
		ErrorType string         `json:"errorType"`
		Error     string         `json:"error"`
	}

	// PromMetricInfo 元数据信息
	PromMetricInfo struct {
		Type string `json:"type"`
		Help string `json:"help"`
		Unit string `json:"unit"`
	}

	// PromMetadataResponse 元数据响应
	PromMetadataResponse struct {
		Status string                      `json:"status"`
		Data   map[string][]PromMetricInfo `json:"data"`
	}

	// PromMetricSeriesResponse 元数据响应
	PromMetricSeriesResponse struct {
		Status    string              `json:"status"`
		Data      []map[string]string `json:"data"`
		Error     string              `json:"error"`
		ErrorType string              `json:"errorType"`
	}
)

func (p *prometheusDatasource) GetBasicInfo() *BasicInfo {
	return &BasicInfo{
		Endpoint:  p.endpoint,
		BasicAuth: p.basicAuth,
		ID:        p.id,
	}
}

func (p *prometheusDatasource) Step() uint32 {
	if p.step == 0 {
		return 14
	}
	return p.step
}

func (p *prometheusDatasource) QueryRange(ctx context.Context, expr string, start, end int64, step uint32) ([]*QueryResponse, error) {
	st := step
	if step == 0 {
		st = p.step
	}
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"start": start,
		"end":   end,
		"step":  st,
	})

	hx := httpx.NewHTTPX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}
	api, err := url.JoinPath(p.endpoint, p.prometheusAPIV1QueryRange)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.GETWithContext(ctx, fmt.Sprintf("%s?%s", api, params))
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp PromQueryResponse
	if err = types.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	if allResp.Error != "" {
		return nil, status.Errorf(400, "query error: %s", allResp.Error)
	}
	data := allResp.Data
	if types.IsNil(data) {
		return []*QueryResponse(nil), nil
	}
	result := make([]*QueryResponse, 0, len(data.Result))
	for _, v := range data.Result {
		values := make([]*QueryValue, 0, len(v.Values))
		for _, vv := range v.Values {
			ts, tsAssertOk := strconv.ParseFloat(fmt.Sprintf("%v", vv[0]), 64)
			if tsAssertOk != nil {
				continue
			}
			metricValue, parseErr := strconv.ParseFloat(fmt.Sprintf("%v", vv[1]), 64)
			if parseErr != nil {
				continue
			}
			values = append(values, &QueryValue{
				Value:     metricValue,
				Timestamp: int64(ts),
			})
		}

		result = append(result, &QueryResponse{
			Labels:     v.Metric,
			Values:     values,
			ResultType: data.ResultType,
		})
	}
	return result, nil
}

func (p *prometheusDatasource) Query(ctx context.Context, expr string, duration int64) ([]*QueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"time":  duration,
	})

	hx := httpx.NewHTTPX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}
	api, err := url.JoinPath(p.endpoint, p.prometheusAPIV1Query)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.GETWithContext(ctx, fmt.Sprintf("%s?%s", api, params))
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp PromQueryResponse
	if err = types.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}

	if allResp.Error != "" {
		return nil, status.Errorf(400, "query error: %s", allResp.Error)
	}
	data := allResp.Data
	if types.IsNil(data) {
		return []*QueryResponse(nil), nil
	}
	result := make([]*QueryResponse, 0, len(data.Result))
	for _, v := range data.Result {
		value := v.Value
		ts, tsAssertOk := strconv.ParseFloat(fmt.Sprintf("%v", value[0]), 64)
		if tsAssertOk != nil {
			continue
		}
		metricValue, parseErr := strconv.ParseFloat(fmt.Sprintf("%v", value[1]), 64)
		if parseErr != nil {
			continue
		}
		result = append(result, &QueryResponse{
			Labels: v.Metric,
			Value: &QueryValue{
				Value:     metricValue,
				Timestamp: int64(ts),
			},
			ResultType: data.ResultType,
		})
	}

	return result, nil
}

func (p *prometheusDatasource) Metadata(ctx context.Context) (*Metadata, error) {
	now := time.Now()
	// 获取元数据
	metadataInfo, err := p.metadata(ctx)
	if err != nil {
		return nil, err
	}
	metricNames := make([]string, 0, len(metadataInfo))
	for metricName := range metadataInfo {
		metricNames = append(metricNames, metricName)
	}

	// metricNames = metricNames[:151]
	metrics := make([]*Metric, 0, len(metricNames))
	lock := new(sync.RWMutex)
	batchNum := 20
	namesLen := len(metricNames)
	eg := new(errgroup.Group)
	// 因为数据量比较大， 这里并发获取各个metric的元数据， 并且将结果写入到metrics中
	for i := 0; i < namesLen; i += batchNum {
		left := i
		right := left + batchNum
		if right > namesLen {
			right = namesLen
		}
		eg.Go(func() error {
			seriesInfo, seriesErr := p.series(context.Background(), now, metricNames[left:right]...)
			if seriesErr != nil {
				log.Warnw("series error", seriesErr)
				return seriesErr
			}

			metricsTmp := make([]*Metric, 0, right-left)
			for metricName, metricInfos := range metadataInfo {
				if len(metricInfos) == 0 {
					continue
				}
				metricInfo := metricInfos[0]
				item := &Metric{
					Type:   metricInfo.Type,
					Name:   metricName,
					Help:   metricInfo.Help,
					Unit:   metricInfo.Unit,
					Labels: seriesInfo[metricName],
				}
				metricsTmp = append(metricsTmp, item)
			}
			lock.Lock()
			defer lock.Unlock()
			metrics = append(metrics, metricsTmp...)
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}

	return &Metadata{
		Metric:    metrics,
		Timestamp: now.Unix(),
	}, nil
}

// metadata 获取原始数据
func (p *prometheusDatasource) metadata(ctx context.Context) (map[string][]PromMetricInfo, error) {
	hx := httpx.NewHTTPX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}
	api, err := url.JoinPath(p.endpoint, p.prometheusAPIV1Metadata)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.GETWithContext(ctx, api)
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp PromMetadataResponse
	if err = types.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	return allResp.Data, nil
}

// series 查询时间序列 map[metricName][labelName][]labelValue
func (p *prometheusDatasource) series(ctx context.Context, t time.Time, metricNames ...string) (map[string]map[string][]string, error) {
	now := t
	// 获取此格式时间2024-04-21T17:58:55.061Z
	start := now.Add(-time.Hour * 24).Format("2006-01-02T15:04:05.000Z")
	end := now.Format("2006-01-02T15:04:05.000Z")

	params := httpx.ParseQuery(map[string]any{
		"start": start,
		"end":   end,
	})
	metricNameParams := make([]string, 0, len(metricNames))
	for _, metricName := range metricNames {
		metricNameParams = append(metricNameParams, httpx.ParseQuery(map[string]any{
			"match[]": metricName,
		}))
	}

	hx := httpx.NewHTTPX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}

	api, err := url.JoinPath(p.endpoint, p.prometheusAPIV1Series)
	if err != nil {
		return nil, err
	}
	reqURL := fmt.Sprintf("%s?%s&%s", api, params, strings.Join(metricNameParams, "&"))
	getResponse, err := hx.GETWithContext(ctx, reqURL)
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp PromMetricSeriesResponse
	if err = types.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
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

// WithPrometheusEndpoint 设置数据源地址
func WithPrometheusEndpoint(endpoint string) PrometheusOption {
	return func(p *prometheusDatasource) {
		p.endpoint = endpoint
	}
}

// WithPrometheusStep 设置步长
func WithPrometheusStep(step uint32) PrometheusOption {
	return func(p *prometheusDatasource) {
		if step <= 0 {
			p.step = 10
			return
		}
		p.step = step
	}
}

// WithPrometheusBasicAuth 设置数据源配置
func WithPrometheusBasicAuth(username, password string) PrometheusOption {
	return func(p *prometheusDatasource) {
		if types.TextIsNull(username) || types.TextIsNull(password) {
			return
		}
		p.basicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

// WithPrometheusID 设置数据源ID
func WithPrometheusID(id uint32) PrometheusOption {
	return func(p *prometheusDatasource) {
		p.id = id
	}
}
