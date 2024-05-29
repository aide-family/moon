package microserver

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"

	hookapi "github.com/aide-cloud/moon/api/rabbit/hook"
	pushapi "github.com/aide-cloud/moon/api/rabbit/push"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-cloud/moon/pkg/vobj"
)

// NewRabbitRpcConn 创建一个rabbit rpc连接
func NewRabbitRpcConn(c *palaceconf.Bootstrap) (*RabbitRpcConn, func(), error) {
	var err error
	var grpcConn *grpc.ClientConn
	microServer := c.GetMicroServer()
	rabbitServer := microServer.GetRabbitServer()
	endpoint := rabbitServer.GetEndpoint()
	network := vobj.NetworkRpc
	httpEndpoint := endpoint
	if strings.HasPrefix(endpoint, "http") {
		network = vobj.NetworkHttp
	} else if strings.HasPrefix(endpoint, "https") {
		network = vobj.NetworkHttps
	} else {
		network = vobj.NetworkRpc
		grpcConn, err = newRpcConn(rabbitServer, microServer.GetDiscovery())
		if err != nil {
			log.Errorw("连接HouYi rpc失败：", err)
			return nil, nil, err
		}
	}

	// 退出时清理资源
	cleanup := func() {
		if grpcConn != nil {
			if err = grpcConn.Close(); err != nil {
				log.Errorw("关闭 reseller rpc 连接失败：", err)
			}
		}
		log.Info("关闭 reseller rpc连接已完成")
	}
	return &RabbitRpcConn{
		client:       grpcConn,
		rpcClient:    grpcConn,
		httpEndpoint: httpEndpoint,
		network:      network,
	}, cleanup, nil
}

var _ hookapi.HookClient = (*RabbitRpcConn)(nil)
var _ pushapi.ConfigClient = (*RabbitRpcConn)(nil)

type RabbitRpcConn struct {
	client *grpc.ClientConn
	// rpc连接
	rpcClient *grpc.ClientConn
	// http请求地址
	httpEndpoint string
	// 网络请求类型
	network vobj.Network
}

// NotifyObject 发送通道配置
func (l *RabbitRpcConn) NotifyObject(ctx context.Context, in *pushapi.NotifyObjectRequest, opts ...grpc.CallOption) (*pushapi.NotifyObjectReply, error) {
	//TODO implement me
	panic("implement me")
}

// SendMsg 发送消息
func (l *RabbitRpcConn) SendMsg(ctx context.Context, in *hookapi.SendMsgRequest, opts ...grpc.CallOption) (*hookapi.SendMsgReply, error) {
	//TODO implement me
	panic("implement me")
}
