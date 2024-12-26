package server

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

func newAlertConsumer(c *palaceconf.Bootstrap, data *data.Data, alertService *service.AlertService) *watch.Watcher {
	opts := []watch.WatcherOption{
		watch.WithWatcherStorage(data.GetAlertConsumerStorage()),
		watch.WithWatcherQueue(data.GetAlertPersistenceDBQueue()),
		watch.WithWatcherTimeout(c.GetWatch().GetAlertEvent().GetTimeout().AsDuration()),
		watch.WithWatcherHandler(watch.NewDefaultHandler(
			watch.WithDefaultHandlerTopicHandle(vobj.TopicAlarm, func(ctx context.Context, msg *watch.Message) error {
				params := msg.GetData().(*bo.CreateAlarmHookRawParams)
				ctx = middleware.WithTeamIDContextKey(ctx, params.TeamID)
				return alertService.CreateAlarmInfo(ctx, params)
			}),
			watch.WithDefaultHandlerTopicHandle(vobj.TopicAlertMsg, func(ctx context.Context, msg *watch.Message) error {
				params, ok := msg.GetData().(*bo.SendMsg)
				if !ok {
					return nil
				}
				return alertService.SendAlertMsg(ctx, params.SendMsgRequest)
			}),
		)),
	}
	return watch.NewWatcher("palace service alertConsumerServer", opts...)
}
