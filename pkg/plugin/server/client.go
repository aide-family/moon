package server

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	ggrpc "google.golang.org/grpc"

	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/middler"
	"github.com/aide-family/moon/pkg/plugin/registry"
)

type InitConfig struct {
	MicroConfig *config.MicroServer
	Registry    *config.Registry
}

func InitHTTPClient(initConfig *InitConfig) (*http.Client, error) {
	if initConfig.MicroConfig.GetNetwork() != config.Network_HTTP {
		return nil, merr.ErrorInternalServerError("network is not http")
	}
	middlewares := []middleware.Middleware{
		recovery.Recovery(),
		middler.Validate(),
	}
	if initConfig.MicroConfig.GetSecret() != "" {
		middlewares = append(middlewares, jwt.Client(func(token *jwtv5.Token) (interface{}, error) {
			return []byte(initConfig.MicroConfig.GetSecret()), nil
		}))
	}

	opts := []http.ClientOption{
		http.WithEndpoint(initConfig.MicroConfig.GetEndpoint()),
		http.WithMiddleware(middlewares...),
	}

	nodeVersion := strings.TrimSpace(initConfig.MicroConfig.GetVersion())
	if nodeVersion != "" {
		nodeFilter := filter.Version(nodeVersion)
		selector.SetGlobalSelector(wrr.NewBuilder())
		opts = append(opts, http.WithNodeFilter(nodeFilter))
	}

	if initConfig.Registry != nil && initConfig.Registry.GetEnable() {
		var err error
		discovery, err := registry.NewDiscovery(initConfig.Registry)
		if err != nil {
			return nil, err
		}
		opts = append(opts, http.WithDiscovery(discovery))
	}

	if initConfig.MicroConfig.GetTimeout() != nil {
		opts = append(opts, http.WithTimeout(initConfig.MicroConfig.GetTimeout().AsDuration()))
	}

	return http.NewClient(context.Background(), opts...)
}

func InitGRPCClient(initConfig *InitConfig) (*ggrpc.ClientConn, error) {
	if initConfig.MicroConfig.GetNetwork() != config.Network_GRPC {
		return nil, merr.ErrorInternalServerError("network is not grpc")
	}
	middlewares := []middleware.Middleware{
		recovery.Recovery(),
		middler.Validate(),
	}
	if initConfig.MicroConfig.GetSecret() != "" {
		middlewares = append(middlewares, jwt.Client(func(token *jwtv5.Token) (interface{}, error) {
			return []byte(initConfig.MicroConfig.GetSecret()), nil
		}))
	}

	opts := []grpc.ClientOption{
		grpc.WithEndpoint(initConfig.MicroConfig.GetEndpoint()),
		grpc.WithMiddleware(middlewares...),
	}

	//nodeVersion := strings.TrimSpace(initConfig.MicroConfig.GetVersion())
	//if nodeVersion != "" {
	//	nodeFilter := filter.Version(nodeVersion)
	//	selector.SetGlobalSelector(wrr.NewBuilder())
	//	opts = append(opts, grpc.WithNodeFilter(nodeFilter))
	//}

	if initConfig.Registry != nil && initConfig.Registry.GetEnable() {
		var err error
		discovery, err := registry.NewDiscovery(initConfig.Registry)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithDiscovery(discovery))
	}

	if initConfig.MicroConfig.GetTimeout() != nil {
		opts = append(opts, grpc.WithTimeout(initConfig.MicroConfig.GetTimeout().AsDuration()))
	}

	return grpc.DialInsecure(context.Background(), opts...)
}
