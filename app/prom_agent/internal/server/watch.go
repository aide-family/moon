package server

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"prometheus-manager/api"
	"prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/service"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/servers"
)

var _ transport.Server = (*Watch)(nil)

type EventHandler func(msg *kafka.Message) error

type Watch struct {
	conf      *conf.WatchProm
	kafkaConf *conf.Kafka
	ticker    *time.Ticker
	log       *log.Helper

	loadService   *service.LoadService
	kafkaMqServer *servers.KafkaMQServer

	groups        *sync.Map
	exitCh        chan struct{}
	eventHandlers map[consts.TopicType]EventHandler
}

func NewWatch(
	c *conf.WatchProm,
	mqConf *conf.MQ,
	loadService *service.LoadService,
	logger log.Logger,
) (*Watch, error) {
	kafkaConf := mqConf.GetKafka()
	kafkaMqServer, err := servers.NewKafkaMQServer(kafkaConf, logger)
	if err != nil {
		return nil, err
	}

	kafkaMqConf := mqConf.GetKafka()

	w := &Watch{
		conf:          c,
		exitCh:        make(chan struct{}, 1),
		ticker:        time.NewTicker(c.GetInterval().AsDuration()),
		log:           log.NewHelper(log.With(logger, "module", "server.watch")),
		loadService:   loadService,
		groups:        new(sync.Map),
		kafkaConf:     mqConf.GetKafka(),
		kafkaMqServer: kafkaMqServer,
		eventHandlers: make(map[consts.TopicType]EventHandler),
	}
	w.eventHandlers = map[consts.TopicType]EventHandler{
		consts.TopicType(kafkaMqConf.GetStrategyGroupAllTopic()): w.loadGroupAllEventHandler,
		consts.TopicType(kafkaMqConf.GetRemoveGroupTopic()):      w.removeGroupEventHandler,
	}

	topics := []string{
		kafkaMqConf.GetStrategyGroupAllTopic(),
		kafkaMqConf.GetRemoveGroupTopic(),
	}
	w.log.Info("topics", topics)
	if err = w.Subscribe(topics); err != nil {
		return nil, err
	}
	w.receiveMessage()

	return w, nil
}

func (w *Watch) Subscribe(topics []string) error {
	return w.kafkaMqServer.Consume(topics, w.handleMessage)
}

func (w *Watch) loadGroupAllEventHandler(msg *kafka.Message) error {
	w.log.Info("strategyGroupAllTopic", string(msg.Value))
	// 把新规则刷进内存
	groupBytes := msg.Value
	var groupDetail *api.GroupSimple
	if err := json.Unmarshal(groupBytes, &groupDetail); err != nil {
		w.log.Warnf("unmarshal groupList error: %s", err.Error())
		return err
	}
	w.log.Info("groupDetail", groupDetail)
	w.groups.Store(groupDetail.GetId(), groupDetail)
	return nil
}

func (w *Watch) removeGroupEventHandler(msg *kafka.Message) error {
	w.log.Info("removeGroupTopic", string(msg.Value))
	value := string(msg.Value)
	groupId, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		w.log.Warnf("parse groupId error: %s", value)
		return nil
	}
	if groupId > 0 {
		w.groups.Delete(uint32(groupId))
	}
	return nil
}

func (w *Watch) handleMessage(msg *kafka.Message) bool {
	topic := *msg.TopicPartition.Topic
	w.log.Infow("topic", topic)
	if handler, ok := w.eventHandlers[consts.TopicType(topic)]; ok {
		if err := handler(msg); err != nil {
			w.log.Errorf("handle message error: %s", err.Error())
		}
	}
	return true
}

func (w *Watch) Start(_ context.Context) error {
	go func() {
		//defer after.Recover(w.log, func(err error) {
		//	w.log.Errorf("recover error: %s", err.Error())
		//})
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				w.log.Info("[Watch] server tick")
				groupList := make([]*api.GroupSimple, 0)
				w.groups.Range(func(key, value any) bool {
					if group, ok := value.(*api.GroupSimple); ok && group != nil {
						groupList = append(groupList, group)
					}
					return true
				})
				_, _ = w.loadService.Evaluate(context.Background(), &agent.EvaluateRequest{GroupList: groupList})
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return w.onlineNotify()
}

func (w *Watch) Stop(_ context.Context) error {
	w.log.Info("[Watch] server stopping")
	if err := w.offlineNotify(); err != nil {
		return err
	}
	close(w.exitCh)
	return nil
}

func (w *Watch) shutdown() {
	w.groups = nil
	w.ticker.Stop()
	w.log.Info("[Watch] server stopped")
}

// onlineNotify 上线通知
func (w *Watch) onlineNotify() error {
	w.log.Info("[Watch] server online notify")
	topic := w.kafkaConf.GetOnlineTopic()
	return w.kafkaMqServer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(w.kafkaConf.GetStrategyGroupAllTopic()),
		Key:   []byte(w.kafkaConf.GetGroupId()),
	})
}

// offlineNotify 下线通知
func (w *Watch) offlineNotify() error {
	w.log.Info("[Watch] server offline notify")
	topic := w.kafkaConf.GetOfflineTopic()
	w.log.Infow("topic", topic)
	err := w.kafkaMqServer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: nil,
		Key:   []byte(w.kafkaConf.GetGroupId()),
	})
	if err != nil {
		return err
	}
	// 等待1秒，等kafka消费完消息
	time.Sleep(1 * time.Second)
	return nil
}

// 接受mq消息
func (w *Watch) receiveMessage() {
	consumer := w.kafkaMqServer.Consumer()
	go func() {
		defer after.Recover(w.log, func(err error) {
			w.log.Warnf("receiveMessage panic")
		})
		events := consumer.Events()
		for event := range events {
			if consumer.IsClosed() {
				w.log.Warnf("consumer is closed")
				return
			}
			switch e := event.(type) {
			case *kafka.Message:
				w.log.Info("=========================")
				w.log.Infow("topic", *e.TopicPartition.Topic, "key", string(e.Key), "value", string(e.Value))
				if string(e.Key) != w.kafkaConf.GetGroupId() {
					break
				}
				handle, ok := w.eventHandlers[consts.TopicType(*e.TopicPartition.Topic)]
				if !ok {
					w.log.Warnf("topic not found: %s", *e.TopicPartition.Topic)
					break
				}
				if err := handle(e); err != nil {
					w.log.Warnf("handle message error: %s", err.Error())
				}
			}
		}
	}()
}
