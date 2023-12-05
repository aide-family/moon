package alarmsuppress

import (
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.AlarmSuppressRepo = (*alarmSuppressImpl)(nil)

type alarmSuppressImpl struct {
	repository.UnimplementedAlarmSuppressRepo
	log *log.Helper
	d   *data.Data
}

func NewAlarmSuppress(data *data.Data, logger log.Logger) repository.AlarmSuppressRepo {
	return &alarmSuppressImpl{
		log: log.NewHelper(log.With(logger, "module", "repository.alarm.suppress")),
		d:   data,
	}
}
