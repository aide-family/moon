package event

import (
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

var _ mq.IMQ = (*RabbitMQEvent)(nil)

type (

	// RabbitMQEvent 实现 IMQ 接口
	RabbitMQEvent struct {
		conn           *amqp.Connection
		channel        *amqp.Channel
		queue          *amqp.Queue
		c              *conf.RabbitMQ
		topicChanelMap *safety.Map[string, chan *mq.Msg]
	}
)

// NewRabbitMQEvent 创建 RabbitMQEvent 实例
func NewRabbitMQEvent(c *conf.RabbitMQ) (*RabbitMQEvent, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotificationSystemError("RabbitMQ 配置为空")
	}

	rabbitEvent := &RabbitMQEvent{
		c:              c,
		topicChanelMap: safety.NewMap[string, chan *mq.Msg](),
	}

	if err := rabbitEvent.init(); err != nil {
		return nil, err
	}
	return rabbitEvent, nil
}

func (r *RabbitMQEvent) Send(_ string, _ []byte) error {
	return nil
}

func (r *RabbitMQEvent) Receive(topic string) <-chan *mq.Msg {
	ch, ok := r.topicChanelMap.Get(topic)
	if ok {
		return ch
	}
	ch = make(chan *mq.Msg, 1000)
	r.topicChanelMap.Set(topic, ch)

	// 声明队列，确保消息的顺利接收
	queue, err := r.channel.QueueDeclare(
		topic, // 队列名称与topic一致
		true,  // 持久化
		false, // 非自动删除
		false, // 非排他
		false, // 非阻塞
		nil,   // 额外参数
	)
	if err != nil {
		r.topicChanelMap.Delete(topic)
		close(ch)
	}

	// 绑定队列
	err = r.channel.QueueBind(
		queue.Name,
		r.c.GetRoutingKey(), // routing key
		topic,
		false,
		nil,
	)

	r.queue = &queue

	if err != nil {
		r.topicChanelMap.Delete(topic)
		close(ch)
	}

	// 开启消费者接收消息
	go func() {
		// 协程异常或者通道关闭会引起此问题
		defer after.RecoverX()
		for {
			messages, err := r.receive(topic)
			if err != nil {
				log.Errorw("method", "RabbitMQEvent.Receive", "err", err)
				continue
			}
			for _, msg := range messages {
				ch <- msg
			}
		}
	}()
	return ch
}

func (r *RabbitMQEvent) receive(topic string) ([]*mq.Msg, error) {
	msgs, err := r.channel.Consume(
		r.queue.Name,
		topic, // 消费者标识符
		true,  // 自动应答
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.topicChanelMap.Delete(topic)
		return nil, err
	}
	msgList := make([]*mq.Msg, 0, len(msgs))
	for msg := range msgs {
		msgList = append(msgList, &mq.Msg{
			Topic: []byte(topic),
			Data:  msg.Body,
		})
	}
	return msgList, nil
}

func (r *RabbitMQEvent) RemoveReceiver(topic string) {
	defer func() {
		r.topicChanelMap.Delete(topic)
		ch, ok := r.topicChanelMap.Get(topic)
		if ok {
			close(ch)
		}
	}()

	if err := r.channel.Cancel(topic, false); err != nil {
		log.Errorw("method", "RocketMQEvent.RemoveReceiver", "err", err)
	}
}

func (r *RabbitMQEvent) Close() {
	// 关闭通道
	for _, ch := range r.topicChanelMap.List() {
		close(ch)
	}
	if err := r.conn.Close(); err != nil {
		log.Errorw("method", "RabbitMQEvent.Close", "conn err", err)
	}
	if err := r.channel.Close(); err != nil {
		log.Errorw("method", "RabbitMQEvent.Close", "channel err", err)
	}
}

func (r *RabbitMQEvent) init() error {
	amqpURL := r.c.GetUrl()
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return err
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}
	r.conn = conn
	r.channel = channel
	return nil
}
