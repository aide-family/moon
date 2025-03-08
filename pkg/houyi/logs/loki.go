package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	// LokiAPIV1Query is the API endpoint for querying logs in Loki.
	lokiAPIV1Query = "/loki/api/v1/query"
	// LokiAPIV1QueryRange is the API endpoint for querying logs in Loki with a time range.
	lokiAPIV1QueryRange = "/loki/api/v1/query_range"
)

type (

	// BasicAuth 基础认证信息
	BasicAuth struct {
		username string
		password string
	}

	// LokiQuery is a log query implementation for Loki.
	LokiQuery struct {
		// endpoint 数据源地址
		endpoint string
		// lokiAPIV1Query 查询日志
		lokiAPIV1Query string
		// lokiAPIV1QueryRange 查询日志区间
		lokiAPIV1QueryRange string
		// 日志排序顺序，支持的值为forward或backward，默认为backward
		Direction string
		// 要返回的最大条目数
		Limit     int64
		basicAuth *BasicAuth
	}

	// LokiOption is a functional option for LokiQuery.
	LokiOption func(l *LokiQuery)
)

func (l *LokiQuery) Check(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

// NewLokiDatasource sets the Loki endpoint.
func NewLokiDatasource(opt ...LokiOption) *LokiQuery {

	l := &LokiQuery{
		lokiAPIV1Query:      lokiAPIV1Query,
		lokiAPIV1QueryRange: lokiAPIV1QueryRange,
	}

	for _, o := range opt {
		o(l)
	}

	return l
}

// WithLokiEndpoint sets the Loki endpoint.
func WithLokiEndpoint(endpoint string) LokiOption {
	return func(l *LokiQuery) {
		l.endpoint = endpoint
	}
}

// WithLokiDirection sets the Loki query direction.
func WithLokiDirection(direction string) LokiOption {
	return func(l *LokiQuery) {
		l.Direction = direction
	}
}

// WithLokiLimit sets the Loki query limit.
func WithLokiLimit(limit int64) LokiOption {
	return func(l *LokiQuery) {
		l.Limit = limit
	}
}

// WithLokiBasicAuth sets the Loki basic auth.
func WithLokiBasicAuth(username, password string) LokiOption {
	return func(l *LokiQuery) {
		l.basicAuth = &BasicAuth{
			username: username,
			password: password,
		}
	}
}

func (l *LokiQuery) QueryLogs(ctx context.Context, expr string, start, end int64) (*datasource.LogResponse, error) {

	hx := httpx.NewHTTPX()
	hx.SetHeader(map[string]string{
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.9",
	})

	if types.TextIsNull(l.Direction) {
		l.Direction = "backward"
	}

	if l.Limit == 0 {
		l.Limit = 100
	}

	if types.IsNotNil(l.basicAuth) {
		hx = hx.SetBasicAuth(l.basicAuth.username, l.basicAuth.password)
	}

	params := httpx.ParseQuery(map[string]any{
		"query":     expr,
		"start":     start,
		"end":       end,
		"limit":     l.Limit,
		"direction": l.Direction,
	})

	api, err := url.JoinPath(l.endpoint, l.lokiAPIV1QueryRange)
	if types.IsNotNil(err) {
		return nil, err
	}

	url := fmt.Sprintf("%s?%s", api, params)

	log.Infof("loki query url: %s", url)

	getResponse, err := hx.GETWithContext(ctx, url)
	if types.IsNotNil(err) {
		return nil, err
	}

	defer getResponse.Body.Close()

	// 检查响应状态
	if getResponse.StatusCode != http.StatusOK {
		return nil, merr.ErrorNotificationSystemError("loki query error: %v", getResponse.Status)
	}

	// 解析响应数据
	var response datasource.LokiQueryResponse
	if err := json.NewDecoder(getResponse.Body).Decode(&response); err != nil {
		return nil, merr.ErrorNotificationSystemError("error decoding response: %v", err.Error())
	}

	var logRes datasource.LogResponse
	var values []string

	// 打印日志内容
	for _, stream := range response.Data.Result {
		for _, value := range stream.Values {
			pair := value.([]interface{})
			logLine := pair[1].(string)
			values = append(values, logLine)
		}
	}
	logRes.DatasourceUrl = l.endpoint
	logRes.Values = values
	logRes.Timestamp = time.Now().Unix()
	return &logRes, nil
}
