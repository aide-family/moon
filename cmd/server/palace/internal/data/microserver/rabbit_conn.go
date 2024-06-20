package microserver

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	pushapi "github.com/aide-family/moon/api/rabbit/push"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"
)

// NewRabbitRpcConn 创建一个rabbit rpc连接
func NewRabbitRpcConn(c *palaceconf.Bootstrap) (*RabbitConn, func(), error) {
	microServer := c.GetMicroServer()
	rabbitServer := microServer.GetRabbitServer()
	rabbitConn := &RabbitConn{}
	if types.IsNil(rabbitServer) {
		return nil, nil, merr.ErrorNotification("未配置MicroServer.RabbitServer")
	}
	switch rabbitServer.GetNetwork() {
	case "http", "HTTP":
		httpConn, err := newHttpConn(rabbitServer, microServer.GetDiscovery())
		if !types.IsNil(err) {
			log.Errorw("连接HouYi http失败：", err)
			return nil, nil, err
		}
		rabbitConn.httpClient = httpConn
		rabbitConn.network = vobj.NetworkHttp
	case "https", "HTTPS":
		httpConn, err := newHttpConn(rabbitServer, microServer.GetDiscovery())
		if !types.IsNil(err) {
			log.Errorw("连接HouYi http失败：", err)
			return nil, nil, err
		}
		rabbitConn.httpClient = httpConn
		rabbitConn.network = vobj.NetworkHttps
	case "rpc", "RPC", "grpc", "GRPC":
		grpcConn, err := newRpcConn(rabbitServer, microServer.GetDiscovery())
		if !types.IsNil(err) {
			log.Errorw("连接HouYi rpc失败：", err)
			return nil, nil, err
		}
		rabbitConn.rpcClient = grpcConn
		rabbitConn.network = vobj.NetworkRpc
	default:
		return nil, nil, merr.ErrorNotification("Rabbit Server暂不支持该网络类型：[%s]", rabbitServer.GetNetwork())
	}

	// 退出时清理资源
	cleanup := func() {
		if rabbitConn.rpcClient != nil {
			if err := rabbitConn.rpcClient.Close(); !types.IsNil(err) {
				log.Errorw("关闭 rabbit rpc 连接失败：", err)
			}
		}
		if rabbitConn.httpClient != nil {
			if err := rabbitConn.httpClient.Close(); !types.IsNil(err) {
				log.Errorw("关闭 rabbit http 连接失败：", err)
			}
		}
		log.Info("关闭 rabbit rpc连接已完成")
	}
	return rabbitConn, cleanup, nil
}

type RabbitConn struct {
	// rpc连接
	rpcClient *grpc.ClientConn
	// 网络请求类型
	network vobj.Network
	// http连接
	httpClient *http.Client
}

// NotifyObject 发送通道配置
func (l *RabbitConn) NotifyObject(ctx context.Context, in *pushapi.NotifyObjectRequest, opts ...Option) (*pushapi.NotifyObjectReply, error) {
	switch l.network {
	case vobj.NetworkHttp, vobj.NetworkHttps:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HttpOpts...)
		}
		return pushapi.NewConfigHTTPClient(l.httpClient).NotifyObject(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RpcOpts...)
		}
		return pushapi.NewConfigClient(l.rpcClient).NotifyObject(ctx, in, rpcOpts...)
	}
}

// SendMsg 发送消息
func (l *RabbitConn) SendMsg(ctx context.Context, in *hookapi.SendMsgRequest, opts ...Option) (*hookapi.SendMsgReply, error) {
	switch l.network {
	case vobj.NetworkHttp, vobj.NetworkHttps:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HttpOpts...)
		}
		return hookapi.NewHookHTTPClient(l.httpClient).SendMsg(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RpcOpts...)
		}
		return hookapi.NewHookClient(l.rpcClient).SendMsg(ctx, in, rpcOpts...)
	}
}
