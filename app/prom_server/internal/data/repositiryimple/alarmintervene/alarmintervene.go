package alarmintervene

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.AlarmInterveneRepo = (*alarmInterveneImpl)(nil)

type alarmInterveneImpl struct {
	repository.UnimplementedAlarmInterveneRepo
	log *log.Helper
	d   *data.Data
}

func NewAlarmIntervene(data *data.Data, logger log.Logger) repository.AlarmInterveneRepo {
	return &alarmInterveneImpl{
		log: log.NewHelper(log.With(logger, "module", "repository.alarm.intervene")),
		d:   data,
	}
}
