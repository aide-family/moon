package msg

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/api/perrors"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vobj"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/strategy"
	"prometheus-manager/pkg/util/hash"
)

var _ repository.MsgRepo = (*msgRepoImpl)(nil)

type msgRepoImpl struct {
	repository.UnimplementedMsgRepo
	log *log.Helper
	d   *data.Data
}

func getHookAlarmTemplateMap(templates []*bo.NotifyTemplateBO) (map[vobj.NotifyApp]string, map[vobj.NotifyType]string) {
	hookTemplateMap := make(map[vobj.NotifyApp]string)
	memberTemplateMap := make(map[vobj.NotifyType]string)
	for _, temp := range templates {
		switch temp.NotifyType {
		case vobj.NotifyTemplateTypeCustom:
			hookTemplateMap[vobj.NotifyAppCustom] = temp.Content
		case vobj.NotifyTemplateTypeEmail:
			memberTemplateMap[vobj.NotifyTypeEmail] = temp.Content
		case vobj.NotifyTemplateTypeSms:
			memberTemplateMap[vobj.NotifyTypeSms] = temp.Content
		case vobj.NotifyTemplateTypeWeChatWork:
			hookTemplateMap[vobj.NotifyAppWeChatWork] = temp.Content
		case vobj.NotifyTemplateTypeDingDing:
			hookTemplateMap[vobj.NotifyAppDingDing] = temp.Content
		case vobj.NotifyTemplateTypeFeiShu:
			hookTemplateMap[vobj.NotifyAppFeiShu] = temp.Content
		default:
			hookTemplateMap[vobj.NotifyAppCustom] = temp.Content
		}
	}
	return hookTemplateMap, memberTemplateMap
}

func (l *msgRepoImpl) SendAlarm(ctx context.Context, req ...*bo.AlarmMsgBo) error {
	for _, v := range req {
		if !l.cacheNotify(v.AlarmInfo) {
			continue
		}

		hookTemplateMap, memberTemplateMap := getHookAlarmTemplateMap(v.Templates)
		// 遍历告警组
		for _, v2 := range v.PromNotifies {
			// 通知到群组
			l.sendAlarmToChatGroups(ctx, v2.GetChatGroups(), hookTemplateMap, v.AlarmInfo)
			// 通知到人员
			_ = l.sendAlarmToMember(ctx, v2.GetBeNotifyMembers(), memberTemplateMap, v.AlarmInfo)
		}
	}
	return nil
}

func (l *msgRepoImpl) cacheNotify(alarmInfo *bo.AlertBo) bool {
	fingerprint := hash.MD5(alarmInfo.Fingerprint + ":" + alarmInfo.Status)
	// 判断是否发送过告警， 如果已经发送过， 不再发送
	return l.d.Cache().SetNX(context.Background(), consts.AlarmNotifyCache.Key(fingerprint).String(), alarmInfo.Bytes(), 2*time.Hour)
}

func (l *msgRepoImpl) sendAlarmToChatGroups(ctx context.Context, chatGroups []*bo.ChatGroupBO, hookTemplateMap map[vobj.NotifyApp]string, alarmInfo *bo.AlertBo) {
	eg := new(errgroup.Group)
	eg.SetLimit(10)
	content := alarmInfo.String()
	notifiedHooks := make(map[string]struct{})
	for _, v := range chatGroups {
		if v == nil {
			continue
		}

		if _, ok := notifiedHooks[v.Hook]; ok {
			continue
		}
		notifiedHooks[v.Hook] = struct{}{}

		chatInfo := *v
		msg := &HookNotifyMsg{
			Content:   content,
			AlarmInfo: alarmInfo,
			Secret:    chatInfo.Secret,
		}
		alarmInfoMap := alarmInfo.ToMap()
		template := hookTemplateMap[chatInfo.NotifyApp]
		if template != "" {
			msg.Content = strategy.Formatter(template, alarmInfoMap)
		}

		eg.Go(func() error {
			return NewHookNotify(chatInfo.NotifyApp).Alarm(ctx, chatInfo.Hook, msg)
		})
	}
	if err := eg.Wait(); err != nil {
		l.log.Warnf("send alarm to chat groups error, %v", err)
	}
}

// SendAlarmToMember 通知到人员
func (l *msgRepoImpl) SendAlarmToMember(ctx context.Context, members []*bo.NotifyMemberBO, memberTemplateMap map[vobj.NotifyType]string, alarmInfo *bo.AlertBo) error {
	return l.sendAlarmToMember(ctx, members, memberTemplateMap, alarmInfo)
}

func (l *msgRepoImpl) sendAlarmToMember(_ context.Context, members []*bo.NotifyMemberBO, memberTemplateMap map[vobj.NotifyType]string, alarmInfo *bo.AlertBo) error {
	l.log.Debug("开始发送邮件通知")
	eg := new(errgroup.Group)
	eg.SetLimit(10)
	// 短信、邮件、电话
	for _, m := range members {
		if m.NotifyType.IsEmail() {
			template := memberTemplateMap[vobj.NotifyTypeEmail]
			if template != "" {
				template = strategy.Formatter(template, alarmInfo.ToMap())
			}
			// 发送邮件
			eg.Go(func() error {
				defer l.log.Debugw("发送邮件通知完成", m.GetMember().Email)
				emailInstance := l.d.Email()
				if emailInstance == nil {
					return perrors.ErrorNotFound("未配置邮件功能")
				}
				return l.d.Email().SetBody(template).
					SetTo(m.GetMember().Email).
					SetSubject("moon监控系统告警").Send()
			})
		}
		if m.NotifyType.IsSms() {
			// TODO 发送短信
			return perrors.ErrorNotFound("未配置短信功能")
		}
		if m.NotifyType.IsPhone() {
			// TODO 发送电话
			return perrors.ErrorNotFound("未配置电话功能")
		}
	}
	if err := eg.Wait(); err != nil {
		l.log.Warnf("send alarm to member error, %v", err)
		return err
	}
	return nil
}

func NewMsgRepo(data *data.Data, logger log.Logger) repository.MsgRepo {
	return &msgRepoImpl{
		log: log.NewHelper(log.With(logger, "module", "repo.msg")),
		d:   data,
	}
}
