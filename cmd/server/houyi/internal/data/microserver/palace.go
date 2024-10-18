package microserver

import (
	"context"

	"github.com/aide-family/moon/api"
	datasourceapi "github.com/aide-family/moon/api/admin/datasource"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/microserver"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"
)

// NewPalaceConn 创建一个palace rpc连接
func NewPalaceConn(c *houyiconf.Bootstrap) (*PalaceConn, func(), error) {
	if !c.GetDependPalace() {
		return nil, func() {}, nil
	}
	palaceServer := c.GetPalaceServer()
	discoveryConf := c.GetDiscovery()

	if types.IsNil(palaceServer) {
		return nil, nil, merr.ErrorNotification("未配置MicroServer.PalaceServer")
	}
	network := vobj.ToNetwork(palaceServer.GetNetwork())
	palaceConn := &PalaceConn{
		rpcClient:  nil,
		network:    network,
		httpClient: nil,
	}
	switch network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpConn, err := microserver.NewHTTPConn(palaceServer, discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接Palace http失败：", err)
			return nil, nil, err
		}
		palaceConn.httpClient = httpConn
	default:
		grpcConn, err := microserver.NewRPCConn(palaceServer, discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接Palace rpc失败：", err)
			return nil, nil, err
		}
		palaceConn.rpcClient = grpcConn
	}
	// 退出时清理资源
	cleanup := func() {
		if !types.IsNil(palaceConn.rpcClient) {
			if err := palaceConn.rpcClient.Close(); !types.IsNil(err) {
				log.Errorw("关闭 palace rpc 连接失败：", err)
			}
		}
		if !types.IsNil(palaceConn.httpClient) {
			if err := palaceConn.httpClient.Close(); !types.IsNil(err) {
				log.Errorw("关闭 palace http 连接失败：", err)
			}
		}
		log.Info("关闭 palace rpc连接已完成")
	}

	return palaceConn, cleanup, nil
}

// PalaceConn Palace服务连接
type PalaceConn struct {
	// rpc连接
	rpcClient *grpc.ClientConn
	// 网络请求类型
	network vobj.Network
	// http连接
	httpClient *http.Client
}

// PushMetric 向palace推送指标数据
func (l *PalaceConn) PushMetric(ctx context.Context, in *datasourceapi.SyncMetricRequest, opts ...microserver.Option) (*datasourceapi.SyncMetricReply, error) {
	switch l.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HTTPOpts...)
		}
		return datasourceapi.NewMetricHTTPClient(l.httpClient).SyncMetric(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RPCOpts...)
		}
		return datasourceapi.NewMetricClient(l.rpcClient).SyncMetric(ctx, in, rpcOpts...)
	}
}

// PushAlarm 向palace推送告警数据
func (l *PalaceConn) PushAlarm(ctx context.Context, in *api.AlarmItem, opts ...microserver.Option) (*api.HookReply, error) {
	switch l.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HTTPOpts...)
		}
		return api.NewAlertHTTPClient(l.httpClient).Hook(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RPCOpts...)
		}
		return api.NewAlertClient(l.rpcClient).Hook(ctx, in, rpcOpts...)
	}
}

// Heartbeat 向palace发送心跳
func (l *PalaceConn) Heartbeat(ctx context.Context, in *api.HeartbeatRequest, opts ...microserver.Option) (*api.HeartbeatReply, error) {
	switch l.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HTTPOpts...)
		}
		return api.NewServerHTTPClient(l.httpClient).Heartbeat(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RPCOpts...)
		}
		return api.NewServerClient(l.rpcClient).Heartbeat(ctx, in, rpcOpts...)
	}
}
