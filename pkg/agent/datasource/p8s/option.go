package p8s

import (
	"github.com/aide-family/moon/pkg/agent"
)

type Option func(*PrometheusDatasource)

// WithEndpoint set endpoint
func WithEndpoint(endpoint string) Option {
	return func(p *PrometheusDatasource) {
		p.endpoint = endpoint
	}
}

// WithBasicAuth set basic auth
func WithBasicAuth(basicAuth *agent.BasicAuth) Option {
	return func(p *PrometheusDatasource) {
		p.basicAuth = basicAuth
	}
}
