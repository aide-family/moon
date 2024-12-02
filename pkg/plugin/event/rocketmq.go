package event

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/apache/rocketmq-clients/golang"
	"github.com/apache/rocketmq-clients/golang/credentials"
)

const (
	// maximum waiting time for receive func
	awaitDuration = time.Second * 5
	// maximum number of messages received at one time
	maxMessageNum int32 = 16
	// invisibleDuration should > 20s
	invisibleDuration = time.Second * 20
	// receive messages in a loop
)

// NewRocketMQEvent 创建RocketMQ事件
func NewRocketMQEvent(c *conf.RocketMQ, opts ...RocketMQEventOption) (*RocketMQEvent, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotificationSystemError("RocketMQ 配置为空")
	}
	r := &RocketMQEvent{c: c, topicChanelMap: safety.NewMap[string, chan *mq.Msg]()}
	defaultOpts := []RocketMQEventOption{
		WithAwaitDuration(awaitDuration),
		WithMaxMessageNum(maxMessageNum),
		WithInvisibleDuration(invisibleDuration),
	}
	defaultOpts = append(defaultOpts, opts...)

	for _, opt := range defaultOpts {
		opt(r)
	}

	if err := r.init(); err != nil {
		return nil, err
	}
	return r, nil
}

var _ mq.IMQ = (*RocketMQEvent)(nil)

type (
	// RocketMQEvent RocketMQ事件
	RocketMQEvent struct {
		c *conf.RocketMQ

		consumer golang.SimpleConsumer

		topicChanelMap *safety.Map[string, chan *mq.Msg]

		awaitDuration     time.Duration
		maxMessageNum     int32
		invisibleDuration time.Duration
	}

	// RocketMQEventOption 选项
	RocketMQEventOption func(r *RocketMQEvent)
)

// Send 发送消息
func (r *RocketMQEvent) Send(_ string, _ []byte) error {
	return nil
}

// Receive 监听topic
func (r *RocketMQEvent) Receive(topic string) <-chan *mq.Msg {
	ch, ok := r.topicChanelMap.Get(topic)
	if ok {
		return ch
	}

	ch = make(chan *mq.Msg, 1000)
	r.topicChanelMap.Set(topic, ch)
	if err := r.consumer.Subscribe(topic, golang.SUB_ALL); err != nil {
		defer func() {
			r.topicChanelMap.Delete(topic)
			close(ch)
		}()
		return ch
	}
	go func() {
		// 协程异常或者通道关闭会引起此问题
		defer after.RecoverX()
		for {
			messages, err := r.receive(topic)
			if err != nil {
				log.Errorw("method", "RocketMQEvent.Receive", "err", err)
				continue
			}
			for _, msg := range messages {
				ch <- msg
			}
		}
	}()
	return ch
}

func (r *RocketMQEvent) receive(topic string) ([]*mq.Msg, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.awaitDuration)
	defer cancel()
	messages, err := r.consumer.Receive(ctx, r.maxMessageNum, r.awaitDuration)
	if err != nil {
		return nil, err
	}
	return r.ConvertMessageView(ctx, topic, messages), nil
}

// ConvertMessageView 转换消息视图
func (r *RocketMQEvent) ConvertMessageView(ctx context.Context, topic string, messages []*golang.MessageView) []*mq.Msg {
	msgList := make([]*mq.Msg, 0, len(messages))
	for _, msg := range messages {
		t := msg.GetTopic()
		if t != topic {
			continue
		}
		msgList = append(msgList, &mq.Msg{
			Topic: []byte(t),
			Data:  msg.GetBody(),
		})
		if err := r.consumer.Ack(ctx, msg); err != nil {
			log.Errorw("method", "RocketMQEvent.Ack", "err", err)
			continue
		}
	}
	return msgList
}

// RemoveReceiver 移除监听
func (r *RocketMQEvent) RemoveReceiver(topic string) {
	defer func() {
		r.topicChanelMap.Delete(topic)
		ch, ok := r.topicChanelMap.Get(topic)
		if ok {
			close(ch)
		}
	}()
	if err := r.consumer.Unsubscribe(topic); err != nil {
		log.Errorw("method", "RocketMQEvent.RemoveReceiver", "err", err)
		return
	}
}

// Close 关闭 mq
func (r *RocketMQEvent) Close() {
	if err := r.consumer.GracefulStop(); err != nil {
		log.Errorw("method", "RocketMQEvent.Close", "err", err)
		return
	}
}

func (r *RocketMQEvent) init() error {
	simpleConsumer, err := golang.NewSimpleConsumer(&golang.Config{
		Endpoint:      r.c.GetEndpoint(),
		NameSpace:     r.c.GetNamespace(),
		ConsumerGroup: r.c.GetGroupName(),
		Credentials: &credentials.SessionCredentials{
			AccessKey:    r.c.GetAccessKey(),
			AccessSecret: r.c.GetSecretKey(),
		},
	},
		golang.WithAwaitDuration(awaitDuration),
	)
	if err != nil {
		return err
	}
	r.consumer = simpleConsumer
	return r.consumer.Start()
}

// WithAwaitDuration 设置等待时间
func WithAwaitDuration(d time.Duration) RocketMQEventOption {
	return func(r *RocketMQEvent) {
		r.awaitDuration = d
	}
}

// WithMaxMessageNum 设置最大消息数
func WithMaxMessageNum(n int32) RocketMQEventOption {
	return func(r *RocketMQEvent) {
		r.maxMessageNum = n
	}
}

// WithInvisibleDuration 设置不可见时间
func WithInvisibleDuration(d time.Duration) RocketMQEventOption {
	return func(r *RocketMQEvent) {
		r.invisibleDuration = d
	}
}
