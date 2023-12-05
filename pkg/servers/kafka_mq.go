package servers

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"

	"prometheus-manager/pkg/conn"
)

var _ transport.Server = (*KafkaMQServer)(nil)

type KafkaMQServer struct {
	log      *log.Helper
	consumer *kafka.Consumer
	producer *kafka.Producer
}

type KafkaMQServerConfig interface {
	GetEndpoints() []string
	GetTopics() []string
	GetGroupId() string
}

func (l *KafkaMQServer) Start(ctx context.Context) (err error) {
	l.log.WithContext(ctx).Info("[KafkaMQServer] server starting")
	for {
		select {
		case <-ctx.Done():
			l.close(ctx)
			return
		}
	}
}

func (l *KafkaMQServer) Stop(ctx context.Context) error {
	l.log.WithContext(ctx).Info("[KafkaMQServer] server stopping")
	return nil
}

func (l *KafkaMQServer) close(ctx context.Context) {
	_ = l.consumer.Close()
	l.producer.Close()
	l.log.WithContext(ctx).Info("[KafkaMQServer] server stopped")
}

func NewKafkaMQServer(c KafkaMQServerConfig, logger log.Logger) (*KafkaMQServer, error) {
	kafkaEndpoints := c.GetEndpoints()
	kafkaGroupID := c.GetGroupId()
	consumer, err := conn.NewKafkaConsumer(kafkaEndpoints, kafkaGroupID)
	if err != nil {
		return nil, err
	}

	producer, err := conn.NewKafkaProducer(kafkaEndpoints)
	if err != nil {
		return nil, err
	}

	return &KafkaMQServer{
		log:      log.NewHelper(log.With(logger, "module", "servers.kafka_mq")),
		consumer: consumer,
		producer: producer,
	}, nil
}

// Consume 消费多个topic kafka消息
func (l *KafkaMQServer) Consume(topics []string, callback func(msg *kafka.Message) bool) error {
	return l.consumer.SubscribeTopics(topics, func(consumer *kafka.Consumer, event kafka.Event) error {
		switch e := event.(type) {
		case kafka.AssignedPartitions:
			// kafka.AssignedPartitions 指定分区
			l.log.Info("AssignedPartitions: %v", e)
			return nil
		case kafka.RevokedPartitions:
			// kafka.RevokedPartitions 撤销分区
			l.log.Info("RevokedPartitions: %v", e)
			return nil
		case *kafka.Message:
			// kafka.Message 消息
			l.log.Info("Message on %s: %s\n", e.TopicPartition, string(e.Value))
			if callback(e) {
				// 确认消息已经收到并处理
				_, err := consumer.CommitMessage(e)
				return err
			}
			return nil
		default:
			return nil
		}
	})
}

// Produce kafka消息
func (l *KafkaMQServer) Produce(msg *kafka.Message) error {
	return l.producer.Produce(msg, nil)
}

// Consumer kafka消费者
func (l *KafkaMQServer) Consumer() *kafka.Consumer {
	return l.consumer
}

// Producer kafka生产者
func (l *KafkaMQServer) Producer() *kafka.Producer {
	return l.producer
}
