package event

import (
	"context"
	"strings"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/IBM/sarama"
	"github.com/go-kratos/kratos/v2/log"
)

var _ mq.IMQ = (*KafkaEvent)(nil)

type (
	// KafkaEvent kafka event
	KafkaEvent struct {
		c              *conf.Kafka
		consumerGroup  sarama.ConsumerGroup
		topicChanelMap *safety.Map[string, chan *mq.Msg]
	}

	// consumerGroupHandler consumer group handler
	consumerGroupHandler struct {
		topic       string
		messageChan chan<- *mq.Msg
	}
)

// NewKafkaEvent 创建kafka event
func NewKafkaEvent(c *conf.Kafka) (*KafkaEvent, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotificationSystemError("KafkaEvent 配置为空")
	}
	k := &KafkaEvent{c: c, topicChanelMap: safety.NewMap[string, chan *mq.Msg]()}
	if err := k.init(); err != nil {
		return nil, err
	}
	return k, nil
}

// Send 发送
func (k *KafkaEvent) Send(_ string, _ []byte) error {
	return nil
}

// Receive 接收
func (k *KafkaEvent) Receive(topic string) <-chan *mq.Msg {
	ch, ok := k.topicChanelMap.Get(topic)
	if ok {
		return ch
	}

	ch = make(chan *mq.Msg, 1000)
	k.topicChanelMap.Set(topic, ch)
	go func() {
		handler := &consumerGroupHandler{
			topic:       topic,
			messageChan: ch,
		}
		for {
			err := k.consumerGroup.Consume(context.Background(), []string{topic}, handler)
			if err != nil {
				log.Infof("error consuming topic %s: %v\n", topic, err)
				continue
			}
		}
	}()

	return ch
}

// RemoveReceiver 移除监听
func (k *KafkaEvent) RemoveReceiver(topic string) {
	defer func() {
		k.topicChanelMap.Delete(topic)
		ch, ok := k.topicChanelMap.Get(topic)
		if ok {
			close(ch)
		}
	}()
}

// Close 关闭
func (k *KafkaEvent) Close() {
	// 关闭通道
	for _, ch := range k.topicChanelMap.List() {
		close(ch)
	}

	if err := k.consumerGroup.Close(); err != nil {
		log.Errorw("method", "KafkaEvent.Close", "err", err)
		return
	}
}

func (k *KafkaEvent) init() error {
	config := sarama.NewConfig()

	if k.c.GetVersion() == "" {
		kafkaVersion, err := sarama.ParseKafkaVersion(sarama.DefaultVersion.String())
		if err != nil {
			return err
		}
		config.Version = kafkaVersion
	} else {
		kafkaVersion, err := sarama.ParseKafkaVersion(k.c.GetVersion())
		if err != nil {
			return err
		}
		config.Version = kafkaVersion
	}

	switch k.c.GetStrategy() {
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	case "roundrobin":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	default:
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	}

	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	if k.c.GetSaslEnable() {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = k.c.GetUsername()
		config.Net.SASL.Password = k.c.GetPassword()
		config.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	}

	consumerGroup, err := sarama.NewConsumerGroup(strings.Split(k.c.GetBrokers(), ","), k.c.GetGroupName(), config)
	if err != nil {
		return merr.ErrorNotificationSystemError("Unrecognized consumer group partition assignor: %v", err)
	}
	k.consumerGroup = consumerGroup
	return nil
}

// ConsumeClaim 消费消息
func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Infof("Message topic:%q partition:%d offset:%d\n", message.Topic, message.Partition, message.Offset)
		c.messageChan <- &mq.Msg{
			Topic: []byte(message.Topic),
			Data:  message.Value,
		}
		// 将消息标记为已使用
		session.MarkMessage(message, "")
	}
	return nil
}

// Setup 初始化
func (c *consumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 清理
func (c *consumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
