package server

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/robfig/cron/v3"
)

const (
	// 策略任务执行间隔， 默认10s
	strategyWatchJobSpec = "@every 10s"
	// 执行任务超时时间
	strategyWatchJobTimeout = 10 * time.Second
)

func newStrategyWatch(c *houyiconf.Bootstrap, data *data.Data, alertService *service.AlertService) *StrategyWatch {
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
		entryIDMap:   safety.NewMap[string, cron.EntryID](),
		strategyMap:  safety.NewMap[int, bo.IStrategy](),
		alertService: alertService,
		interval:     interval,
		timeout:      timeout,
	}
}

var _ transport.Server = (*StrategyWatch)(nil)

// StrategyWatch 策略任务执行器
type StrategyWatch struct {
	data         *data.Data
	cronInstance *cron.Cron
	stopCh       chan struct{}
	//entryIDMap   map[string]cron.EntryID
	entryIDMap *safety.Map[string, cron.EntryID]
	// 策略数据
	strategyMap *safety.Map[int, bo.IStrategy]
	interval    string
	timeout     time.Duration

	alertService *service.AlertService
}

// Start 启动策略任务执行器
func (s *StrategyWatch) Start(_ context.Context) error {
	if types.IsNil(s) || types.IsNil(s.cronInstance) {
		return merr.ErrorNotification("strategy watch is nil")
	}
	if types.IsNil(s.data) {
		return merr.ErrorNotification("data is nil")
	}
	if types.IsNil(s.data.GetStrategyQueue()) {
		return merr.ErrorNotification("strategy queue is nil")
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
				strategyMsg, ok := msg.GetData().(bo.IStrategy)
				if !ok {
					log.Warnf("strategy watch get data error: %v", msg.GetData())
					continue
				}
				if err := s.addStrategy(strategyMsg); err != nil {
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
	return nil
}

// 添加策略任务
func (s *StrategyWatch) addStrategy(strategyMsg bo.IStrategy) error {
	if strategyMsg.GetStatus().IsEnable() {
		return s.triggerEnableStrategy(strategyMsg)
	}
	return s.triggerDisableStrategy(strategyMsg)
}

// triggerDisableStrategy 触发策略关闭
func (s *StrategyWatch) triggerDisableStrategy(strategyMsg bo.IStrategy) error {
	log.Info("strategy watch remove job")
	// 移除任务
	id, exist := s.entryIDMap.Get(strategyMsg.Index())
	if exist {
		s.cronInstance.Remove(id)
	}
	s.entryIDMap.Delete(strategyMsg.Index())
	// 生成告警恢复事件（如果有告警发生过）
	return s.alertResolve(strategyMsg)
}

// triggerEnableStrategy 触发策略开启
func (s *StrategyWatch) triggerEnableStrategy(strategyMsg bo.IStrategy) error {
	// 如果任务已经存在，则更新策略数据
	if entryID, exist := s.entryIDMap.Get(strategyMsg.Index()); exist {
		s.strategyMap.Set(int(entryID), strategyMsg)
		return nil
	}

	// 不存在，则添加任务
	entryID, err := s.cronInstance.AddFunc(strategyWatchJobSpec, func() {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		jobID, exist := s.entryIDMap.Get(strategyMsg.Index())
		if !exist {
			log.Warnf("strategy watch job not exist: %s", strategyMsg.Index())
			return
		}
		strategyItem, exist := s.strategyMap.Get(int(jobID))
		if !exist {
			log.Warnf("strategy watch strategy not exist: %s", strategyMsg.Index())
			return
		}
		innerAlarm, err := s.alertService.InnerAlarm(ctx, strategyItem)
		if err != nil {
			log.Warnw("inner alarm err", err)
			return
		}

		if err := s.data.GetAlertQueue().Push(innerAlarm.Message()); err != nil {
			log.Warnw("push inner alarm err", err)
			return
		}
	})
	if err != nil {
		return err
	}
	s.entryIDMap.Set(strategyMsg.Index(), entryID)
	s.strategyMap.Set(int(entryID), strategyMsg)
	log.Infow("strategy watch add job", entryID)

	return nil
}

func (s *StrategyWatch) alertResolve(strategyMsg bo.IStrategy) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	innerAlarm, err := s.alertService.InnerAlarm(ctx, strategyMsg)
	if err != nil {
		log.Warnw("inner alarm err", err)
		return err
	}

	log.Debugw("method", "alertResolve", "innerAlarm", innerAlarm)
	if err := s.data.GetAlertQueue().Push(innerAlarm.Message()); err != nil {
		log.Warnw("push inner alarm err", err)
		return err
	}
	return nil
}
