package server

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	pb "prometheus-manager/api/strategy/v1/load"

	"prometheus-manager/pkg/servers"

	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
)

func NewTimer(
	reloadStrategy *conf.Strategy,
	logger log.Logger,
	loadService *service.LoadService,
) *servers.Timer {
	ticker := time.NewTicker(reloadStrategy.GetLoadInterval().AsDuration())
	var count int64
	loggerHelper := log.NewHelper(log.With(logger, "module", "server/Timer"))

	call := func(ctx context.Context) {
		if !reloadStrategy.GetEnable() {
			return
		}
		count++
		log.Info("TimerCallFunc: ", count)
		reload, err := loadService.Reload(ctx, &pb.ReloadRequest{Nodes: nil})
		if err != nil {
			loggerHelper.Errorf("[Timer] call error: %v", err)
		}

		log.Info("Reload: ", reload)
	}

	return servers.NewTimer(call, ticker, logger)
}
