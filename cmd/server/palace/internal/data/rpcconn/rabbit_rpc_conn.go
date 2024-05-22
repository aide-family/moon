package rpcconn

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	mmd "github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/types"
)

// NewRabbitRpcConn 创建一个rabbit rpc连接
func NewRabbitRpcConn(c *palaceconf.Bootstrap) (*RabbitRpcConn, func(), error) {
	microServer := c.GetMicroServer()
	rabbitServer := microServer.GetRabbitServer()
	endpoint := rabbitServer.GetEndpoint()

	// 创建路由 Filter：筛选版本号为"2.0.0"的实例
	nodeFilter := filter.Version("2.0.0")
	// 由于 gRPC 框架的限制，只能使用全局 balancer name 的方式来注入 selector
	selector.SetGlobalSelector(wrr.NewBuilder())
	opts := []kgrpc.ClientOption{
		kgrpc.WithMiddleware(
			mmd.Client(),
			jwt.Client(func(token *jwtv5.Token) (interface{}, error) {
				return []byte(c.GetMicroServer().GetRabbitServer().GetToken()), nil
			}),
		),
		kgrpc.WithEndpoint(endpoint),
		kgrpc.WithTimeout(rabbitServer.GetTimeout().AsDuration()),
		kgrpc.WithOptions(grpc.WithConnectParams(defaultGrpcConnectParam)),
		kgrpc.WithNodeFilter(nodeFilter),
	}

	if !types.IsNil(microServer.GetDiscovery()) {
		dis, err := conn.NewDiscovery(microServer.GetDiscovery())
		if err == nil {
			return nil, nil, err
		}
		opts = append(opts, kgrpc.WithDiscovery(dis))
	}

	grpcConn, err := kgrpc.DialInsecure(context.Background(), opts...)
	if err != nil {
		log.Errorw("连接rabbit rpc失败：", err)
		return nil, nil, err
	}
	// 退出时清理资源
	cleanup := func() {
		if grpcConn != nil {
			if err = grpcConn.Close(); err != nil {
				log.Errorw("关闭 reseller rpc 连接失败：", err)
			}
		}
		fmt.Println("关闭 reseller rpc连接已完成")
	}
	return &RabbitRpcConn{
		client: grpcConn,
	}, cleanup, nil
}

type RabbitRpcConn struct {
	client *grpc.ClientConn
}

var defaultGrpcConnectParam = grpc.ConnectParams{
	Backoff: backoff.Config{
		MaxDelay: 0,
	},
	MinConnectTimeout: 3 * time.Second,
}
