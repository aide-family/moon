package p8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/httpx"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

const (
	// prometheusApiV1Query 查询接口
	prometheusApiV1Query = "/api/v1/query"
	// prometheusApiV1Metadata 元数据查询接口
	prometheusApiV1Metadata = "/api/v1/metadata"
	// prometheusApiV1Series /api/v1/series
	prometheusApiV1Series = "/api/v1/series"
)

func NewPrometheusDatasource(opts ...Option) agent.Datasource {
	p := &PrometheusDatasource{
		category: agent.Prometheus,
	}
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// PrometheusDatasource 数据源
type PrometheusDatasource struct {
	// 地址
	endpoint string
	// 数据源类型
	category agent.Category
	// 基础认证
	basicAuth *agent.BasicAuth

	mut sync.RWMutex
}

func (p *PrometheusDatasource) Metadata(ctx context.Context) (*agent.Metadata, error) {
	now := time.Now()
	metadataInfo, err := p.metadata(ctx)
	if err != nil {
		return nil, err
	}
	metricNames := make([]string, 0, len(metadataInfo))
	for metricName := range metadataInfo {
		metricNames = append(metricNames, metricName)
	}

	//metricNames = metricNames[:151]
	metrics := make([]*agent.MetricDetail, 0, len(metricNames))
	lock := new(sync.RWMutex)
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
			seriesInfo, seriesErr := p.series(context.Background(), now, metricNames[left:right]...)
			if seriesErr != nil {
				log.Warnw("series error", seriesErr)
				return seriesErr
			}

			metricsTmp := make([]*agent.MetricDetail, 0, right-left)
			for metricName, metricInfos := range metadataInfo {
				if len(metricInfos) == 0 {
					continue
				}
				metricInfo := metricInfos[0]
				item := &agent.MetricDetail{
					Type:   metricInfo.GetType(),
					Name:   metricName,
					Help:   metricInfo.GetHelp(),
					Unit:   metricInfo.GetUnit(),
					Labels: agent.Labels(seriesInfo[metricName]),
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

	return &agent.Metadata{
		Metric: metrics,
		Unix:   now.Unix(),
	}, nil
}

// metadata 获取原始数据
func (p *PrometheusDatasource) metadata(ctx context.Context) (map[string][]MetricInfo, error) {
	hx := httpx.NewHttpX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}
	getResponse, err := hx.GETWithContext(ctx, fmt.Sprintf("%s%s", p.endpoint, prometheusApiV1Metadata))
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp MetadataResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	return allResp.GetData(), nil
}

// series 查询时间序列
func (p *PrometheusDatasource) series(ctx context.Context, t time.Time, metricNames ...string) (map[string]MetricLabel, error) {
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

	hx := httpx.NewHttpX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}

	url := fmt.Sprintf(
		"%s%s?%s&%s",
		p.endpoint,
		prometheusApiV1Series,
		params,
		strings.Join(metricNameParams, "&"),
	)

	getResponse, err := hx.GETWithContext(ctx, url)
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp MetricSeriesResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}

	res := make(map[string]MetricLabel)
	for _, v := range allResp.GetData() {
		metricName := v.GetMetricName()
		if metricName == "" {
			continue
		}
		res[metricName] = v
	}
	return res, nil
}

func (p *PrometheusDatasource) Query(ctx context.Context, expr string, duration int64) (*agent.QueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"time":  duration,
	})

	hx := httpx.NewHttpX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if p.basicAuth != nil {
		hx = hx.SetBasicAuth(p.basicAuth.Username, p.basicAuth.Password)
	}
	getResponse, err := hx.GETWithContext(ctx, fmt.Sprintf("%s%s?%s", p.endpoint, prometheusApiV1Query, params))
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp QueryResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	data := allResp.GetData()
	result := make([]*agent.Result, 0, len(data.GetResult()))
	for _, v := range data.GetResult() {
		value := v.GetValue()
		if len(value) < 2 {
			continue
		}
		ts, tsAssertOk := value[0].(float64)
		if !tsAssertOk {
			continue
		}
		metricValue, parseErr := strconv.ParseFloat(fmt.Sprintf("%v", value[1]), 64)
		if parseErr != nil {
			continue
		}
		result = append(result, &agent.Result{
			Metric: v.GetMetric(),
			Ts:     ts,
			Value:  metricValue,
		})
	}

	return &agent.QueryResponse{
		Status: allResp.GetStatus(),
		Data: &agent.Data{
			ResultType: data.GetResultType(),
			Result:     result,
		},
		ErrorType: allResp.GetErrorType(),
		Error:     allResp.GetError(),
	}, nil
}

func (p *PrometheusDatasource) GetCategory() agent.Category {
	return p.category
}

func (p *PrometheusDatasource) GetEndpoint() string {
	return p.endpoint
}

func (p *PrometheusDatasource) GetBasicAuth() *agent.BasicAuth {
	return p.basicAuth
}

func (p *PrometheusDatasource) WithBasicAuth(basicAuth *agent.BasicAuth) agent.Datasource {
	p.mut.Lock()
	defer p.mut.Unlock()
	p.basicAuth = basicAuth
	return p
}