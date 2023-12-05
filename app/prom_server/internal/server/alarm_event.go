package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"

	alarmhookPb "prometheus-manager/api/alarm/hook"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/servers"

	"prometheus-manager/app/prom_server/internal/conf"
	"prometheus-manager/app/prom_server/internal/service/alarmservice"
)

type EventHandler func(msg *kafka.Message) error

type AlarmEvent struct {
	log *log.Helper
	c   *conf.Bootstrap
	*servers.KafkaMQServer
	eventHandlers map[consts.TopicType]EventHandler
	hookService   *alarmservice.HookService
}

func NewAlarmEvent(
	c *conf.Bootstrap,
	hookService *alarmservice.HookService,
	logger log.Logger,
) (*AlarmEvent, error) {
	kafkaConf := c.GetMq().GetKafka()
	kafkaMqServer, err := servers.NewKafkaMQServer(kafkaConf, logger)
	if err != nil {
		return nil, err
	}

	l := &AlarmEvent{
		log:           log.NewHelper(log.With(logger, "module", "server.alarm.event")),
		c:             c,
		eventHandlers: make(map[consts.TopicType]EventHandler),
		KafkaMQServer: kafkaMqServer,
		hookService:   hookService,
	}

	// 注册topic处理器
	for _, topic := range kafkaConf.GetTopics() {
		switch topic {
		case string(consts.AlertHookTopic):
			l.eventHandlers[consts.AlertHookTopic] = l.alertHookHandler
		default:
			return nil, fmt.Errorf("topic %s not support", topic)
		}
	}

	if err = l.Subscribe(kafkaConf.GetTopics()); err != nil {
		return nil, err
	}

	if err = l.Consume(); err != nil {
		return nil, err
	}

	return l, nil
}

// alertHook 处理alert hook数据
func (l *AlarmEvent) alertHookHandler(msg *kafka.Message) error {
	var req alarmhookPb.HookV1Request
	err := json.Unmarshal(msg.Value, &req)
	if err != nil {
		return err
	}
	if err = req.ValidateAll(); err != nil {
		return err
	}
	resp, err := l.hookService.V1(context.Background(), &req)
	if err != nil {
		return err
	}
	l.log.Debugf("hook resp: %s", resp.String())
	return nil
}

// Subscribe 订阅消息
func (l *AlarmEvent) Subscribe(topics []string) error {
	return l.KafkaMQServer.Consumer().SubscribeTopics(topics, func(consumer *kafka.Consumer, event kafka.Event) error {
		return nil
	})
}

// Consume 消费消息
func (l *AlarmEvent) Consume() error {
	consumer := l.KafkaMQServer.Consumer()
	go func() {
		events := consumer.Events()
		for event := range events {
			if consumer.IsClosed() {
				l.log.Warnf("consumer is closed")
				return
			}
			switch e := event.(type) {
			case *kafka.Message:
				// 处理消息, 根据不同的topic做不同的处理
				l.log.Debugf("Message on %s:\n%s\n", e.TopicPartition, string(e.Value))
				if e.TopicPartition.Topic == nil {
					break
				}
				topic := consts.TopicType(*e.TopicPartition.Topic)
				handler, ok := l.eventHandlers[topic]
				if !ok {
					l.log.Warnf("no handler found for topic: %s, event: %v", topic, e)
					break
				}
				if err := handler(e); err != nil {
					l.log.Errorf("handle event error: %v", err)
				}
			case kafka.Error:
				fmt.Printf("%% Error: %v\n", e)
			}
		}
	}()
	return nil
}
