package conn

import (
	"fmt"
	"math"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

type (
	// IConsumer kafka消费者接口
	IConsumer interface {
		Consume(callback IConsumerCallback)
	}

	// IProducer kafka生产者接口
	IProducer interface {
		Produce(message *kafka.Message) error
	}

	// KafkaConsumer kafka消费者
	KafkaConsumer struct {
		kafkaEndpoints []string
		topic          string
		*kafka.Consumer
		*log.Helper
	}

	// KafkaProducer kafka生产者
	KafkaProducer struct {
		kafkaEndpoints []string
		*kafka.Producer
		*log.Helper
	}

	// IConsumerCallback kafka消费者回调
	IConsumerCallback func(msg *kafka.Message) bool
)

var _ IProducer = (*KafkaProducer)(nil)
var _ IConsumer = (*KafkaConsumer)(nil)

// NewKafkaConsumer 创建kafka消费对象
func NewKafkaConsumer(kafkaEndpoints, topics []string, logger log.Logger) (*KafkaConsumer, error) {
	kafkaGroupID := "consumer-" + uuid.New().String()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":               strings.Join(kafkaEndpoints, ","),
		"group.id":                        kafkaGroupID,
		"auto.offset.reset":               "latest", // earliest, latest, none
		"session.timeout.ms":              6000,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.partition.eof":            true,
	})
	if err != nil {
		return nil, err
	}

	if err := c.SubscribeTopics(topics, nil); err != nil {
		return nil, err
	} //订阅主题，可以订阅多个

	return &KafkaConsumer{
		Consumer:       c,
		kafkaEndpoints: kafkaEndpoints,
		Helper:         log.NewHelper(log.With(logger, "module", "kafka-writer")),
	}, nil
}

// NewKafkaProducer 创建kafka生产对象
func NewKafkaProducer(endpoints []string, logger log.Logger) (*KafkaProducer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":             strings.Join(endpoints, ","),
		"api.version.request":           "true",
		"message.max.bytes":             1000000,
		"linger.ms":                     500,
		"sticky.partitioning.linger.ms": 1000,
		"retries":                       math.MaxInt32,
		"retry.backoff.ms":              1000,
		"acks":                          "1",
		"security.protocol":             "plaintext",
	})
	if err != nil {
		return nil, err
	}

	return &KafkaProducer{
		Producer:       p,
		kafkaEndpoints: endpoints,
		Helper:         log.NewHelper(log.With(logger, "module", "kafka-reader")),
	}, nil
}

// Consume 消费者
func (l *KafkaConsumer) Consume(callback IConsumerCallback) {
	go func() {
		for event := range l.Events() {
			switch et := event.(type) {
			case kafka.AssignedPartitions: // 重新分配分区
				l.Info("AssignedPartitions")
				if err := l.Assign(et.Partitions); err != nil {
					l.Errorf("%% Error assigning partitions: %v\n", err)
				}
			case kafka.RevokedPartitions: // 取消分配分区
				l.Info("RevokedPartitions")
				if err := l.Unassign(); err != nil {
					l.Errorf("%% Error unassigning partitions: %v\n", err)
				}
			case *kafka.Message:
				if !callback(et) {
					return
				}
			case kafka.PartitionEOF:
				fmt.Printf("Reached %v\n", et)
			case kafka.Error:
				l.Errorf("%% Error: %v\n", et)
				if et.Code() == kafka.ErrAllBrokersDown {
					l.Errorf("%% All brokers down: %v\n", et)
				}
				return
			}
		}
	}()
}

// Produce 生产者
func (l *KafkaProducer) Produce(message *kafka.Message) error {
	if err := l.Producer.Produce(message, nil); err != nil {
		return err
	}
	go l.Flush(15 * 1000)
	return nil
}
