package msg

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
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

func (l *msgRepoImpl) SendAlarm(ctx context.Context, req ...*bo.AlarmMsgBo) error {
	for _, v := range req {
		if !l.cacheNotify(v.AlarmInfo) {
			continue
		}
		// 遍历告警组
		for _, v2 := range v.PromNotifies {
			// 通知到群组
			l.sendAlarmToChatGroups(ctx, v2.GetChatGroups(), v.AlarmInfo)
			// 通知到人员
			l.sendAlarmToMember(ctx, v2.GetBeNotifyMembers(), v.AlarmInfo)
		}
	}
	return nil
}

func (l *msgRepoImpl) cacheNotify(alarmInfo *bo.AlertBo) bool {
	fingerprint := hash.MD5(alarmInfo.Fingerprint + ":" + alarmInfo.Status)
	// 判断是否发送过告警， 如果已经发送过， 不再发送
	return l.d.Cache().SetNX(context.Background(), consts.AlarmNotifyCache.Key(fingerprint).String(), alarmInfo.Bytes(), 2*time.Hour)
}

func (l *msgRepoImpl) sendAlarmToChatGroups(ctx context.Context, chatGroups []*bo.ChatGroupBO, alarmInfo *bo.AlertBo) {
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
		if chatInfo.Template != "" {
			msg.Content = strategy.Formatter(chatInfo.Template, alarmInfoMap)
		}

		eg.Go(func() error {
			return NewHookNotify(chatInfo.NotifyApp).Alarm(ctx, chatInfo.Hook, msg)
		})
	}
	if err := eg.Wait(); err != nil {
		l.log.Warnf("send alarm to chat groups error, %v", err)
	}
}

// 通知到人员
func (l *msgRepoImpl) sendAlarmToMember(_ context.Context, members []*bo.NotifyMemberBO, alarmInfo *bo.AlertBo) {
	l.log.Debug("开始发送邮件通知")
	eg := new(errgroup.Group)
	eg.SetLimit(10)
	// 短信、邮件、电话
	for _, m := range members {
		if m.NotifyType.IsEmail() {
			// 发送邮件
			eg.Go(func() error {
				defer l.log.Debugw("发送邮件通知完成", m.GetMember().Email)
				return l.d.Email().SetBody(alarmInfo.String()).
					SetTo(m.GetMember().Email).
					SetSubject("prometheus moon系统邮件告警").Send()
			})
		}
		if m.NotifyType.IsSms() {
			// TODO 发送短信
		}
		if m.NotifyType.IsPhone() {
			// TODO 发送电话
		}
	}
	if err := eg.Wait(); err != nil {
		l.log.Warnf("send alarm to member error, %v", err)
	}
}

func NewMsgRepo(data *data.Data, logger log.Logger) repository.MsgRepo {
	return &msgRepoImpl{
		log: log.NewHelper(log.With(logger, "module", "repo.msg")),
		d:   data,
	}
}
