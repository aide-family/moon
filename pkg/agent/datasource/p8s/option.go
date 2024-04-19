package p8s

import (
	"github.com/aide-family/moon/pkg/agent/datasource"
)

type Option func(*PrometheusDatasource)

// WithEndpoint set endpoint
func WithEndpoint(endpoint string) Option {
	return func(p *PrometheusDatasource) {
		p.endpoint = endpoint
	}
}

// WithBasicAuth set basic auth
func WithBasicAuth(basicAuth *datasource.BasicAuth) Option {
	return func(p *PrometheusDatasource) {
		p.basicAuth = basicAuth
	}
}
