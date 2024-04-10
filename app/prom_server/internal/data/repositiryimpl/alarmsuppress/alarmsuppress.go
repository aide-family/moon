package alarmsuppress

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
)

var _ repository.AlarmSuppressRepo = (*alarmSuppressImpl)(nil)

type alarmSuppressImpl struct {
	repository.UnimplementedAlarmSuppressRepo
	log  *log.Helper
	data *data.Data
}

func NewAlarmSuppress(data *data.Data, logger log.Logger) repository.AlarmSuppressRepo {
	return &alarmSuppressImpl{
		log:  log.NewHelper(log.With(logger, "module", "repository.alarm.suppress")),
		data: data,
	}
}
