package datasource

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/status"
)

var _ MetricDatasource = (*victoriaMetricsDatasource)(nil)

// NewVictoriaMetricsDatasource creates a new victoriaMetrics datasource.
func NewVictoriaMetricsDatasource(opts ...VictoriaMetricsDatasourceOption) MetricDatasource {
	v := &victoriaMetricsDatasource{
		step:          14,
		queryAPI:      victoriaMetricsQueryAPI,
		queryRangeAPI: victoriaMetricsQueryRangeAPI,
		metadataAPI:   victoriaMetricsMetadataAPI,
		seriesAPI:     victoriaMetricsSeriesAPI,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

type (
	// VictoriaMetricsDatasourceOption victoriaMetrics数据源选项
	VictoriaMetricsDatasourceOption func(*victoriaMetricsDatasource)

	// victoriaMetricsDatasource victoriaMetrics数据源
	victoriaMetricsDatasource struct {
		// basicAuth 数据源基础认证
		basicAuth *BasicAuth

		// endpoint 数据源地址
		endpoint string
		// step 默认步长
		step uint32

		id uint32

		queryAPI      string
		queryRangeAPI string
		seriesAPI     string
		metadataAPI   string
	}

	// VictoriaMetricsMetadataResponse victoriaMetrics元数据响应
	VictoriaMetricsMetadataResponse struct {
		Status string   `json:"status"`
		Data   []string `json:"data"`
	}
)

func (p *victoriaMetricsDatasource) GetBasicInfo() *BasicInfo {
	return &BasicInfo{
		Endpoint:  p.endpoint,
		BasicAuth: p.basicAuth,
		ID:        p.id,
	}
}

const (
	victoriaMetricsQueryAPI      = "/api/v1/query"
	victoriaMetricsQueryRangeAPI = "/api/v1/query_range"
	victoriaMetricsSeriesAPI     = "/api/v1/series"
	victoriaMetricsMetadataAPI   = "/api/v1/label/__name__/values"
)

func (p *victoriaMetricsDatasource) Step() uint32 {
	if p.step == 0 {
		return 14
	}
	return p.step
}

func (p *victoriaMetricsDatasource) Query(ctx context.Context, expr string, duration int64) ([]*QueryResponse, error) {
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
	api, err := url.JoinPath(p.endpoint, p.queryAPI)
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
	for _, queryResult := range data.Result {
		value := queryResult.Value
		ts, tsAssertOk := strconv.ParseFloat(fmt.Sprintf("%v", value[0]), 64)
		if tsAssertOk != nil {
			continue
		}
		metricValue, parseErr := strconv.ParseFloat(fmt.Sprintf("%v", value[1]), 64)
		if parseErr != nil {
			continue
		}
		result = append(result, &QueryResponse{
			Labels: queryResult.Metric,
			Value: &QueryValue{
				Value:     metricValue,
				Timestamp: int64(ts),
			},
			ResultType: data.ResultType,
		})
	}

	return result, nil
}

func (p *victoriaMetricsDatasource) QueryRange(ctx context.Context, expr string, start, end int64, step uint32) ([]*QueryResponse, error) {
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
	api, err := url.JoinPath(p.endpoint, p.queryRangeAPI)
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

func (p *victoriaMetricsDatasource) Metadata(ctx context.Context) (*Metadata, error) {
	now := time.Now()
	// 获取元数据
	metadataInfo, err := p.metadata(ctx)
	if err != nil {
		return nil, err
	}
	metricNameMap := make(map[string]PromMetricInfo)
	metricNames := make([]string, 0, len(metadataInfo))
	for metricName := range metadataInfo {
		metricNames = append(metricNames, metricName)
		if len(metadataInfo[metricName]) == 0 {
			continue
		}
		metricNameMap[metricName] = metadataInfo[metricName][0]
	}

	metricList := newMetrics(len(metricNames))
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
			for _, metricName := range metricNames[left:right] {
				metricInfo := metricNameMap[metricName]
				item := &Metric{
					Type:   metricInfo.Type,
					Name:   metricName,
					Help:   metricInfo.Help,
					Unit:   metricInfo.Unit,
					Labels: seriesInfo[metricName],
				}
				metricsTmp = append(metricsTmp, item)
			}

			metricList.append(metricsTmp...)
			return nil
		})
	}
	if err = eg.Wait(); err != nil {
		return nil, err
	}

	return &Metadata{
		Metric:    metricList.list,
		Timestamp: now.Unix(),
	}, nil
}

// metadata 获取原始数据
func (p *victoriaMetricsDatasource) metadata(ctx context.Context) (map[string][]PromMetricInfo, error) {
	hx := httpx.NewHTTPX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}
	api, err := url.JoinPath(p.endpoint, p.metadataAPI)
	if err != nil {
		return nil, err
	}
	getResponse, err := hx.GETWithContext(ctx, api)
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp VictoriaMetricsMetadataResponse
	if err = types.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	metadataResp := make(map[string][]PromMetricInfo, len(allResp.Data))
	for _, datum := range allResp.Data {
		metadataResp[datum] = []PromMetricInfo{{}}
	}
	return metadataResp, nil
}

// series 查询时间序列 map[metricName][labelName][]labelValue
func (p *victoriaMetricsDatasource) series(ctx context.Context, t time.Time, metricNames ...string) (map[string]map[string][]string, error) {
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

	api, err := url.JoinPath(p.endpoint, p.seriesAPI)
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

// WithVictoriaMetricsEndpoint 设置数据源地址
func WithVictoriaMetricsEndpoint(endpoint string) VictoriaMetricsDatasourceOption {
	return func(p *victoriaMetricsDatasource) {
		p.endpoint = endpoint
	}
}

// WithVictoriaMetricsStep 设置步长
func WithVictoriaMetricsStep(step uint32) VictoriaMetricsDatasourceOption {
	return func(p *victoriaMetricsDatasource) {
		if step <= 0 {
			p.step = 10
			return
		}
		p.step = step
	}
}

// WithVictoriaMetricsBasicAuth 设置数据源配置
func WithVictoriaMetricsBasicAuth(username, password string) VictoriaMetricsDatasourceOption {
	return func(p *victoriaMetricsDatasource) {
		if types.TextIsNull(username) || types.TextIsNull(password) {
			return
		}
		p.basicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

// WithVictoriaMetricsID 设置数据源ID
func WithVictoriaMetricsID(id uint32) VictoriaMetricsDatasourceOption {
	return func(p *victoriaMetricsDatasource) {
		p.id = id
	}
}
