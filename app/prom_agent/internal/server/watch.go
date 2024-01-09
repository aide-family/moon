package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/pkg/strategy"
)

var _ transport.Server = (*Watch)(nil)

type Watch struct {
	conf   *conf.WatchProm
	log    *log.Helper
	exitCh chan struct{}
	ticker *time.Ticker

	alerting strategy.Alerter
}

const defaultDatasource = "https://prom-server.aide-cloud.cn"

func (w *Watch) Start(ctx context.Context) error {
	rule := &strategy.Rule{
		Alert: "test-alert",
		Expr:  "up == 1",
		For:   "3s",
		Labels: map[string]string{
			"job": "test-job",
		},
		Annotations: map[string]string{
			"summary":     "test-summary",
			"description": "test-description",
		},
	}
	rule.SetEndpoint(defaultDatasource)

	group := &strategy.Group{
		Name:  "test-group",
		Rules: []*strategy.Rule{rule},
	}
	go func() {
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				w.log.Info("[Watch] server tick")
				results, err := w.alerting.Eval(context.Background(), group, rule)
				if err != nil {
					w.log.Warnf("[Watch] server tick error: %v", err)
					continue
				}
				if len(results) > 0 {
					// 更新告警缓存
					w.log.Info("[Watch] server tick results:")
					for _, result := range results {
						w.log.Infof("[Watch] server tick result: %v", result.GetMetric().String())
					}
				} else {
					// 消除告警
					w.log.Infof("[Watch] server tick results: %v", results)
				}
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return nil
}

func (w *Watch) Stop(ctx context.Context) error {
	w.log.Info("[Watch] server stopping")
	close(w.exitCh)

	return nil
}

func (w *Watch) shutdown() {
	w.log.Info("[Watch] server stopped")
}

func NewWatch(c *conf.WatchProm, logger log.Logger) *Watch {
	return &Watch{
		conf:     c,
		exitCh:   make(chan struct{}, 1),
		ticker:   time.NewTicker(c.GetInterval().AsDuration()),
		log:      log.NewHelper(log.With(logger, "module", "server.watch")),
		alerting: strategy.NewAlerting(logger),
	}
}
