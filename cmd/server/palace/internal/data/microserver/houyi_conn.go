package microserver

import (
	"context"

	"github.com/aide-family/moon/api"
	metadataapi "github.com/aide-family/moon/api/houyi/metadata"
	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// NewHouYiConn 创建一个HouYi rpc连接
func NewHouYiConn(c *palaceconf.Bootstrap) (*HouYiConn, func(), error) {
	microServer := c.GetMicroServer()
	discoveryConf := c.GetDiscovery()
	if !types.IsNil(microServer.GetHouYiServers()) && len(microServer.GetHouYiServers()) > 0 {
		houYiConn := new(HouYiConn)
		cleanupList := make([]func(), 0, len(microServer.GetHouYiServers()))
		for _, houYiServer := range microServer.GetHouYiServers() {
			houYiConnInstance, cleanup1, err1 := newHouYiConn(discoveryConf, houYiServer)
			if !types.IsNil(err1) {
				return nil, nil, err1
			}
			cleanupList = append(cleanupList, cleanup1)
			houYiConn.houYiConns = append(houYiConn.houYiConns, houYiConnInstance)
		}
		cleanup := func() {
			for _, cleanup1 := range cleanupList {
				cleanup1()
			}
		}
		return houYiConn, cleanup, nil
	} else {
		houYiConn, cleanup, err := newHouYiConn(discoveryConf, microServer.GetHouYiServer())
		if !types.IsNil(err) {
			return nil, nil, err
		}
		return &HouYiConn{
			houYiConns: []*HouYiConn{houYiConn},
		}, cleanup, nil
	}
}

// newHouYiConn 创建一个HouYi rpc连接
func newHouYiConn(discoveryConf *conf.Discovery, houYiServer *conf.MicroServer) (*HouYiConn, func(), error) {
	houYiConn := &HouYiConn{}
	if types.IsNil(houYiServer) {
		return nil, nil, merr.ErrorNotification("未配置MicroServer.HouYiServer")
	}
	switch houYiServer.GetNetwork() {
	case "http", "HTTP":
		httpConn, err := newHTTPConn(houYiServer, discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接HouYi http失败：", err)
			return nil, nil, err
		}
		houYiConn.httpClient = httpConn
		houYiConn.network = vobj.NetworkHTTP
	case "https", "HTTPS":
		httpConn, err := newHTTPConn(houYiServer, discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接HouYi http失败：", err)
			return nil, nil, err
		}
		houYiConn.httpClient = httpConn
		houYiConn.network = vobj.NetworkHTTPS
	case "rpc", "RPC", "grpc", "GRPC":
		grpcConn, err := newRPCConn(houYiServer, discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接HouYi rpc失败：", err)
			return nil, nil, err
		}
		houYiConn.rpcClient = grpcConn
		houYiConn.network = vobj.NetworkRPC
	default:
		return nil, nil, merr.ErrorNotification("HouYi Server暂不支持该网络类型：[%s]", houYiServer.GetNetwork())
	}
	// 退出时清理资源
	cleanup := func() {
		if !types.IsNil(houYiConn.rpcClient) {
			if err := houYiConn.rpcClient.Close(); !types.IsNil(err) {
				log.Errorw("关闭 houYi rpc 连接失败：", err)
			}
		}
		if !types.IsNil(houYiConn.httpClient) {
			if err := houYiConn.httpClient.Close(); !types.IsNil(err) {
				log.Errorw("关闭 houYi http 连接失败：", err)
			}
		}
		log.Info("关闭 houYi rpc连接已完成")
	}

	return houYiConn, cleanup, nil
}

// HouYiConn HouYi服务连接
type HouYiConn struct {
	// rpc连接
	rpcClient *grpc.ClientConn
	// 网络请求类型
	network vobj.Network
	// http连接
	httpClient *http.Client

	// 多实例模式
	houYiConns []*HouYiConn
}

func getConn(conns []*HouYiConn) *HouYiConn {
	// TODO ...
	return conns[0]
}

// Sync 同步数据
func (l *HouYiConn) Sync(ctx context.Context, in *metadataapi.SyncMetadataRequest, opts ...Option) (*metadataapi.SyncMetadataReply, error) {
	// 轮询算法获取conn
	conn := getConn(l.houYiConns)
	switch conn.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HTTPOpts...)
		}
		return metadataapi.NewMetricHTTPClient(conn.httpClient).SyncMetadata(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RPCOpts...)
		}
		return metadataapi.NewMetricClient(conn.rpcClient).SyncMetadata(ctx, in, rpcOpts...)
	}
}

// SyncV2 同步数据
func (l *HouYiConn) SyncV2(ctx context.Context, in *metadataapi.SyncMetadataV2Request, opts ...Option) (*metadataapi.SyncMetadataV2Reply, error) {
	// 轮询算法获取conn
	conn := getConn(l.houYiConns)
	switch conn.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HTTPOpts...)
		}
		return metadataapi.NewMetricHTTPClient(conn.httpClient).SyncMetadataV2(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RPCOpts...)
		}
		return metadataapi.NewMetricClient(conn.rpcClient).SyncMetadataV2(ctx, in, rpcOpts...)
	}
}

// Query 查询数据
func (l *HouYiConn) Query(ctx context.Context, in *metadataapi.QueryRequest, opts ...Option) (*metadataapi.QueryReply, error) {
	conn := getConn(l.houYiConns)
	switch conn.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpOpts := make([]http.CallOption, 0)
		for _, opt := range opts {
			httpOpts = append(httpOpts, opt.HTTPOpts...)
		}
		return metadataapi.NewMetricHTTPClient(conn.httpClient).Query(ctx, in, httpOpts...)
	default:
		rpcOpts := make([]grpc.CallOption, 0)
		for _, opt := range opts {
			rpcOpts = append(rpcOpts, opt.RPCOpts...)
		}
		return metadataapi.NewMetricClient(conn.rpcClient).Query(ctx, in, rpcOpts...)
	}
}

// PushStrategy 推送策略
func (l *HouYiConn) PushStrategy(ctx context.Context, in *strategyapi.PushStrategyRequest, opts ...Option) (*strategyapi.PushStrategyReply, error) {
	eg := new(errgroup.Group)
	for _, connItem := range l.houYiConns {
		conn := connItem
		eg.Go(func() error {
			var err error
			switch conn.network {
			case vobj.NetworkHTTP, vobj.NetworkHTTPS:
				httpOpts := make([]http.CallOption, 0)
				for _, opt := range opts {
					httpOpts = append(httpOpts, opt.HTTPOpts...)
				}
				_, err = strategyapi.NewStrategyHTTPClient(conn.httpClient).PushStrategy(ctx, in, httpOpts...)
			default:
				rpcOpts := make([]grpc.CallOption, 0)
				for _, opt := range opts {
					rpcOpts = append(rpcOpts, opt.RPCOpts...)
				}
				_, err = strategyapi.NewStrategyClient(conn.rpcClient).PushStrategy(ctx, in, rpcOpts...)
			}
			return err
		})
	}

	return &strategyapi.PushStrategyReply{}, eg.Wait()
}

// Health 健康检查
func (l *HouYiConn) Health(ctx context.Context, req *api.CheckRequest) (*api.CheckReply, error) {
	eg := new(errgroup.Group)
	for _, connItem := range l.houYiConns {
		conn := connItem
		eg.Go(func() error {
			var err error
			switch conn.network {
			case vobj.NetworkHTTP, vobj.NetworkHTTPS:
				_, err = api.NewHealthHTTPClient(conn.httpClient).Check(ctx, req)
			default:
				_, err = api.NewHealthClient(conn.rpcClient).Check(ctx, req)
			}
			return err
		})
	}
	return &api.CheckReply{}, eg.Wait()
}
