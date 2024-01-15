package server

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/service"
	"prometheus-manager/pkg/after"
)

var _ transport.Server = (*Watch)(nil)

type Watch struct {
	conf   *conf.WatchProm
	log    *log.Helper
	exitCh chan struct{}
	ticker *time.Ticker

	loadService *service.LoadService

	groups *sync.Map
}

func NewWatch(
	c *conf.WatchProm,
	loadService *service.LoadService,
	logger log.Logger,
) *Watch {
	return &Watch{
		conf:        c,
		exitCh:      make(chan struct{}, 1),
		ticker:      time.NewTicker(c.GetInterval().AsDuration()),
		log:         log.NewHelper(log.With(logger, "module", "server.watch")),
		loadService: loadService,
		groups:      new(sync.Map),
	}
}

func (w *Watch) Start(ctx context.Context) error {
	groupAll, err := w.loadService.StrategyGroupAll(ctx, &agent.StrategyGroupAllRequest{})
	if err != nil {
		w.log.Errorf("[Watch] load groups error: %v", err)
		return err
	}
	for _, group := range groupAll.GroupList {
		w.groups.Store(group.GetId(), group)
	}
	go func() {
		defer after.Recover(w.log)
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				eg := new(errgroup.Group)
				eg.SetLimit(100)
				w.groups.Range(func(key, value any) bool {
					groupDetail, ok := value.(*agent.GroupSimple)
					if !ok {
						return true
					}
					eg.Go(func() error {
						_, _ = w.loadService.Evaluate(context.Background(), &agent.EvaluateRequest{GroupList: []*agent.GroupSimple{groupDetail}})
						return nil
					})
					return true
				})
				_ = eg.Wait()
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return nil
}

func (w *Watch) Stop(_ context.Context) error {
	w.log.Info("[Watch] server stopping")
	close(w.exitCh)
	return nil
}

func (w *Watch) shutdown() {
	w.groups = nil
	w.ticker.Stop()
	w.log.Info("[Watch] server stopped")
}
