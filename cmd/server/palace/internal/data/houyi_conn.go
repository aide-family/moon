package data

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/api"
	metadataapi "github.com/aide-family/moon/api/houyi/metadata"
	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/plugin/microserver"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

// NewHouYiConn 创建一个HouYi rpc连接
func NewHouYiConn(c *palaceconf.Bootstrap, data *Data) (*HouYiConn, func(), error) {
	discoveryConf := c.GetDiscovery()
	conn := &HouYiConn{
		data:          data,
		srvs:          NewSrvList(c.GetDependHouYi()),
		discoveryConf: discoveryConf,
	}
	// 退出时清理资源
	cleanup := func() {
		for _, srv := range conn.srvs.getSrvs() {
			if !types.IsNil(srv.rpcClient) {
				if err := srv.rpcClient.Close(); !types.IsNil(err) {
					log.Errorw("关闭 palace rpc 连接失败：", err)
				}
			}
			if !types.IsNil(srv.httpClient) {
				if err := srv.httpClient.Close(); !types.IsNil(err) {
					log.Errorw("关闭 palace http 连接失败：", err)
				}
			}
		}
		log.Info("关闭 Houyi 连接已完成")
	}
	return conn, cleanup, nil
}

// HouYiConn HouYi服务连接
type HouYiConn struct {
	data *Data
	lock sync.Mutex
	// 服务实例原始信息
	srvs          *SrvList
	discoveryConf *conf.Discovery
	teamIds       []uint32
}

// GetServerList 获取houyi服务列表
func (l *HouYiConn) GetServerList() (*api.GetServerListReply, error) {
	var list []*api.ServerItem
	for _, conn := range l.srvs.getSrvs() {
		var httpEndpoint string
		var grpcEndpoint string
		if conn.srvInfo.Network == "http" || conn.srvInfo.Network == "https" {
			httpEndpoint = conn.srvInfo.Endpoint
		} else if conn.srvInfo.Network == "rpc" {
			grpcEndpoint = conn.srvInfo.Endpoint
		}
		now := time.Now().Unix()
		firstRegisterTime := conn.firstRegisterTime.Unix()
		upTime := time.Unix(now, 0).Sub(time.Unix(firstRegisterTime, 0)).String()
		list = append(list, &api.ServerItem{
			Version: conn.srvInfo.NodeVersion,
			Server: &conf.Server{
				Name:         conn.srvInfo.Name,
				HttpEndpoint: httpEndpoint,
				GrpcEndpoint: grpcEndpoint,
				Network:      conn.srvInfo.Network,
				StartTime:    types.NewTime(conn.firstRegisterTime).String(),
				UpTime:       upTime,
			},
		})
	}
	return &api.GetServerListReply{
		List: list,
	}, nil
}

func (l *HouYiConn) setTeamIds(teamIds []uint32) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.teamIds = teamIds
}

// Sync 同步数据
func (l *HouYiConn) Sync(ctx context.Context, in *metadataapi.SyncMetadataRequest, opts ...microserver.Option) (*metadataapi.SyncMetadataReply, error) {
	// 轮询算法获取conn
	for _, conn := range l.srvs.getSrvs() {
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
	return nil, merr.ErrorNotificationSystemError("服务不可用")
}

// SyncV2 同步数据
func (l *HouYiConn) SyncV2(ctx context.Context, in *metadataapi.SyncMetadataV2Request, opts ...microserver.Option) (*metadataapi.SyncMetadataV2Reply, error) {
	// 轮询算法获取conn
	for _, conn := range l.srvs.getSrvs() {
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
	return nil, merr.ErrorNotificationSystemError("服务不可用")
}

// Query 查询数据
func (l *HouYiConn) Query(ctx context.Context, in *metadataapi.QueryRequest, opts ...microserver.Option) (*metadataapi.QueryReply, error) {
	for _, conn := range l.srvs.getSrvs() {
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
	return nil, merr.ErrorNotificationSystemError("服务不可用")
}

// PushStrategy 推送策略
func (l *HouYiConn) PushStrategy(ctx context.Context, in *strategyapi.PushStrategyRequest, opts ...microserver.Option) (*strategyapi.PushStrategyReply, error) {
	eg := new(errgroup.Group)
	for _, connItem := range l.srvs.getSrvs() {
		conn := connItem
		eg.Go(func() error {
			_, err := l.pushStrategy(ctx, conn, in, opts...)
			return err
		})
	}

	return &strategyapi.PushStrategyReply{}, eg.Wait()
}

// PushStrategy 推送策略
func (l *HouYiConn) pushStrategy(ctx context.Context, conn *Srv, in *strategyapi.PushStrategyRequest, opts ...microserver.Option) (*strategyapi.PushStrategyReply, error) {
	teamIDMap := types.ToMap(conn.teamIds, func(t uint32) uint32 {
		return t
	})

	if len(teamIDMap) > 0 {
		in.Strategies = types.Filter(in.Strategies, func(item *api.MetricStrategyItem) bool {
			return teamIDMap[item.TeamID] > 0
		})

		in.MqStrategies = types.Filter(in.MqStrategies, func(item *api.EventStrategyItem) bool {
			return teamIDMap[item.TeamID] > 0
		})

		in.DomainStrategies = types.Filter(in.DomainStrategies, func(item *api.DomainStrategyItem) bool {
			return teamIDMap[item.TeamID] > 0
		})

		in.HttpStrategies = types.Filter(in.HttpStrategies, func(item *api.HttpStrategyItem) bool {
			return teamIDMap[item.TeamID] > 0
		})

		in.PingStrategies = types.Filter(in.PingStrategies, func(item *api.PingStrategyItem) bool {
			return teamIDMap[item.TeamID] > 0
		})
	}

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
	return &strategyapi.PushStrategyReply{}, err
}

// Health 健康检查
func (l *HouYiConn) Health(ctx context.Context, req *api.CheckRequest) (*api.CheckReply, error) {
	eg := new(errgroup.Group)
	for _, connItem := range l.srvs.getSrvs() {
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

// Heartbeat 心跳
func (l *HouYiConn) Heartbeat(_ context.Context, req *api.HeartbeatRequest) error {
	// 存储心跳数据
	srvKey := genSrvUniqueKey(req.GetServer())
	if !req.GetOnline() {
		l.srvs.removeSrv(srvKey)
		return nil
	}
	_, ok := l.srvs.getSrv(srvKey, true)
	if ok {
		return nil
	}
	srv, err := l.srvRegister(srvKey, req.GetServer(), req.GetTeamIds())
	if !types.IsNil(err) {
		log.Errorw("method", "srvRegister", "err", err)
		return err
	}

	go func() {
		defer after.RecoverX()
		strategiesCh, err := l.getStrategies(srv)
		if !types.IsNil(err) {
			log.Errorw("获取策略失败", err)
			return
		}
		for strategies := range strategiesCh {
			if err := l.syncStrategies(srv, strategies); err != nil {
				return
			}
		}
	}()
	return nil
}

func (l *HouYiConn) syncStrategies(srv *Srv, strategies []*bo.Strategy) error {
	if len(strategies) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	items := builder.NewParamsBuild(ctx).StrategyModuleBuilder().BoStrategyBuilder().ToAPIs(strategies)
	_, err := l.pushStrategy(ctx, srv, &strategyapi.PushStrategyRequest{Strategies: items})
	if !types.IsNil(err) {
		log.Errorw("同步策略失败：", err)
		return err
	}
	return nil
}

func (l *HouYiConn) getStrategies(srv *Srv) (<-chan []*bo.Strategy, error) {
	ctx := context.Background()
	// 查询所有团队
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	wheres := []gen.Condition{mainQuery.SysTeam.Status.Eq(vobj.StatusEnable.GetValue())}
	if len(srv.teamIds) > 0 {
		wheres = append(wheres, mainQuery.SysTeam.ID.In(srv.teamIds...))
	}
	teamList, err := mainQuery.SysTeam.WithContext(ctx).Where(wheres...).Find()
	if !types.IsNil(err) {
		return nil, err
	}
	ch := make(chan []*bo.Strategy, len(teamList))
	go func() {
		defer after.RecoverX()
		for _, teamItem := range teamList {
			ctx = middleware.WithTeamIDContextKey(ctx, teamItem.ID)
			bizDB, err := l.data.GetBizGormDB(teamItem.ID)
			if !types.IsNil(err) {
				log.Errorw("获取业务数据库失败：", err, "teamId", teamItem.ID)
				continue
			}
			bizQuery := bizquery.Use(bizDB)
			log.Infof("开始获取【%s】需要同步的策略", teamItem.Name)
			// 关联查询等级等明细信息
			strategies, err := bizQuery.Strategy.WithContext(ctx).Unscoped().
				Where(bizQuery.Strategy.Status.Eq(vobj.StatusEnable.GetValue())).
				Preload(field.Associations).
				Preload(bizQuery.Strategy.AlarmNoticeGroups).
				Find()
			if !types.IsNil(err) {
				log.Errorw("查询策略失败：", err, "teamId", teamItem.ID)
				continue
			}
			list := make([]*bo.Strategy, 0, len(strategies)*5)
			for _, strategy := range strategies {
				items := builder.NewParamsBuild(ctx).StrategyModuleBuilder().DoStrategyBuilder().ToBos(strategy)
				if len(items) == 0 {
					continue
				}
				list = append(list, items...)
			}
			ch <- list
			log.Infow("团队【"+teamItem.Name+"】策略", len(list), "teamId", teamItem.ID)
		}
		for len(ch) > 0 {
			time.Sleep(time.Millisecond * 100)
		}
		close(ch)
	}()
	return ch, nil
}

// srvRegister 服务注册
func (l *HouYiConn) srvRegister(key string, microServer *conf.MicroServer, teamIds []uint32) (*Srv, error) {
	srv, ok := l.srvs.getSrv(key)
	if ok {
		return srv, nil
	}
	network := vobj.ToNetwork(microServer.GetNetwork())
	srv = &Srv{
		srvInfo:      microServer,
		teamIds:      teamIds,
		rpcClient:    nil,
		network:      network,
		httpClient:   nil,
		registerTime: time.Now(),
	}
	switch network {
	case vobj.NetworkHTTP, vobj.NetworkHTTPS:
		httpConn, err := microserver.NewHTTPConn(microServer, l.discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接HouYi http失败", err)
			return nil, err
		}
		srv.httpClient = httpConn
	default:
		grpcConn, err := microserver.NewRPCConn(microServer, l.discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接HouYi rpc失败", err)
			return nil, err
		}
		srv.rpcClient = grpcConn
	}

	l.srvs.appendSrv(key, srv)
	log.Infow("服务注册成功", microServer.GetName(), "network", network.String(), "endpoint", microServer.GetEndpoint(), "timeout", microServer.GetTimeout())

	// 同步数据到 HouYi
	return srv, nil
}
