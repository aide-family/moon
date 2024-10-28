package server

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/data"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/service"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

func newConsumer(_ *rabbitconf.Bootstrap, data *data.Data, sendService *service.HookService) *watch.Watcher {
	opts := []watch.WatcherOption{
		watch.WithWatcherStorage(data.GetWatcherStorage()),
		watch.WithWatcherQueue(data.GetWatcherQueue()),
		watch.WithWatcherHandler(watch.NewDefaultHandler(
			watch.WithDefaultHandlerTopicHandle(vobj.TopicAlertMsg, func(ctx context.Context, msg *watch.Message) error {
				msgParams, ok := msg.GetData().(*bo.SendMsgParams)
				if !ok {
					return nil
				}
				time.Sleep(time.Second * 1)
				return sendService.Send(ctx, msgParams)
			}),
		)),
	}
	return watch.NewWatcher("notice worker", opts...)
}
