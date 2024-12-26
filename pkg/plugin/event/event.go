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
)

// NewEvent 创建消息队列
func NewEvent(c *conf.Event) (mq.IMQ, error) {
	switch strings.ToLower(c.GetType()) {
	case rocketMQ:
		return NewRocketMQEvent(c.GetRocketMQ())
	case mockMQ:
		return mq.NewMockMQ(), nil
	case mqttMQ:
		return NewMqttEvent(c.GetMqtt())
	case kafkaMQ:
		return NewKafkaEvent(c.GetKafka())
	default:
		return nil, merr.ErrorNotificationSystemError("不支持的消息队列类型")
	}
}
