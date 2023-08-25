package data

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	"prometheus-manager/pkg/alert"

	"prometheus-manager/apps/node/internal/biz"
	"prometheus-manager/apps/node/internal/conf"
)

type (
	AlertRepo struct {
		logger *log.Helper
		data   *Data
	}
)

var _ biz.IAlertRepo = (*AlertRepo)(nil)

func NewAlertRepo(data *Data, logger log.Logger) *AlertRepo {
	return &AlertRepo{data: data, logger: log.NewHelper(log.With(logger, "module", alertModuleName))}
}

func (l *AlertRepo) V1(ctx context.Context) string {
	ctx, span := otel.Tracer(alertModuleName).Start(ctx, "AlertRepo.V1")
	defer span.End()
	return "AlertRepo.V1"
}

func (l *AlertRepo) SyncAlert(ctx context.Context, alertData *alert.Data) error {
	ctx, span := otel.Tracer(alertModuleName).Start(ctx, "AlertRepo.SyncAlert")
	defer span.End()
	if !l.data.kafkaConf.GetEnable() {
		l.logger.Warnf("Not enabel kafka")
		return nil
	}
	topic := l.data.kafkaConf.GetAlertTopic()
	return l.data.kafkaProducer.Produce(&kafka.Message{
		Key:   []byte(conf.Get().GetEnv().GetName()),
		Value: alertData.Byte(),
		TopicPartition: kafka.TopicPartition{
			Topic: &topic,
		},
	})
}
