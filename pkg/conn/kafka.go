package conn

import (
	"math"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// NewKafkaConsumer 创建kafka消费对象
func NewKafkaConsumer(kafkaEndpoints []string, kafkaGroupID string) (*kafka.Consumer, error) {
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

	return c, nil
}

// NewKafkaProducer 创建kafka生产对象
func NewKafkaProducer(endpoints []string) (*kafka.Producer, error) {
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

	return p, nil
}
