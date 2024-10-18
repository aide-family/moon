package data

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/api"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	pushapi "github.com/aide-family/moon/api/rabbit/push"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/microserver"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"

	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"
)

// NewRabbitRPCConn 创建一个rabbit rpc连接
func NewRabbitRPCConn(c *palaceconf.Bootstrap) (*RabbitConn, error) {
	discoveryConf := c.GetDiscovery()
	return &RabbitConn{
		srvs:          NewSrvList(c.GetDependRabbit()),
		discoveryConf: discoveryConf,
	}, nil
}

// RabbitConn rabbit服务连接
type RabbitConn struct {
	lock sync.Mutex
	// 服务实例原始信息
	srvs          *SrvList
	discoveryConf *conf.Discovery
}

// checkSrvIsAlive 检查服务是否存活
func (l *RabbitConn) checkSrvIsAlive(ctx context.Context, srv *Srv) (err error) {
	defer func() {
		// 如果有错误，则移除服务
		if !types.IsNil(err) {
			l.srvs.deleteSrv(genSrvUniqueKey(srv.srvInfo))
		}
	}()
	// 判断服务注册时间是否大于10秒 （当前时间是否在10之后）
	if !time.Now().Before(srv.registerTime.Add(10 * time.Second)) {
		return nil
	}
	// 检测服务是否存活
	switch srv.network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		_, err = api.NewHealthHTTPClient(srv.httpClient).Check(ctx, &api.CheckRequest{})
	default:
		_, err = api.NewHealthClient(srv.rpcClient).Check(ctx, &api.CheckRequest{})
	}
	if !types.IsNil(err) {
		return merr.ErrorNotificationSystemError("Rabbit 服务不可用")
	}
	return
}

// NotifyObject 发送通道配置
func (l *RabbitConn) NotifyObject(ctx context.Context, in *pushapi.NotifyObjectRequest, opts ...microserver.Option) error {
	eg := new(errgroup.Group)
	for _, srv := range l.srvs.getSrvs() {
		conn := srv
		eg.Go(func() error {
			if err := l.checkSrvIsAlive(ctx, conn); !types.IsNil(err) {
				return err
			}
			switch srv.network {
			case vobj.NetworkHTTP, vobj.NetworkHTTPS:
				httpOpts := make([]http.CallOption, 0)
				for _, opt := range opts {
					httpOpts = append(httpOpts, opt.HTTPOpts...)
				}
				_, err := pushapi.NewConfigHTTPClient(srv.httpClient).NotifyObject(ctx, in, httpOpts...)
				return err
			default:
				rpcOpts := make([]grpc.CallOption, 0)
				for _, opt := range opts {
					rpcOpts = append(rpcOpts, opt.RPCOpts...)
				}
				_, err := pushapi.NewConfigClient(srv.rpcClient).NotifyObject(ctx, in, rpcOpts...)
				return err
			}
		})
	}
	return eg.Wait()
}

// SendMsg 发送消息
func (l *RabbitConn) SendMsg(ctx context.Context, in *hookapi.SendMsgRequest, opts ...microserver.Option) error {
	eg := new(errgroup.Group)
	for _, srv := range l.srvs.getSrvs() {
		conn := srv
		eg.Go(func() error {
			if err := l.checkSrvIsAlive(ctx, conn); !types.IsNil(err) {
				return err
			}
			switch conn.network {
			case vobj.NetworkHTTP, vobj.NetworkHTTPS:
				httpOpts := make([]http.CallOption, 0)
				for _, opt := range opts {
					httpOpts = append(httpOpts, opt.HTTPOpts...)
				}
				_, err := hookapi.NewHookHTTPClient(conn.httpClient).SendMsg(ctx, in, httpOpts...)
				return err
			default:
				rpcOpts := make([]grpc.CallOption, 0)
				for _, opt := range opts {
					rpcOpts = append(rpcOpts, opt.RPCOpts...)
				}
				_, err := hookapi.NewHookClient(conn.rpcClient).SendMsg(ctx, in, rpcOpts...)
				return err
			}
		})
	}
	return eg.Wait()
}

// Heartbeat 心跳
func (l *RabbitConn) Heartbeat(_ context.Context, req *api.HeartbeatRequest) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	// 存储心跳数据
	srvKey := genSrvUniqueKey(req.GetServer())
	_, ok := l.srvs.getSrv(srvKey)
	if !ok {
		return l.SrvRegister(srvKey, req.GetServer())
	}
	return nil
}

// SrvRegister 服务注册
func (l *RabbitConn) SrvRegister(key string, microServer *conf.MicroServer) error {
	network := vobj.ToNetwork(microServer.GetNetwork())
	srv := &Srv{
		srvInfo:      microServer,
		rpcClient:    nil,
		network:      network,
		httpClient:   nil,
		registerTime: time.Now(),
	}
	switch network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpConn, err := microserver.NewHTTPConn(microServer, l.discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接Rabbit http失败：", err)
			return err
		}
		srv.httpClient = httpConn
	default:
		grpcConn, err := microserver.NewRPCConn(microServer, l.discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接Rabbit rpc失败：", err)
			return err
		}
		srv.rpcClient = grpcConn
	}
	l.srvs.appendSrv(key, srv)
	return nil
}
