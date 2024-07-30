package server

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service/build"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

func newAlertWatch(c *houyiconf.Bootstrap, data *data.Data, alertService *service.AlertService) *watch.Watcher {
	opts := []watch.WatcherOption{
		watch.WithWatcherStorage(data.GetAlertStorage()),
		watch.WithWatcherQueue(data.GetAlertQueue()),
		watch.WithWatcherTimeout(c.GetWatch().GetAlertEvent().GetTimeout().AsDuration()),
		watch.WithWatcherHandler(watch.NewDefaultHandler(
			watch.WithDefaultHandlerTopicHandle(vobj.TopicAlert, func(ctx context.Context, msg *watch.Message) error {
				alarmData, ok := msg.GetData().(*bo.Alarm)
				if !ok || types.IsNil(alarmData) {
					return nil
				}
				alarmAPIData := build.NewAlarmBuilder(alarmData).ToAPI()
				_, err := alertService.Hook(ctx, alarmAPIData)
				return err
			}),
		)),
	}
	return watch.NewWatcher("alertWatchServer", opts...)
}
