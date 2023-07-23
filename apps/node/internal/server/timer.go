package server

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/strategy/v1/load"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/apps/node/internal/service"
	"prometheus-manager/pkg/servers"
	"time"
)

func NewTimer(
	conf *conf.Strategy,
	logger log.Logger,
	loadService *service.LoadService,
) *servers.Timer {
	ticker := time.NewTicker(conf.LoadInterval.AsDuration())
	var count int64

	call := func(ctx context.Context) error {
		count++
		log.Info("TimerCallFunc: ", count)
		reload, err := loadService.Reload(ctx, &pb.ReloadRequest{Nodes: nil})
		if err != nil {
			return err
		}

		log.Info("Reload: ", reload)
		return nil
	}

	return servers.NewTimer(call, ticker, logger)
}
