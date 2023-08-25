package server

import (
	"context"
	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	pb "prometheus-manager/api/alert/v1"

	"prometheus-manager/pkg/conn"

	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/apps/master/internal/service"
)

type WatchServer struct {
	watchService *service.WatchService
	logger       *log.Helper
	consumer     *conn.KafkaConsumer
}

var _ transport.Server = (*WatchServer)(nil)

func NewWatchServer(kafkaConf *conf.Kafka, watchService *service.WatchService, logger log.Logger) (*WatchServer, error) {
	logHelper := log.NewHelper(log.With(logger, "module", "server/server"))
	consumer, err := conn.NewKafkaConsumer(kafkaConf.GetEndpoints(), []string{kafkaConf.GetAlertTopic()}, log.DefaultLogger)
	if err != nil {
		logHelper.Error("kafka消费者初始化失败")
		return nil, err
	}

	return &WatchServer{
		watchService: watchService,
		logger:       logHelper,
		consumer:     consumer,
	}, nil
}

func (l *WatchServer) Start(ctx context.Context) error {
	l.logger.Info("[WatchServer] server starting")
	l.consumer.Consume(func(msg *kafka.Message) (flag bool) {
		flag = l.callbackFunc(ctx, msg)
		return
	})

	return nil
}

func (l *WatchServer) Stop(_ context.Context) error {
	defer l.logger.Info("[WatchServer] server stopped")
	return l.consumer.Close()
}

func (l *WatchServer) callbackFunc(ctx context.Context, msg *kafka.Message) (flag bool) {
	flag = true
	l.logger.Infof("message at topic:%v time:%d %s = %s\n", msg.TopicPartition, msg.Timestamp.Unix(), string(msg.Key), string(msg.Value))

	var req pb.WatchRequest
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		l.logger.Errorf("kafka消息错误: %v", err)
		return
	}

	alert, err := l.watchService.WatchAlert(ctx, &req)
	if err != nil {
		l.logger.Errorf("kafka消费失败: %v", err)
		return
	}
	l.logger.Infof("kafka消费成功, %v", alert)
	return
}
