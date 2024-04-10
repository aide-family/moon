package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aide-family/moon/pkg/httpx"
)

var _ Datasource = (*PromDatasource)(nil)

const (
	prometheusApiV1Query = "/api/v1/query"
)

type PromDatasource struct {
	// 数据源类型
	category DatasourceName
	// 地址
	endpoint string
	// 基础认证
	basicAuth *BasicAuth
}

func (d *PromDatasource) WithBasicAuth(basicAuth *BasicAuth) Datasource {
	d.basicAuth = basicAuth
	return d
}

func (d *PromDatasource) GetBasicAuth() *BasicAuth {
	return d.basicAuth
}

// Query 调用数据源查询数据
//
//	curl 'https://<domain>/api/v1/query?query=go_memstats_sys_bytes&time=1704785907'
func (d *PromDatasource) Query(_ context.Context, expr string, duration int64) (*QueryResponse, error) {
	params := httpx.ParseQuery(map[string]any{
		"query": expr,
		"time":  duration,
	})

	hx := httpx.NewHttpX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})
	if d.basicAuth != nil {
		hx = hx.SetBasicAuth(d.basicAuth.Username, d.basicAuth.Password)
	}
	getResponse, err := hx.GET(fmt.Sprintf("%s%s?%s", d.endpoint, prometheusApiV1Query, params))
	if err != nil {
		return nil, err
	}
	defer getResponse.Body.Close()
	var allResp QueryResponse
	if err = json.NewDecoder(getResponse.Body).Decode(&allResp); err != nil {
		return nil, err
	}
	return &allResp, nil
}

func (d *PromDatasource) GetCategory() string {
	if d == nil {
		return ""
	}
	return string(d.category)
}

func (d *PromDatasource) GetEndpoint() string {
	if d == nil {
		return ""
	}
	return d.endpoint
}

// NewPrometheusDatasource 实例化数据源对象
func NewPrometheusDatasource(domain string) *PromDatasource {
	return &PromDatasource{
		category: PrometheusDatasource,
		endpoint: strings.TrimSuffix(domain, "/"),
	}
}
