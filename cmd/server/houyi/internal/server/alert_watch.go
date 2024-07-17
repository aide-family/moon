package server

import (
	"context"

	alertapi "github.com/aide-family/moon/api/houyi/alert"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
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
				// TODO implement alert hook
				_, err := alertService.Hook(ctx, &alertapi.HookRequest{})
				return err
			}),
		)),
	}
	return watch.NewWatcher("alertWatchServer", opts...)
}
