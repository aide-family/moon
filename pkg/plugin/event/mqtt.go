package event

import (
	"time"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/houyi/mq"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/random"
	"github.com/aide-family/moon/pkg/util/safety"
	"github.com/aide-family/moon/pkg/util/types"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	// DisconnectTime 断开重连时间
	DisconnectTime uint = 250
)

// NewMqttEvent 创建MQTT 事件
func NewMqttEvent(c *conf.MQTT, opts ...MqttEventOption) (*MqttEvent, error) {
	if types.IsNil(c) {
		return nil, merr.ErrorNotificationSystemError("MQTT 配置为空")
	}

	mqttEvent := &MqttEvent{
		c:              c,
		topicChanelMap: safety.NewMap[string, chan *mq.Msg](),
	}

	defaultOpts := []MqttEventOption{
		WithDisconnectTime(DisconnectTime),
	}

	defaultOpts = append(defaultOpts, opts...)
	for _, opt := range defaultOpts {
		opt(mqttEvent)
	}
	if err := mqttEvent.init(); err != nil {
		return nil, err
	}
	return mqttEvent, nil
}

var _ mq.IMQ = (*MqttEvent)(nil)

type (
	// MqttEvent MQTT 事件
	MqttEvent struct {
		c *conf.MQTT

		client mqtt.Client

		topicChanelMap *safety.Map[string, chan *mq.Msg]

		DisconnectTime uint
	}

	// MqttEventOption 选项
	MqttEventOption func(r *MqttEvent)
)

// Send 发送消息
func (m *MqttEvent) Send(_ string, _ []byte) error {
	return nil
}

// Receive 接收消息
func (m *MqttEvent) Receive(topic string) <-chan *mq.Msg {
	ch, ok := m.topicChanelMap.Get(topic)
	if ok {
		return ch
	}
	ch = make(chan *mq.Msg, 1000)
	m.topicChanelMap.Set(topic, ch)

	// 订阅消息
	if token := m.client.Subscribe(topic, byte(m.c.GetQos()), func(client mqtt.Client, message mqtt.Message) {
		log.Info("MqttEvent.Receive", "topic", string(message.Topic()), "payload", string(message.Payload()))
		ch <- &mq.Msg{
			Topic:     []byte(message.Topic()),
			Data:      message.Payload(),
			Timestamp: types.NewTime(time.Now()),
		}
	}); token.Wait() && token.Error() != nil {
		defer func() {
			m.topicChanelMap.Delete(topic)
			close(ch)
		}()
	}
	return ch
}

// RemoveReceiver 移除接收者
func (m *MqttEvent) RemoveReceiver(topic string) {
	defer func() {
		m.topicChanelMap.Delete(topic)
		ch, ok := m.topicChanelMap.Get(topic)
		if ok {
			close(ch)
		}
	}()

	if err := m.client.Unsubscribe(topic); err != nil {
		log.Errorw("method", "MqttEvent.RemoveReceiver", "err", err)
		return
	}
}

// Close 关闭
func (m *MqttEvent) Close() {
	// 关闭通道
	for _, ch := range m.topicChanelMap.List() {
		close(ch)
	}
	m.client.Disconnect(m.DisconnectTime)
}

// WithDisconnectTime 设置断开重连时间
func WithDisconnectTime(quiesce uint) MqttEventOption {
	return func(m *MqttEvent) {
		m.DisconnectTime = quiesce
	}
}

func (m *MqttEvent) init() error {
	// 设置 MQTT 客户端选项
	opts := mqtt.NewClientOptions()
	opts.AddBroker(m.c.GetBroker())
	opts.SetClientID(random.UUIDToUpperCase(true))
	opts.SetUsername(m.c.GetUsername())
	opts.SetAutoReconnect(m.c.GetAutoReconnect())

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return merr.ErrorNotificationSystemError("failed to connect to MQTT broker: %v", token.Error())
	}
	m.client = client
	return nil
}
