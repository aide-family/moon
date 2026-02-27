// Package connect is a package for connecting to services.
package connect

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/server/middler"
)

func InitHTTPClient(c InitConfig, opts ...InitOption) (*http.Client, error) {
	cfg, err := NewInitConfig(c, opts...)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.ToUpper(cfg.protocol), ProtocolHTTP) {
		return nil, merr.ErrorInternalServer("protocol is not HTTP, got %s", cfg.protocol)
	}
	middlewares := []middleware.Middleware{
		recovery.Recovery(),
		middler.Validate(),
		metadata.Client(),
		middler.JwtClient(),
	}

	clientOpts := []http.ClientOption{
		http.WithEndpoint(cfg.endpoint),
		http.WithMiddleware(middlewares...),
	}

	if pointer.IsNotNil(cfg.discovery) {
		clientOpts = append(clientOpts, http.WithDiscovery(cfg.discovery), http.WithBlock())
		filterOpts := make([]selector.NodeFilter, 0, 2)
		filterOpts = append(filterOpts, SelectNodeFilterOr(cfg.nodeFilters...))
		if nodeVersion := strings.TrimSpace(cfg.nodeVersion); nodeVersion != "" {
			filterOpts = append(filterOpts, filter.Version(nodeVersion))
		}
		clientOpts = append(clientOpts, http.WithNodeFilter(filterOpts...))
	}

	if cfg.timeout > 0 {
		clientOpts = append(clientOpts, http.WithTimeout(cfg.timeout))
	}

	return http.NewClient(context.Background(), clientOpts...)
}
