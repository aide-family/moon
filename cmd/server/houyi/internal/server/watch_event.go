package server

import (
	"context"

	"github.com/aide-family/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/houyi/internal/data"
	"github.com/aide-family/moon/cmd/server/houyi/internal/houyiconf"
	"github.com/aide-family/moon/cmd/server/houyi/internal/service"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/plugin/event"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

func newEventStrategyWatch(c *houyiconf.Bootstrap, data *data.Data, alertService *service.AlertService) *eventStrategyWatch {
	return &eventStrategyWatch{
		stopCh:       make(chan struct{}),
		c:            c,
		data:         data,
		alertService: alertService,
		strategyMap:  safety.NewMap[string, bo.IStrategy](),
		mqMap:        safety.NewMap[string, mq.IMQ](),
	}
}

var _ transport.Server = (*eventStrategyWatch)(nil)

type (
	eventStrategyWatch struct {
		stopCh chan struct{}

		c            *houyiconf.Bootstrap
		data         *data.Data
		alertService *service.AlertService

		// 策略数据
		strategyMap *safety.Map[string, bo.IStrategy]
		mqMap       *safety.Map[string, mq.IMQ]
	}
)

func (m *eventStrategyWatch) registerEvent(c *bo.MQDatasource) (mq.IMQ, error) {
	if mqCli, ok := m.mqMap.Get(c.Index()); ok {
		if !c.Status.IsEnable() {
			mqCli.Close()
			m.mqMap.Delete(c.Index())
		}
		return mqCli, nil
	}
	mqCli, err := event.NewEvent(c.GetMQConfig())
	if err != nil {
		log.Errorf("[eventStrategyWatch] 创建 mq 失败: %v", err)
		return nil, err
	}
	m.mqMap.Set(c.Index(), mqCli)
	return mqCli, nil
}

func (m *eventStrategyWatch) Start(_ context.Context) error {
	go func() {
		defer after.RecoverX()
		for mqConf := range m.data.GetEventMQQueue().Next() {
			if !mqConf.GetTopic().IsMqdatasource() {
				log.Warnw("method", "eventStrategyWatch", "topic", mqConf.GetTopic().String())
				continue
			}
			c := mqConf.GetData().(*bo.MQDatasource)
			if _, err := m.registerEvent(c); err != nil {
				log.Errorf("[eventStrategyWatch] 创建 mq 失败: %v", err)
				continue
			}
		}
	}()

	go func() {
		defer after.RecoverX()
		for msg := range m.data.GetEventStrategyQueue().Next() {
			if !msg.GetTopic().IsEventstrategy() {
				log.Warnw("method", "eventStrategyWatch", "topic", msg.GetTopic().String())
				continue
			}
			strategyMsg, ok := msg.GetData().(bo.IStrategyEvent)
			if !ok {
				log.Warnf("strategy watch get data error: %v", msg.GetData())
				continue
			}
			var err error
			for _, datasource := range strategyMsg.GetDatasource() {
				mqCli, ok := m.mqMap.Get(datasource.Index())
				if !ok {
					mqCli, err = m.registerEvent(datasource)
					if err != nil {
						continue
					}
				}
				if !strategyMsg.GetStatus().IsEnable() {
					mqCli.RemoveReceiver(strategyMsg.GetTopic())
					continue
				}

				m.receive(mqCli, strategyMsg)
			}
		}
	}()

	log.Infof("[eventStrategyWatch] started")
	return nil
}

func (m *eventStrategyWatch) receive(mqCli mq.IMQ, strategyMsg bo.IStrategyEvent) {
	go func(cli mq.IMQ, strategy bo.IStrategyEvent) {
		defer after.RecoverX()
		for eventMsg := range cli.Receive(strategy.GetTopic()) {
			// 往 InnerAlarm 推送
			if _, err := m.alertService.InnerAlarm(context.Background(), strategyMsg.SetValue(eventMsg)); err != nil {
				log.Errorw("method", "eventStrategyWatch.receive", "err", err)
				continue
			}
		}
	}(mqCli, strategyMsg)
}

func (m *eventStrategyWatch) Stop(ctx context.Context) error {
	log.Infof("[eventStrategyWatch] stopped")
	return nil
}
