package logs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/datasource"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/elastic/go-elasticsearch/v8"
)

type (
	// Elasticsearch es日志存储
	Elasticsearch struct {
		// Elasticsearch client
		client *elasticsearch.Client
		// Elasticsearch configuration
		c *conf.Elasticsearch
		// 索引值
		searchIndex string
		// endpoint 数据源地址
		endpoint string
	}

	// EsOption Elasticsearch option
	EsOption func(e *Elasticsearch)
)

func (e *Elasticsearch) Check(_ context.Context) error {
	if types.IsNil(e.c) {
		return merr.ErrorNotificationSystemError("Elasticsearch 配置为空")
	}

	if types.TextIsNull(e.searchIndex) {
		return merr.ErrorNotificationSystemError("Elasticsearch 索引为空")
	}

	if types.IsNil(e.c.Endpoint) {
		return merr.ErrorNotificationSystemError("Elasticsearch endpoint 为空")
	}

	return nil
}

func NewElasticsearch(c *conf.Elasticsearch, opt ...EsOption) (*Elasticsearch, error) {

	e := &Elasticsearch{c: c, searchIndex: c.GetSearchIndex()}
	if err := e.init(); err != nil {
		return nil, err
	}

	for _, o := range opt {
		o(e)
	}
	return e, nil
}

// WithEsEndpoint sets the es endpoint.
func WithEsEndpoint(endpoint string) EsOption {
	return func(l *Elasticsearch) {
		l.endpoint = endpoint
	}
}

func (e *Elasticsearch) QueryLogs(ctx context.Context, expr string, _, _ int64) (*datasource.LogResponse, error) {
	if err := e.Check(ctx); types.IsNotNil(err) {
		return nil, err
	}

	response, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(e.searchIndex),
		e.client.Search.WithBody(strings.NewReader(expr)),
	)
	if types.IsNotNil(err) {
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, merr.ErrorNotificationSystemError("es query error: %v", response.String())
	}
	// 解析响应
	var result datasource.LogResponse
	var esRes datasource.EsResponse
	if err := json.NewDecoder(response.Body).Decode(&esRes); err != nil {
		return nil, merr.ErrorNotificationSystemError("error decoding the response: %v", err.Error())
	}

	if types.IsNil(esRes.Hits) {
		return &result, nil
	}

	resValue := make([]string, 0, len(esRes.Hits.Hits))
	for _, item := range esRes.Hits.Hits {
		resValue = append(resValue, types.ConvertString(item.Source))
	}
	result.Values = resValue
	result.DatasourceUrl = e.endpoint
	result.Timestamp = time.Now().Unix()
	return &result, nil
}

func (e *Elasticsearch) init() error {
	cfg := elasticsearch.Config{Addresses: strings.Split(e.c.GetEndpoint(), ",")}
	cloudId := e.c.GetCloudId()
	apiKey := e.c.GetApiKey()
	username := e.c.GetUsername()
	password := e.c.GetPassword()
	serviceToken := e.c.GetServiceToken()

	if types.TextNotNull(serviceToken) {
		client, err := elasticsearch.NewClient(cfg)
		if types.IsNotNil(err) {
			return err
		}
		e.client = client
	}

	if types.IsNotNil(cloudId) && types.TextNotNull(apiKey) {
		cfg.CloudID = cloudId
		cfg.APIKey = apiKey
		client, err := elasticsearch.NewClient(cfg)
		if types.IsNotNil(err) {
			return err
		}
		e.client = client
	}
	if types.TextNotNull(username) && types.TextNotNull(password) {
		cfg.Username = username
		cfg.Password = password
		cfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client, err := elasticsearch.NewClient(cfg)
		if types.IsNotNil(err) {
			return err
		}
		e.client = client
	}

	if types.IsNil(e.client) {
		return merr.ErrorNotificationSystemError("es client initialization configuration failed!,config:%v", e.c)
	}
	return nil
}
