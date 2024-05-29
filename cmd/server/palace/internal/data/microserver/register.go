package microserver

import (
	"context"
	"time"

	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	mmd "github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

var ProviderSetRpcConn = wire.NewSet(
	NewHouYiConn,
	NewRabbitRpcConn,
)

func newRpcConn(microServerConf *api.Server, discovery *api.Discovery) (*grpc.ClientConn, error) {
	// 创建路由 Filter：筛选版本号为"2.0.0"的实例
	nodeFilter := filter.Version("2.0.0")
	timeout := microServerConf.GetTimeout().AsDuration()
	endpoint := microServerConf.GetEndpoint()
	// 由于 gRPC 框架的限制，只能使用全局 balancer name 的方式来注入 selector
	selector.SetGlobalSelector(wrr.NewBuilder())
	opts := []kgrpc.ClientOption{
		kgrpc.WithMiddleware(
			mmd.Client(),
			jwt.Client(func(token *jwtv5.Token) (interface{}, error) {
				return []byte(microServerConf.GetToken()), nil
			}),
		),
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(timeout),
		kgrpc.WithOptions(grpc.WithConnectParams(defaultGrpcConnectParam)),
		kgrpc.WithNodeFilter(nodeFilter),
	}

	if !types.IsNil(discovery) {
		dis, err := conn.NewDiscovery(discovery)
		if err == nil {
			return nil, err
		}
		opts = append(opts, kgrpc.WithDiscovery(dis))
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	grpcConn, err := kgrpc.DialInsecure(ctx, opts...)
	if err != nil {
		log.Errorw("连接rpc失败：", err, "endpoint", endpoint)
		return nil, err
	}
	return grpcConn, nil
}

var defaultGrpcConnectParam = grpc.ConnectParams{
	Backoff: backoff.Config{
		MaxDelay: 0,
	},
	MinConnectTimeout: 3 * time.Second,
}
