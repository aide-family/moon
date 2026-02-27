package datasource

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/aide-family/magicbox/httpx"
)

type MetricConfig interface {
	GetEndpoint() string
	GetHeaders() http.Header
	GetMethod() httpx.Method
	GetBasicAuth() *httpx.BasicAuth
	GetTLS() *tls.ConnectionState
	GetCA() string
	GetScrapeInterval() time.Duration
}
