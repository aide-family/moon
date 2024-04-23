package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/pkg/after"
	"github.com/aide-family/moon/pkg/helper/consts"
	"github.com/aide-family/moon/pkg/servers"
	"github.com/aide-family/moon/pkg/util/interflow"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
)

var _ interflow.Interflow = (*kafkaInterflow)(nil)

type (
	kafkaInterflow struct {
		kafkaMQServer *servers.KafkaMQServer
		log           *log.Helper
		handles       map[consts.TopicType]interflow.Callback
		lock          sync.RWMutex

		agentTopic string
	}
)

func (l *kafkaInterflow) OnlineNotify() error {
	topic := string(consts.AgentOnlineTopic)

	msg := &interflow.HookMsg{
		Topic: topic,
		Value: []byte(l.agentTopic),
	}

	go func() {
		defer after.Recover(l.log)
		for {
			ctx, cancel := context.WithTimeout(context.Background(), interflow.Timeout)
			err := l.Send(ctx, msg)
			cancel()
			if err == nil {
				break
			}
			l.log.Warnw("send online notify error", err)
			time.Sleep(10 * time.Second)
		}
	}()
	return nil
}

func (l *kafkaInterflow) OfflineNotify() error {
	topic := string(consts.AgentOfflineTopic)
	l.log.Infow("topic", topic)
	msg := &interflow.HookMsg{
		Topic: topic,
		Value: []byte(l.agentTopic),
	}
	count := 1
	for {
		if count > 3 {
			break
		}
		ctx, cancel := context.WithTimeout(context.Background(), interflow.Timeout)
		if err := l.Send(ctx, msg); err != nil {
			cancel()
			l.log.Warnw("send offline notify error", err)
			count++
			// 等待1秒
			time.Sleep(1 * time.Second)
			continue
		}
		cancel()
		break
	}

	return nil
}

func (l *kafkaInterflow) Close() error {
	return nil
}

func (l *kafkaInterflow) SetHandles(handles map[consts.TopicType]interflow.Callback) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.handles = handles
	topics := make([]string, 0, len(l.handles))
	for topic := range l.handles {
		topics = append(topics, string(topic))
	}

	return l.subscribe(topics)
}

func (l *kafkaInterflow) Receive() error {
	consumer := l.kafkaMQServer.Consumer()
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
				l.log.Debugf("Message on %s\n", e.TopicPartition)
				if e.TopicPartition.Topic == nil {
					break
				}
				topic := consts.TopicType(*e.TopicPartition.Topic)
				handler, ok := l.handles[topic]
				if !ok {
					l.log.Errorf("handle event not-fund")
					continue
				}
				if err := handler(topic, e.Value); err != nil {
					l.log.Errorf("handle event error: %v", err)
				}
				// 确认消息
				if _, err := consumer.CommitMessage(e); err != nil {
					l.log.Errorf("commit message error: %v", err)
				}
			case kafka.Error:
				l.log.Warnf("Receive Error: %v\n", e)
			}
		}
	}()
	return nil
}

func (l *kafkaInterflow) Send(_ context.Context, msg *interflow.HookMsg) error {
	sendMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Value: msg.Value,
	}
	return l.kafkaMQServer.Produce(sendMsg)
}

// Subscribe 订阅消息
func (l *kafkaInterflow) subscribe(topics []string) error {
	return l.kafkaMQServer.Consume(topics, l.handleMessage)
}

func (l *kafkaInterflow) handleMessage(msg *kafka.Message) bool {
	topic := consts.TopicType(*msg.TopicPartition.Topic)
	l.log.Infow("topic", topic)
	if handler, ok := l.handles[topic]; ok {
		if err := handler(topic, msg.Value); err != nil {
			l.log.Errorf("handle message error: %s", err.Error())
		}
	}
	return true
}

func NewKafkaInterflow(kafkaMQServer *servers.KafkaMQServer, log *log.Helper) (interflow.Interflow, error) {
	instance := &kafkaInterflow{
		kafkaMQServer: kafkaMQServer,
		log:           log,
	}

	return instance, nil
}
