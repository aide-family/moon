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

func (l *msgRepoImpl) SendAlarm(ctx context.Context, hookBytes []byte, req ...*bo.AlarmMsgBo) error {
	for _, v := range req {
		if !l.cacheNotify(v.AlarmInfo) {
			continue
		}
		// 遍历告警组
		for _, v2 := range v.PromNotifies {
			// 通知到群组
			l.sendAlarmToChatGroups(hookBytes, v2.ChatGroups, v.AlarmInfo)
		}
	}
	return nil
}

func (l *msgRepoImpl) cacheNotify(alarmInfo *bo.AlertBo) bool {
	fingerprint := hash.MD5(alarmInfo.Fingerprint + ":" + alarmInfo.Status)
	// 判断十分发送过告警， 如果已经发送过， 不再发送
	return l.d.Client().SetNX(context.Background(), consts.AlarmNotifyCache.Key(fingerprint).String(), alarmInfo.String(), 2*time.Hour).Val()
}

func (l *msgRepoImpl) sendAlarmToChatGroups(hookBytes []byte, chatGroups []*bo.ChatGroupBO, alarmInfo *bo.AlertBo) {
	eg := new(errgroup.Group)
	eg.SetLimit(10)
	content := alarmInfo.String()
	for _, v := range chatGroups {
		if v == nil {
			continue
		}
		msg := &HookNotifyMsg{
			Context:   content,
			Title:     "",
			AlarmInfo: alarmInfo,
			HookBytes: hookBytes,
		}
		chatInfo := *v
		if chatInfo.Template != "" {
			msg.Context = strategy.Formatter(chatInfo.Template, alarmInfo.ToMap())
		}
		if chatInfo.Title != "" {
			msg.Title = strategy.Formatter(chatInfo.Title, alarmInfo.ToMap())
		}
		eg.Go(func() error {
			return NewHookNotify(v.NotifyApp).Alarm(v.Hook, msg)
		})
	}
	if err := eg.Wait(); err != nil {
		l.log.Errorf("send alarm to chat groups error, %v", err)
	}
}

func NewMsgRepo(data *data.Data, logger log.Logger) repository.MsgRepo {
	return &msgRepoImpl{
		log: log.NewHelper(log.With(logger, "module", "repo.msg")),
		d:   data,
	}
}
