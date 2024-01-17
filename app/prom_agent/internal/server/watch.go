package server

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"golang.org/x/sync/errgroup"
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
	}

	topics := []string{
		kafkaMqConf.GetStrategyGroupAllTopic(),
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
		defer after.Recover(w.log)
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				w.log.Info("[Watch] server tick")
				eg := new(errgroup.Group)
				eg.SetLimit(100)
				w.groups.Range(func(key, value any) bool {
					groupDetail, ok := value.(*api.GroupSimple)
					if !ok {
						return true
					}
					eg.Go(func() error {
						_, _ = w.loadService.Evaluate(context.Background(), &agent.EvaluateRequest{GroupList: []*api.GroupSimple{groupDetail}})
						return nil
					})
					return true
				})
				_ = eg.Wait()
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return w.onlineNotify()
}

func (w *Watch) Stop(_ context.Context) error {
	w.log.Info("[Watch] server stopping")
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

// 接受mq消息
func (w *Watch) receiveMessage() {
	consumer := w.kafkaMqServer.Consumer()
	go func() {
		defer after.Recover(w.log)
		events := consumer.Events()
		strategyGroupAllTopic := w.kafkaConf.GetStrategyGroupAllTopic()
		for event := range events {
			if consumer.IsClosed() {
				w.log.Warnf("consumer is closed")
				return
			}
			switch e := event.(type) {
			case *kafka.Message:
				w.log.Infow("topic", *e.TopicPartition.Topic, "key", string(e.Key), "value", string(e.Value))
				if string(e.Key) != w.kafkaConf.GetGroupId() {
					break
				}
				switch *e.TopicPartition.Topic {
				case strategyGroupAllTopic:
					w.log.Info("strategyGroupAllTopic", string(e.Value))
					// 把新规则刷进内存
					groupBytes := e.Value
					var groupDetail *api.GroupSimple
					if err := json.Unmarshal(groupBytes, &groupDetail); err != nil {
						w.log.Warnf("unmarshal groupList error: %s", err.Error())
						break
					}
					w.log.Info("groupDetail", groupDetail)
					w.groups.Store(groupDetail.GetId(), groupDetail)
				}
			}
		}
	}()
}
