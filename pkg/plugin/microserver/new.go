package microserver

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	mmd "github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

func NewRPCConn(microServerConf *conf.MicroServer, discovery *conf.Discovery) (*grpc.ClientConn, error) {
	timeout := microServerConf.GetTimeout().AsDuration()
	endpoint := microServerConf.GetEndpoint()
	// 由于 gRPC 框架的限制，只能使用全局 balancer name 的方式来注入 selector
	selector.SetGlobalSelector(wrr.NewBuilder())
	maxMsgSize := 1024 * 1024 * 50 // 50 MB
	opts := []kgrpc.ClientOption{
		kgrpc.WithMiddleware(
			recovery.Recovery(),
			mmd.Client(),
			jwt.Client(func(token *jwtv5.Token) (interface{}, error) {
				return []byte(microServerConf.GetSecret()), nil
			}),
		),
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(timeout),
		kgrpc.WithOptions(grpc.WithConnectParams(defaultGrpcConnectParam)),
		kgrpc.WithOptions(
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(maxMsgSize),
				grpc.MaxCallRecvMsgSize(maxMsgSize),
			),
		),
	}
	//if !types.TextIsNull(microServerConf.GetNodeVersion()) {
	//	// 创建路由 Filter：筛选版本号为"2.0.0"的实例
	//	nodeFilter := filter.Version(microServerConf.GetNodeVersion())
	//	opts = append(opts, kgrpc.WithNodeFilter(nodeFilter))
	//}

	if !types.IsNil(discovery) {
		dis, err := conn.NewDiscovery(discovery, conn.WithDiscoveryConfigEtcd(discovery.GetEtcd()))
		if !types.IsNil(err) {
			return nil, err
		}
		opts = append(opts, kgrpc.WithDiscovery(dis))
	}

	grpcConn, err := kgrpc.DialInsecure(context.Background(), opts...)
	if !types.IsNil(err) {
		log.Errorw("连接rpc失败：", err, "endpoint", endpoint)
		return nil, err
	}
	return grpcConn, nil
}

func NewHTTPConn(microServerConf *conf.MicroServer, discovery *conf.Discovery) (*http.Client, error) {
	timeout := microServerConf.GetTimeout().AsDuration()
	endpoint := microServerConf.GetEndpoint()
	opts := []http.ClientOption{
		http.WithEndpoint(endpoint),
		http.WithMiddleware(
			recovery.Recovery(),
			mmd.Client(),
			jwt.Client(func(token *jwtv5.Token) (interface{}, error) {
				return []byte(microServerConf.GetSecret()), nil
			}),
		),
		http.WithTimeout(timeout),
	}
	if !types.TextIsNull(microServerConf.GetNodeVersion()) {
		// 创建路由 Filter：筛选版本号为"2.0.0"的实例
		nodeFilter := filter.Version(microServerConf.GetNodeVersion())
		opts = append(opts, http.WithNodeFilter(nodeFilter))
	}

	if !types.IsNil(discovery) {
		dis, err := conn.NewDiscovery(discovery, conn.WithDiscoveryConfigEtcd(discovery.GetEtcd()))
		if !types.IsNil(err) {
			return nil, err
		}
		opts = append(opts, http.WithDiscovery(dis))
	}

	ctx := context.Background()
	newClient, err := http.NewClient(ctx, opts...)
	if !types.IsNil(err) {
		log.Errorw("连接http失败：", err, "endpoint", endpoint)
		return nil, err
	}
	return newClient, nil
}

var defaultGrpcConnectParam = grpc.ConnectParams{
	Backoff:           backoff.DefaultConfig,
	MinConnectTimeout: 3 * time.Second,
}

func init() {
	defaultGrpcConnectParam.Backoff.MaxDelay = 0
}

// Option grpc/http option
type Option struct {
	RPCOpts  []grpc.CallOption
	HTTPOpts []http.CallOption
}
