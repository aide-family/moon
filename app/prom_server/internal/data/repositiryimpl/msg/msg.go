package msg

import (
	"context"
	"strconv"
	"time"

	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/app/prom_server/internal/data"
	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/strategy"
	"github.com/aide-family/moon/pkg/util/hash"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
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
		if v.StrategyBO == nil {
			continue
		}
		if !l.cacheNotify(v.AlarmInfo, v.StrategyBO.SendInterval) {
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

func (l *msgRepoImpl) cacheNotify(alarmInfo *bo.AlertBo, sendInterval string) bool {
	fingerprint := hash.MD5(alarmInfo.Fingerprint + ":" + alarmInfo.Status)
	// 判断是否发送过告警， 如果已经发送过， 不再发送
	return l.d.Cache().SetNX(context.Background(), consts.AlarmNotifyCache.Key(fingerprint).String(), alarmInfo.Bytes(), ConvertTimeFromStringToDuration(sendInterval))
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
		template := hookTemplateMap[chatInfo.NotifyApp]
		if template != "" {
			msg.Content = strategy.Formatter(template, alarmInfo)
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
				template = strategy.Formatter(template, alarmInfo)
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

func ConvertTimeFromStringToDuration(duration string) time.Duration {
	return time.Duration(BuildDuration(duration)) * time.Second
}

// BuildDuration 字符串转为api时间
func BuildDuration(duration string) int64 {
	durationLen := len(duration)
	if duration == "" || durationLen < 2 {
		return 0
	}
	value, _ := strconv.Atoi(duration[:durationLen-1])
	// 获取字符串最后一个字符
	unit := string(duration[durationLen-1])
	switch unit {
	case "s":
		return int64(value)
	case "m":
		return int64(value) * 60
	case "h":
		return int64(value) * 60 * 60
	case "d":
		return int64(value) * 60 * 60 * 24
	default:
		return 0
	}
}
