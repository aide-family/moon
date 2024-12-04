package event

import (
	"strings"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	mockMQ   = "mock"
	rocketMQ = "rocketmq"
	kafkaMQ  = "kafka"
	mqttMQ   = "mqtt"
	rabbitmq = "rabbitmq"
)

// NewEvent 创建消息队列
func NewEvent(c *conf.MQ) (mq.IMQ, error) {
	switch strings.ToLower(c.GetType()) {
	case rocketMQ:
		return NewRocketMQEvent(c.GetRocketMQ())
	case mockMQ:
		return mq.NewMockMQ(), nil
	case mqttMQ:
		return NewMqttEvent(c.GetMqtt())
	case kafkaMQ:
		return NewKafkaEvent(c.GetKafka())
	case rabbitmq:
		return NewRabbitMQEvent(c.GetRabbitMQ())
	default:
		return nil, merr.ErrorNotificationSystemError("不支持的消息队列类型")
	}
}
