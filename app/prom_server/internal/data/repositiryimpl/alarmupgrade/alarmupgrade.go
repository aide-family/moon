package alarmupgrade

import (
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/data"
)

var _ repository.AlarmUpgradeRepo = (*alarmUpgradeImpl)(nil)

type alarmUpgradeImpl struct {
	repository.UnimplementedAlarmUpgradeRepo
	log  *log.Helper
	data *data.Data
}

func NewAlarmUpgrade(data *data.Data, logger log.Logger) repository.AlarmUpgradeRepo {
	return &alarmUpgradeImpl{
		log:  log.NewHelper(log.With(logger, "module", "repository.alarm.upgrade")),
		data: data,
	}
}
