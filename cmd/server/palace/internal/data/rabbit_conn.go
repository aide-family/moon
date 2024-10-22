package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	pushapi "github.com/aide-family/moon/api/rabbit/push"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/plugin/microserver"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/go-kratos/kratos/v2/transport/http"
	"google.golang.org/grpc"
)

// NewRabbitRPCConn 创建一个rabbit rpc连接
func NewRabbitRPCConn(c *palaceconf.Bootstrap, data *Data) (*RabbitConn, error) {
	discoveryConf := c.GetDiscovery()
	return &RabbitConn{
		data:          data,
		srvs:          NewSrvList(c.GetDependRabbit()),
		discoveryConf: discoveryConf,
	}, nil
}

// RabbitConn rabbit服务连接
type RabbitConn struct {
	data *Data
	lock sync.Mutex
	// 服务实例原始信息
	srvs          *SrvList
	discoveryConf *conf.Discovery
}

// NotifyObject 发送通道配置
func (l *RabbitConn) NotifyObject(ctx context.Context, in *pushapi.NotifyObjectRequest, opts ...microserver.Option) error {
	eg := new(errgroup.Group)
	for _, srv := range l.srvs.getSrvs() {
		conn := srv
		eg.Go(func() error {
			return l.notifyObject(ctx, conn, in, opts...)
		})
	}
	return eg.Wait()
}

func (l *RabbitConn) notifyObject(ctx context.Context, srv *Srv, in *pushapi.NotifyObjectRequest, opts ...microserver.Option) error {
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
}

// SendMsg 发送消息
func (l *RabbitConn) SendMsg(ctx context.Context, in *hookapi.SendMsgRequest, opts ...microserver.Option) error {
	eg := new(errgroup.Group)
	for _, srv := range l.srvs.getSrvs() {
		conn := srv
		eg.Go(func() error {
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
		if err := l.sync(srv); !types.IsNil(err) {
			log.Errorw("method", "sync", "err", err)
		}
	}()
	return nil
}

func (l *RabbitConn) getTeamEmailConfig(teamId uint32) (*model.SysTeamEmail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysTeamEmail.Where(mainQuery.SysTeamEmail.TeamID.Eq(teamId)).First()
}

func (l *RabbitConn) sync(srv *Srv) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	wheres := []gen.Condition{
		mainQuery.SysTeam.Status.Eq(vobj.StatusEnable.GetValue()),
	}
	if len(srv.teamIds) > 0 {
		wheres = append(wheres, mainQuery.SysTeam.ID.In(srv.teamIds...))
	}

	teamDos, err := query.Use(l.data.GetMainDB(ctx)).SysTeam.Where(wheres...).Find()
	if !types.IsNil(err) {
		return err
	}
	for _, teamDo := range teamDos {
		emailConfigDo, _ := mainQuery.SysTeamEmail.Where(mainQuery.SysTeamEmail.TeamID.Eq(teamDo.ID)).First()
		// 获取所有的有效告警组
		teamDB, err := l.data.GetBizGormDB(teamDo.ID)
		if !types.IsNil(err) {
			return err
		}
		teamBizQuery := bizquery.Use(teamDB)
		noticeGroupItems, err := teamBizQuery.WithContext(ctx).AlarmNoticeGroup.
			Where(teamBizQuery.AlarmNoticeGroup.Status.Eq(vobj.StatusEnable.GetValue())).
			Preload(field.Associations).
			Find()
		if !types.IsNil(err) {
			log.Errorw("获取告警组失败", err)
			continue
		}
		var emailConfig *conf.EmailConfig
		if emailConfigDo != nil {
			emailConfig = &conf.EmailConfig{
				User: emailConfigDo.User,
				Pass: emailConfigDo.Pass,
				Host: emailConfigDo.Host,
				Port: emailConfigDo.Port,
			}
		}
		for _, noticeGroupItem := range noticeGroupItems {
			if err := l.syncNoticeGroup(srv, teamDo.ID, emailConfig, noticeGroupItem); !types.IsNil(err) {
				log.Errorw("同步告警组失败", err)
				continue
			}
		}
	}
	return nil
}

func (l *RabbitConn) syncNoticeGroup(srv *Srv, teamID uint32, teamEmailConfig *conf.EmailConfig, noticeGroupItem *bizmodel.AlarmNoticeGroup) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	member := types.SliceToWithFilter(noticeGroupItem.NoticeMembers, func(member *bizmodel.AlarmNoticeMember) (*bizmodel.SysTeamMember, bool) {
		m := member.GetMember()
		return m, m != nil || member.AlarmNoticeType.IsEmail() // TODO 后面兼容短信
	})
	members := builder.NewParamsBuild().WithContext(ctx).TeamMemberModuleBuilder().DoTeamMemberBuilder().ToAPIs(member)
	params := &pushapi.NotifyObjectRequest{
		Receivers: map[string]*conf.Receiver{
			fmt.Sprintf("team_%d_%d", teamID, noticeGroupItem.ID): {
				Hooks: types.SliceTo(noticeGroupItem.AlarmHooks, func(hook *bizmodel.AlarmHook) *conf.ReceiverHook {
					return &conf.ReceiverHook{
						Dingtalk: types.Ternary(hook.APP.IsDingtalk(), &conf.ReceiverHookDingTalk{
							Webhook:  hook.URL,
							Secret:   hook.Secret,
							Content:  "",
							Template: hook.APP.EnUSString(), // TODO 先固定模板， 后面再替换自定义模板
						}, nil),
						Feishu: types.Ternary(hook.APP.IsFeishu(), &conf.ReceiverHookFeiShu{
							Webhook:  hook.URL,
							Secret:   hook.Secret,
							Content:  "",
							Template: hook.APP.EnUSString(), // TODO 先固定模板， 后面再替换自定义模板
						}, nil),
						Wechat: types.Ternary(hook.APP.IsWechat(), &conf.ReceiverHookWechatWork{
							Webhook:  hook.URL,
							Content:  "",
							Template: hook.APP.EnUSString(), // TODO 先固定模板， 后面再替换自定义模板
						}, nil),
						Other: types.Ternary(hook.APP.IsWebhook(), &conf.ReceiverHookOther{
							Webhook:  hook.URL,
							Content:  "",
							Template: hook.APP.EnUSString(), // TODO 先固定模板， 后面再替换自定义模板
						}, nil),
					}
				}),
				Phones: nil,
				Emails: types.SliceToWithFilter(members, func(memberItem *adminapi.TeamMemberItem) (*conf.ReceiverEmail, bool) {
					user := memberItem.GetUser()
					if user == nil {
						return nil, false
					}
					return &conf.ReceiverEmail{
						To:          user.GetEmail(),
						Subject:     "Moon监控告警",
						Content:     "",
						Template:    "email", // 先固定模板
						Cc:          nil,
						AttachUrl:   nil,
						ContentType: "text/plain",
					}, true
				}),
				EmailConfig: teamEmailConfig,
			},
		},
		Templates: nil,
	}
	log.Infow("syncNoticeGroup", "开始推送通知对象")
	// 推送策略
	return l.notifyObject(ctx, srv, params)
}

// srvRegister 服务注册
func (l *RabbitConn) srvRegister(key string, microServer *conf.MicroServer, teamIds []uint32) (*Srv, error) {
	network := vobj.ToNetwork(microServer.GetNetwork())
	srv := &Srv{
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
			log.Errorw("连接Rabbit http失败：", err)
			return nil, err
		}
		srv.httpClient = httpConn
	default:
		grpcConn, err := microserver.NewRPCConn(microServer, l.discoveryConf)
		if !types.IsNil(err) {
			log.Errorw("连接Rabbit rpc失败：", err)
			return nil, err
		}
		srv.rpcClient = grpcConn
	}
	l.srvs.appendSrv(key, srv)
	log.Infow("服务注册成功", microServer.GetName(), "network", network.String(), "endpoint", microServer.GetEndpoint(), "timeout", microServer.GetTimeout())
	return srv, nil
}
