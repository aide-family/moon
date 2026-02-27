package connect

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	kGrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/server/middler"
)

func InitGRPCClient(c InitConfig, opts ...InitOption) (*grpc.ClientConn, error) {
	cfg, err := NewInitConfig(c, opts...)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.ToUpper(cfg.protocol), ProtocolGRPC) {
		return nil, merr.ErrorInternalServer("protocol is not GRPC, got %s", cfg.protocol)
	}
	middlewares := []middleware.Middleware{
		recovery.Recovery(),
		middler.Validate(),
		metadata.Client(),
		middler.JwtClient(),
	}

	clientOpts := []kGrpc.ClientOption{
		kGrpc.WithEndpoint(cfg.endpoint),
		kGrpc.WithMiddleware(middlewares...),
	}

	if pointer.IsNotNil(cfg.discovery) {
		clientOpts = append(clientOpts, kGrpc.WithDiscovery(cfg.discovery), kGrpc.WithPrintDiscoveryDebugLog(false))
		filterOpts := make([]selector.NodeFilter, 0, 2)
		filterOpts = append(filterOpts, SelectNodeFilterOr(cfg.nodeFilters...))
		if nodeVersion := strings.TrimSpace(cfg.nodeVersion); nodeVersion != "" {
			filterOpts = append(filterOpts, filter.Version(nodeVersion))
		}
		clientOpts = append(clientOpts, kGrpc.WithNodeFilter(filterOpts...))
	}

	if cfg.timeout > 0 {
		clientOpts = append(clientOpts, kGrpc.WithTimeout(cfg.timeout))
	}

	return kGrpc.DialInsecure(context.Background(), clientOpts...)
}
