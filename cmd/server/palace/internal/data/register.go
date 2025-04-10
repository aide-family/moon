package data

import (
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/google/wire"
	"google.golang.org/grpc"
)

// ProviderSetRPCConn wire set
var ProviderSetRPCConn = wire.NewSet(
	NewHouYiConn,
	NewRabbitRPCConn,
)

// NewSrvList 创建服务列表
func NewSrvList(depend bool) *SrvList {
	return &SrvList{
		srvs:   make(map[string]*Srv, 10),
		depend: depend,
	}
}

// SrvList 服务列表
type SrvList struct {
	lock   sync.Mutex
	srvs   map[string]*Srv
	depend bool
}

// Srv 服务
type Srv struct {
	// 服务实例信息
	srvInfo *conf.MicroServer
	// 处理的团队列表
	teamIds []uint32
	// rpc连接
	rpcClient *grpc.ClientConn
	// 网络请求类型
	network vobj.Network
	// http连接
	httpClient *http.Client
	// 服务注册时间
	registerTime time.Time
	// 服务开始时间
	firstRegisterTime time.Time
	// 服务启动时间 如果启动时间不一致，需要重新注册和推送
	uuid string
}

// SetUuid 设置uuid
func (l *Srv) SetUuid(uuid string) {
	l.uuid = uuid
}

// IsSameUuid 判断uuid是否一致
func (l *Srv) IsSameUuid(uuid string) bool {
	return l.uuid == uuid
}

// appendSrv 添加服务
func (l *SrvList) appendSrv(key string, srv *Srv) {
	if !l.depend {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	srv.firstRegisterTime = time.Now()
	oldSrv, ok := l.srvs[key]
	if !ok {
		l.srvs[key] = srv
		return
	}
	oldSrv.registerTime = srv.registerTime
	*srv = *oldSrv
}

// getSrv 获取服务
func (l *SrvList) getSrv(key string, isRegister ...bool) (*Srv, bool) {
	if !l.depend {
		return nil, false
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	srv, ok := l.srvs[key]
	if !ok {
		return nil, false
	}
	if len(isRegister) > 0 && isRegister[0] {
		srv.registerTime = time.Now()
		return srv, true
	}
	if err := srv.checkSrvIsAlive(); err != nil {
		return nil, false
	}
	return srv, ok
}

// removeSrv 删除服务
func (l *SrvList) removeSrv(key string) {
	if !l.depend {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	delete(l.srvs, key)
}

// getSrvs 获取服务列表
func (l *SrvList) getSrvs() []*Srv {
	if !l.depend {
		return []*Srv{}
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	srvs := make([]*Srv, 0, len(l.srvs))
	for _, srv := range l.srvs {
		if err := srv.checkSrvIsAlive(); err != nil {
			continue
		}
		srvs = append(srvs, srv)
	}
	return srvs
}

// close 关闭服务
func (l *Srv) close() {
	if l.rpcClient != nil {
		l.rpcClient.Close()
	}
	if l.httpClient != nil {
		l.httpClient.Close()
	}
}

// checkSrvIsAlive 检查服务是否存活
func (l *Srv) checkSrvIsAlive() (err error) {
	// 判断服务注册时间是否大于1分钟
	if time.Now().Before(l.registerTime.Add(1 * time.Minute)) {
		return nil
	}
	return merr.ErrorNotificationSystemError("%s 服务不可用", l.srvInfo.GetName())
}

// genSrvUniqueKey 生成服务唯一标识
func genSrvUniqueKey(srv *conf.MicroServer) string {
	return types.MD5(types.TextJoin(srv.GetName(), srv.GetEndpoint()))
}
