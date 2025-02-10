package data

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	pushapi "github.com/aide-family/moon/api/rabbit/push"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
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
	// 服务实例原始信息
	srvs          *SrvList
	discoveryConf *conf.Discovery
}

// GetServerList 获取rabbit服务列表
func (l *RabbitConn) GetServerList() (*api.GetServerListReply, error) {
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

// NotifyObject 发送通道配置
func (l *RabbitConn) NotifyObject(ctx context.Context, in *pushapi.NotifyObjectRequest, opts ...microserver.Option) error {
	eg := new(errgroup.Group)
	for _, srv := range l.srvs.getSrvs() {
		conn := srv
		eg.Go(func() error {
			return l.notifyObject(types.CopyValueCtx(ctx), conn, in, opts...)
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
	srvs := l.srvs.getSrvs()
	if len(srvs) == 0 {
		// 把消息存入队列， 防止消息丢失
		log.Infow("method", "SendMsg", "消息存入队列", in)
		return merr.ErrorNotificationSystemError("没有可用的rabbit服务")
	}
	for _, srv := range l.srvs.getSrvs() {
		conn := srv
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
	}
	return nil
}

// Heartbeat 心跳
func (l *RabbitConn) Heartbeat(_ context.Context, req *api.HeartbeatRequest) error {
	// 存储心跳数据
	srvKey := genSrvUniqueKey(req.GetServer())
	if !req.GetOnline() {
		l.srvs.removeSrv(srvKey)
		return nil
	}
	srv, ok := l.srvs.getSrv(srvKey, true)
	if ok && srv.IsSameUuid(req.GetUuid()) {
		return nil
	}
	srv, err := l.srvRegister(srvKey, req.GetServer(), req.GetTeamIds())
	if !types.IsNil(err) {
		log.Errorw("method", "srvRegister", "err", err)
		return err
	}
	srv.SetUuid(req.GetUuid())
	go func() {
		defer after.RecoverX()
		if err := l.sync(srv); !types.IsNil(err) {
			log.Errorw("method", "sync", "err", err)
		}
	}()
	return nil
}

// getTeamEmailConfig 获取团队邮箱配置
func (l *RabbitConn) getTeamEmailConfig(teamID uint32) (*model.SysTeamConfig, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	return mainQuery.WithContext(ctx).SysTeamConfig.Where(mainQuery.SysTeamConfig.TeamID.Eq(teamID)).First()
}

// SyncTeam 同步团队
func (l *RabbitConn) SyncTeam(ctx context.Context, teamID uint32, srvs ...*Srv) error {
	if teamID <= 0 {
		return nil
	}
	if len(srvs) == 0 {
		srvs = l.srvs.getSrvs()
	}
	newSrvs := make([]*Srv, 0, len(srvs))
	// 根据teamID 过滤srvs
	for _, srv := range srvs {
		if types.ContainsOf(srv.teamIds, func(teamId uint32) bool { return teamId == teamID }) {
			newSrvs = append(newSrvs, srv)
		}
	}
	if len(newSrvs) == 0 {
		return nil
	}
	mainQuery := query.Use(l.data.GetMainDB(ctx))
	teamConfigDo, _ := mainQuery.SysTeamConfig.Where(mainQuery.SysTeamConfig.TeamID.Eq(teamID)).First()
	// 获取所有的有效告警组
	teamDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}

	teamBizQuery := bizquery.Use(teamDB)
	noticeGroupItems, err := teamBizQuery.WithContext(ctx).AlarmNoticeGroup.
		Where(teamBizQuery.AlarmNoticeGroup.Status.Eq(vobj.StatusEnable.GetValue())).
		Preload(field.Associations).
		Preload(teamBizQuery.AlarmNoticeGroup.NoticeMembers.Member).
		Preload(teamBizQuery.AlarmNoticeGroup.Templates).
		Find()
	if !types.IsNil(err) {
		log.Errorw("获取告警组失败", err)
		return err
	}
	// 获取所有告警模板
	templates, err := teamBizQuery.WithContext(ctx).SysSendTemplate.Find()
	if !types.IsNil(err) {
		log.Errorw("获取告警模板失败", err)
		return err
	}

	templateMap := make(map[string]string)
	// 获取公告告警模板
	noticeTemplates, err := teamBizQuery.WithContext(ctx).SysSendTemplate.Find()
	if !types.IsNil(err) {
		log.Errorw("获取公告告警模板失败", err)
		return err
	}
	for _, template := range noticeTemplates {
		templateMap[template.GetSendType().EnUSString()] = template.Content
	}
	for _, template := range templates {
		key := fmt.Sprintf("team_%d_%d_%s", teamID, template.ID, template.GetSendType().EnUSString())
		templateMap[key] = template.Content
	}

	emailConfig := teamConfigDo.GetEmailConfig().ToConf()
	syncNoticeGroupMap := make(map[string]*conf.Receiver)
	syncNoticeGroups := types.SliceTo(noticeGroupItems, func(noticeGroupItem *bizmodel.AlarmNoticeGroup) map[string]*conf.Receiver {
		return l.buildSyncNoticeGroup(teamID, emailConfig, noticeGroupItem)
	})
	for _, receiver := range syncNoticeGroups {
		for key, value := range receiver {
			syncNoticeGroupMap[key] = value
		}
	}

	params := &pushapi.NotifyObjectRequest{
		Receivers: syncNoticeGroupMap,
		Templates: templateMap,
	}
	if err := l.syncNotifyObject(ctx, newSrvs, params); !types.IsNil(err) {
		log.Errorw("同步告警模板失败", err)
		return err
	}

	return nil
}

// syncTemplate 同步告警模板
func (l *RabbitConn) syncNotifyObject(ctx context.Context, srvs []*Srv, params *pushapi.NotifyObjectRequest) error {
	for _, srv := range srvs {
		if err := l.notifyObject(types.CopyValueCtx(ctx), srv, params); !types.IsNil(err) {
			log.Errorw("同步告警模板失败", err, "params", params)
			continue
		}
	}
	return nil
}

// sync 同步团队通知对象
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

	teamDos, err := mainQuery.SysTeam.Where(wheres...).Find()
	if !types.IsNil(err) {
		return err
	}
	for _, teamDo := range teamDos {
		if err = l.SyncTeam(ctx, teamDo.ID, srv); !types.IsNil(err) {
			log.Errorw("method", "同步团队告警对象失败", "err", err)
		}
	}
	return nil
}

// getTemplate 获取模板
func getTemplate(teamID uint32, templateMap map[string]*bizmodel.SysSendTemplate, sendType string) string {
	if template, ok := templateMap[sendType]; ok {
		return fmt.Sprintf("team_%d_%d_%s", teamID, template.ID, template.GetSendType().EnUSString())
	}
	return sendType
}

// buildSyncNoticeGroup 构建同步告警组数据
func (l *RabbitConn) buildSyncNoticeGroup(teamID uint32, teamEmailConfig *conf.EmailConfig, noticeGroupItem *bizmodel.AlarmNoticeGroup) map[string]*conf.Receiver {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	member := types.SliceToWithFilter(noticeGroupItem.NoticeMembers, func(member *bizmodel.AlarmNoticeMember) (*bizmodel.SysTeamMember, bool) {
		m := member.GetMember()
		return m, m != nil || member.AlarmNoticeType.IsEmail() // TODO 后面兼容短信
	})
	templateMap := types.ToMap(noticeGroupItem.Templates, func(template *bizmodel.SysSendTemplate) string {
		return template.GetSendType().EnUSString()
	})
	members := builder.NewParamsBuild(ctx).TeamMemberModuleBuilder().DoTeamMemberBuilder().ToAPIs(member)
	return map[string]*conf.Receiver{
		fmt.Sprintf("team_%d_%d", teamID, noticeGroupItem.ID): {
			Hooks: types.SliceTo(noticeGroupItem.AlarmHooks, func(hook *bizmodel.AlarmHook) *conf.ReceiverHook {
				return &conf.ReceiverHook{
					Type:     hook.APP.EnUSString(),
					Webhook:  hook.URL,
					Template: getTemplate(teamID, templateMap, hook.APP.EnUSString()),
					Secret:   hook.Secret,
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
					Template:    getTemplate(teamID, templateMap, "email"),
					Cc:          nil,
					AttachUrl:   nil,
					ContentType: "text/plain",
				}, true
			}),
			EmailConfig: teamEmailConfig,
		},
	}
}

// srvRegister 服务注册
func (l *RabbitConn) srvRegister(key string, microServer *conf.MicroServer, teamIds []uint32) (*Srv, error) {
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
