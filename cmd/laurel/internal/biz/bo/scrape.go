package bo

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/pkg/util/httpx"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/validate"
)

type ScrapeTarget struct {
	JobName     string
	Target      string
	Labels      kv.StringMap
	Interval    time.Duration
	Timeout     time.Duration
	BasicAuth   *conf.BasicAuth
	TLS         *conf.TLS
	Params      kv.StringMap
	Headers     kv.StringMap
	MetricsPath string
	Scheme      string
}

func (s *ScrapeTarget) Do(ctx context.Context) (*http.Response, error) {
	api, err := s.api()
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	for key, value := range s.Params {
		params.Add(key, value)
	}
	hx := s.client(ctx)
	if s.Timeout > 0 {
		ctx, cancel := context.WithTimeout(ctx, s.Timeout)
		defer cancel()
		hx = hx.WithContext(ctx)
	}
	return hx.Get(api, params)
}

func (s *ScrapeTarget) api() (string, error) {
	if s.Scheme == "" {
		s.Scheme = "http"
	}
	if validate.IsNotNil(s.TLS) {
		s.Scheme = "https"
	}
	u, err := url.Parse(s.Scheme + "://" + s.Target)
	if err != nil {
		return "", err
	}
	u.Path = "metrics"
	if s.MetricsPath != "" {
		u.Path = s.MetricsPath
	}
	return u.String(), nil
}

func (s *ScrapeTarget) client(ctx context.Context) httpx.Client {
	hx := httpx.NewClient().WithContext(ctx)
	hx = hx.WithHeader(http.Header{
		"Accept":          []string{"*/*"},
		"Accept-Language": []string{"zh-CN,zh;q=0.9"},
	})

	// Configure TLS if available
	if tls := s.TLS; validate.IsNotNil(tls) {
		hx = hx.WithTLSClientConfig(tls.GetClientCert(), tls.GetClientKey())
		if serverName := tls.GetServerName(); serverName != "" {
			hx = hx.WithServerName(serverName)
		}
		// Configure CA if available
		if ca := tls.GetCa(); validate.TextIsNotNull(ca) {
			hx = hx.WithRootCA(ca)
		}
	}

	// Configure authentication if available
	if basicAuth := s.BasicAuth; validate.IsNotNil(basicAuth) {
		hx = hx.WithBasicAuth(basicAuth.GetUsername(), basicAuth.GetPassword())
	}

	// Configure custom headers if available
	if headers := s.Headers; len(headers) > 0 {
		for key, value := range headers {
			hx = hx.WithHeader(http.Header{key: []string{value}})
		}
	}
	return hx
}
