package server

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/apps/master/internal/service"
	"prometheus-manager/pkg/servers"
	"time"
)

func NewTimer(
	conf *conf.PushStrategy,
	logger log.Logger,
	pushService *service.PushService,
) *servers.Timer {
	interval := conf.GetInterval().AsDuration()
	if interval <= 0 {
		interval = time.Second * 10
	}
	ticker := time.NewTicker(interval)
	var count int64

	call := func(ctx context.Context) error {
		count++
		//log.Info("TimerCallFunc: ", count)
		//pushed, err := pushService.Call(ctx, &pb.CallRequest{Name: "prometheus-manager-node"})
		//if err != nil {
		//	return err
		//}

		//log.Info("pushed: ", pushed)
		return nil
	}

	return servers.NewTimer(call, ticker, logger)
}
