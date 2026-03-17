package impl

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

const proxyTimeout = 30 * time.Second

// NewMetricDatasourceProxy returns a MetricDatasourceProxy that forwards HTTP to the datasource.
func NewMetricDatasourceProxy() repository.MetricDatasourceProxy {
	return &metricDatasourceProxy{}
}

type metricDatasourceProxy struct{}

// Proxy implements repository.MetricDatasourceProxy.
func (p *metricDatasourceProxy) Proxy(ctx context.Context, ds *bo.DatasourceItemBo, path, method string, body []byte) (int, []byte, error) {
	if ds == nil || ds.Type != enum.DatasourceType_METRICS {
		return 0, nil, merr.ErrorInvalidArgument("datasource is not a metrics type")
	}
	baseURL := strings.TrimRight(ds.URL, "/")
	if baseURL == "" {
		return 0, nil, merr.ErrorInvalidArgument("datasource url is empty")
	}
	path = strings.TrimPrefix(path, "/")
	fullURL := baseURL + "/" + path

	client := &http.Client{Timeout: proxyTimeout}
	var req *http.Request
	var err error
	if len(body) > 0 && (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch) {
		req, err = http.NewRequestWithContext(ctx, method, fullURL, bytes.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, fullURL, nil)
	}
	if err != nil {
		return 0, nil, merr.ErrorInternalServer("create proxy request failed").WithCause(err)
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	if ds.Metadata != nil {
		if user, ok := ds.Metadata[metadataKeyBasicAuthUsername]; ok {
			if pass, ok := ds.Metadata[metadataKeyBasicAuthPassword]; ok && user != "" {
				req.SetBasicAuth(user, pass)
			}
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, merr.ErrorInternalServer("proxy request failed").WithCause(err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, merr.ErrorInternalServer("read proxy response failed").WithCause(err)
	}
	return resp.StatusCode, respBody, nil
}
