package server

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/cmd/server/palace/internal/palaceconf"
	"github.com/aide-family/moon/cmd/server/palace/internal/service"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/watch"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"
)

const (
	// 策略任务执行间隔， 默认10s
	strategyWatchJobSpec = "@every 10s"
	// 测了执行任务超时时间
	strategyWatchJobTimeout = 10 * time.Second
)

func newStrategyWatch(c *palaceconf.Bootstrap, data *data.Data, alertService *service.AlertService) *StrategyWatch {
	cronInstance := cron.New()
	strategyConf := c.GetWatch().GetStrategy()
	interval := strategyWatchJobSpec
	if strategyConf.GetInterval() != "" {
		interval = strategyConf.GetInterval()
	}
	timeout := strategyWatchJobTimeout
	if strategyConf.GetTimeout().AsDuration() > 0 {
		timeout = strategyConf.GetTimeout().AsDuration()
	}
	return &StrategyWatch{
		data:         data,
		cronInstance: cronInstance,
		stopCh:       make(chan struct{}),
		entryIDMap:   make(map[string]cron.EntryID),
		alertService: alertService,
		interval:     interval,
		timeout:      timeout,
		// 没有配置后裔服务着启动内置简易的告警功能
		dependHouYi: c.GetDependHouYi(),
	}
}

var _ transport.Server = (*StrategyWatch)(nil)

// StrategyWatch 策略任务执行器
type StrategyWatch struct {
	data         *data.Data
	cronInstance *cron.Cron
	stopCh       chan struct{}
	entryIDMap   map[string]cron.EntryID
	interval     string
	timeout      time.Duration

	alertService *service.AlertService

	dependHouYi bool
}

// Start 启动策略任务执行器
func (s *StrategyWatch) Start(_ context.Context) error {
	if types.IsNil(s) || types.IsNil(s.cronInstance) {
		return merr.ErrorNotificationSystemError("strategy watch is nil")
	}
	if types.IsNil(s.data) {
		return merr.ErrorNotificationSystemError("data is nil")
	}
	if types.IsNil(s.data.GetStrategyQueue()) {
		return merr.ErrorNotificationSystemError("strategy queue is nil")
	}
	go func() {
		defer after.RecoverX()
		for {
			select {
			case <-s.stopCh:
				return
			case msg, ok := <-s.data.GetStrategyQueue().Next():
				if !ok || !msg.GetTopic().IsStrategy() {
					continue
				}

				if err := s.addJob(msg.GetData()); err != nil {
					log.Errorw("add job err", err)
				}
			}
		}
	}()
	s.cronInstance.Start()
	log.Infof("[StrategyWatch] server started")
	return nil
}

// Stop 停止策略任务执行器
func (s *StrategyWatch) Stop(_ context.Context) error {
	defer log.Infof("[StrategyWatch] server stopped")
	close(s.stopCh)
	s.cronInstance.Stop()
	return s.data.GetStrategyQueue().Close()
}

func (s *StrategyWatch) addJob(strategyMsg watch.Indexer) error {
	// 转换数据
	if s.dependHouYi {
		strategyBo, ok := strategyMsg.(*bo.Strategy)
		if !ok {
			return merr.ErrorNotificationSystemError("strategy msg type error")
		}
		// 推送到houyi服务去
		return s.alertService.PushStrategy(context.Background(), strategyBo)
	}

	log.Warnw("本地未实现告警功能，策略任务将不会执行")

	return nil
}
