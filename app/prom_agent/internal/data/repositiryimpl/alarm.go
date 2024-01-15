package repositiryimpl

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_agent/internal/biz/do"
	"prometheus-manager/app/prom_agent/internal/biz/repository"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/data"
)

var _ repository.AlarmRepo = (*alarmRepoImpl)(nil)

type alarmRepoImpl struct {
	log  *log.Helper
	data *data.Data

	producerConf *conf.Kafka
}

func (l *alarmRepoImpl) Alarm(_ context.Context, alarmDo *do.AlarmDo) error {
	if l.data.Producer() == nil {
		return status.Error(codes.Unavailable, "producer is not ready")
	}

	if len(l.producerConf.GetTopics()) == 0 {
		return status.Error(codes.Unavailable, "topics is not ready")
	}

	if alarmDo == nil {
		return status.Error(codes.InvalidArgument, "alarm do is nil")
	}

	for _, topic := range l.producerConf.GetTopics() {
		msg := l.genMsg(alarmDo, topic)
		if err := l.data.Producer().Produce(msg, nil); err != nil {
			l.log.Errorf("failed to produce message to topic %s: %v", topic, err)
			continue
		}
	}

	return nil
}

func (l *alarmRepoImpl) genMsg(alarmDo *do.AlarmDo, topic string) *kafka.Message {
	return &kafka.Message{
		Key: []byte(alarmDo.Receiver),
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: alarmDo.Bytes(),
	}
}

func NewAlarmRepo(data *data.Data, producerConf *conf.Kafka, logger log.Logger) repository.AlarmRepo {
	return &alarmRepoImpl{
		log:          log.NewHelper(log.With(logger, "module", "alarmRepoImpl")),
		data:         data,
		producerConf: producerConf,
	}
}
