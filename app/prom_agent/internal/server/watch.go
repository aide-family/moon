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
	"prometheus-manager/api/agent"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/service"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/conn"
)

var _ transport.Server = (*Watch)(nil)

type Watch struct {
	conf      *conf.WatchProm
	kafkaConf *conf.Kafka
	ticker    *time.Ticker
	producer  *kafka.Producer
	consumer  *kafka.Consumer
	log       *log.Helper

	loadService *service.LoadService

	groups *sync.Map
	exitCh chan struct{}
}

func NewWatch(
	c *conf.WatchProm,
	mqConf *conf.MQ,
	loadService *service.LoadService,
	logger log.Logger,
) (*Watch, error) {
	producer, err := conn.NewKafkaProducer(mqConf.GetKafka().GetEndpoints())
	if err != nil {
		return nil, err
	}
	consumer, err := conn.NewKafkaConsumer(mqConf.GetKafka().GetEndpoints(), mqConf.GetKafka().GetGroupId())
	if err != nil {
		return nil, err
	}
	return &Watch{
		conf:        c,
		exitCh:      make(chan struct{}, 1),
		ticker:      time.NewTicker(c.GetInterval().AsDuration()),
		log:         log.NewHelper(log.With(logger, "module", "server.watch")),
		loadService: loadService,
		groups:      new(sync.Map),
		kafkaConf:   mqConf.GetKafka(),
		producer:    producer,
		consumer:    consumer,
	}, nil
}

func (w *Watch) Start(_ context.Context) error {
	if err := w.onlineNotify(); err != nil {
		return err
	}
	w.receiveMessage()
	go func() {
		defer after.Recover(w.log)
		for {
			select {
			case <-w.exitCh:
				w.shutdown()
				return
			case <-w.ticker.C:
				eg := new(errgroup.Group)
				eg.SetLimit(100)
				w.groups.Range(func(key, value any) bool {
					groupDetail, ok := value.(*agent.GroupSimple)
					if !ok {
						return true
					}
					eg.Go(func() error {
						_, _ = w.loadService.Evaluate(context.Background(), &agent.EvaluateRequest{GroupList: []*agent.GroupSimple{groupDetail}})
						return nil
					})
					return true
				})
				_ = eg.Wait()
			}
		}
	}()
	w.log.Info("[Watch] server started")
	return nil
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
	topic := w.kafkaConf.GetOnlineTopic()
	return w.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: []byte(w.kafkaConf.GetStrategyGroupAllTopic()),
		Key:   []byte(w.kafkaConf.GetGroupId()),
	}, nil)
}

// 接受mq消息
func (w *Watch) receiveMessage() {
	go func() {
		defer after.Recover(w.log)
		events := w.consumer.Events()
		strategyGroupAllTopic := w.kafkaConf.GetStrategyGroupAllTopic()
		for event := range events {
			if w.consumer.IsClosed() {
				w.log.Warnf("consumer is closed")
				return
			}
			switch e := event.(type) {
			case *kafka.Message:
				if string(e.Key) != w.kafkaConf.GetGroupId() {
					break
				}
				switch *e.TopicPartition.Topic {
				case strategyGroupAllTopic:
					// 把新规则刷进内存
					groupBytes := e.Value
					var groupList []*agent.GroupSimple
					if err := json.Unmarshal(groupBytes, &groupList); err != nil {
						break
					}
					for _, group := range groupList {
						w.groups.Store(group.GetId(), group)
					}
				}
			}
		}
	}()
}
