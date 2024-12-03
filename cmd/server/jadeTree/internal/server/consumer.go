package server

import (
	"context"

	"github.com/aide-family/moon/cmd/server/jadeTree/internal/data"
	"github.com/aide-family/moon/cmd/server/jadeTree/internal/jadetreeconf"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

func newConsumer(_ *jadetreeconf.Bootstrap, data *data.Data) *watch.Watcher {
	opts := []watch.WatcherOption{
		watch.WithWatcherStorage(data.GetWatcherStorage()),
		watch.WithWatcherQueue(data.GetWatcherQueue()),
		watch.WithWatcherHandler(watch.NewDefaultHandler(
			watch.WithDefaultHandlerTopicHandle(vobj.TopicAlertMsg, func(ctx context.Context, msg *watch.Message) error {
				//msgParams, ok := msg.GetData().(*bo.SendMsgParams)
				//if !ok {
				//	return nil
				//}
				//time.Sleep(time.Second * 1)
				return nil
				//return sendService.Send(ctx, msgParams)
			}),
		)),
	}
	return watch.NewWatcher("notice worker", opts...)
}
