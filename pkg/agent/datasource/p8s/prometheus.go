package p8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/aide-family/moon/pkg/agent"
	"github.com/aide-family/moon/pkg/httpx"
)

const (
	prometheusApiV1Query = "/api/v1/query"
)

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

func NewPrometheusDatasource(opts ...Option) agent.Datasource {
	p := &PrometheusDatasource{
		category: agent.Prometheus,
	}
	for _, opt := range opts {
		opt(p)
	}

	return p
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
