package repositiryimpl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"prometheus-manager/app/prom_agent/internal/biz/do"
	"prometheus-manager/app/prom_agent/internal/biz/repository"
	"prometheus-manager/app/prom_agent/internal/conf"
	"prometheus-manager/app/prom_agent/internal/data"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/interflow"
)

var _ repository.AlarmRepo = (*alarmRepoImpl)(nil)

type alarmRepoImpl struct {
	log           *log.Helper
	data          *data.Data
	interflowConf *conf.Interflow
}

func (l *alarmRepoImpl) Alarm(_ context.Context, alarmDo *do.AlarmDo) error {
	if l.data.Interflow() == nil {
		return status.Error(codes.Unavailable, "interflow is not ready")
	}

	if alarmDo == nil {
		return status.Error(codes.InvalidArgument, "alarm do is nil")
	}

	if len(alarmDo.Alerts) == 0 {
		return nil
	}

	msg := l.genMsg(alarmDo, string(consts.AlertHookTopic))
	if err := l.data.Interflow().Send(context.Background(), l.interflowConf.GetServer(), msg); err != nil {
		l.log.Errorf("failed to produce message to topic %s: %v", msg.Topic, err)
		return err
	}
	return nil
}

func (l *alarmRepoImpl) genMsg(alarmDo *do.AlarmDo, topic string) *interflow.HookMsg {
	serverUrl := l.interflowConf.GetServer()
	return &interflow.HookMsg{
		Topic: topic,
		Value: alarmDo.Bytes(),
		Key:   []byte(serverUrl),
	}
}

func NewAlarmRepo(data *data.Data, interflowConf *conf.Interflow, logger log.Logger) repository.AlarmRepo {
	return &alarmRepoImpl{
		log:           log.NewHelper(log.With(logger, "module", "alarmRepoImpl")),
		data:          data,
		interflowConf: interflowConf,
	}
}
