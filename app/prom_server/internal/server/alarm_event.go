package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
	"prometheus-manager/api/prom/strategy/group"
	"prometheus-manager/app/prom_server/internal/service/promservice"

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
	groupService  *promservice.GroupService
}

func NewAlarmEvent(
	c *conf.Bootstrap,
	hookService *alarmservice.HookService,
	groupService *promservice.GroupService,
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
		groupService:  groupService,
	}

	// 注册topic处理器
	for _, topic := range kafkaConf.GetTopics() {
		switch topic {
		case string(consts.AlertHookTopic):
			l.eventHandlers[consts.AlertHookTopic] = l.alertHookHandler
		case string(consts.AgentOnlineTopic):
			l.eventHandlers[consts.AgentOnlineTopic] = l.agentOnlineEventHandler
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
	// TODO 后期是否判断key
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

// agentOnlineEventHandler 处理agent online消息
func (l *AlarmEvent) agentOnlineEventHandler(msg *kafka.Message) error {
	// TODO 1. 记录节点状态

	// 2. 拉取全量规则组及规则
	listAllGroupDetail, err := l.groupService.ListAllGroupDetail(context.Background(), &group.ListAllGroupDetailRequest{})
	if err != nil {
		return err
	}

	eg := new(errgroup.Group)
	eg.SetLimit(100)
	for _, groupDetail := range listAllGroupDetail.GetList() {
		if groupDetail == nil {
			continue
		}
		groupDetailBytes, _ := json.Marshal(groupDetail)
		eg.Go(func() error {
			// 3. 推送规则组消息(按规则组粒度)
			topic := string(msg.Value)
			sendMsg := &kafka.Message{
				TopicPartition: kafka.TopicPartition{
					Topic:     &topic,
					Partition: kafka.PartitionAny,
				},
				Value: groupDetailBytes,
				Key:   msg.Key,
			}
			return l.Produce(sendMsg)
		})
	}

	return eg.Wait()
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
				//l.log.Warnf("Error: %v\n", e)
			}
		}
	}()
	return nil
}
