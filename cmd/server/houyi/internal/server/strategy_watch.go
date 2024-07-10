package server

import (
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/pkg/watch"
)

func newStrategyWatch(c *houyiconf.Bootstrap, data *data.Data) *watch.Watcher {
	opts := []watch.WatcherOption{
		watch.WithWatcherStorage(data.GetStrategyStorage()),
		watch.WithWatcherQueue(data.GetStrategyQueue()),
		watch.WithWatcherTimeout(c.GetWatch().GetStrategy().GetTimeout().AsDuration()),
	}
	return watch.NewWatcher(opts...)
}
