package p8s

import (
	"strings"

	"github.com/aide-family/moon/pkg/agent"
)

type Option func(*PrometheusDatasource)

// WithEndpoint set endpoint
func WithEndpoint(endpoint string) Option {
	return func(p *PrometheusDatasource) {
		p.mut.Lock()
		defer p.mut.Unlock()
		// 去除 最后的 /
		p.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithBasicAuth set basic auth
func WithBasicAuth(basicAuth *agent.BasicAuth) Option {
	return func(p *PrometheusDatasource) {
		p.mut.Lock()
		defer p.mut.Unlock()
		p.basicAuth = basicAuth
	}
}
