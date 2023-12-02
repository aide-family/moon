package biz

import (
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/repository"
)

type AlarmRealtime struct {
	log *log.Helper

	dataRepo      repository.DataRepo
	realtimeRepo  repository.AlarmRealtimeRepo
	interveneRepo repository.AlarmInterveneRepo
	upgradeRepo   repository.AlarmUpgradeRepo
	suppressRepo  repository.AlarmSuppressRepo
}

func NewAlarmRealtime(
	dataRepo repository.DataRepo,
	realtimeRepo repository.AlarmRealtimeRepo,
	interveneRepo repository.AlarmInterveneRepo,
	upgradeRepo repository.AlarmUpgradeRepo,
	suppressRepo repository.AlarmSuppressRepo,
	logger log.Logger,
) *AlarmRealtime {
	return &AlarmRealtime{
		log: log.NewHelper(log.With(logger, "module", "biz.AlarmRealtime")),

		dataRepo:      dataRepo,
		realtimeRepo:  realtimeRepo,
		interveneRepo: interveneRepo,
		upgradeRepo:   upgradeRepo,
		suppressRepo:  suppressRepo,
	}
}
