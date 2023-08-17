package server

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/node"
	"time"

	"prometheus-manager/pkg/servers"

	"prometheus-manager/apps/master/internal/conf"
	"prometheus-manager/apps/master/internal/service"
)

func NewTimer(
	pushStrategy *conf.PushStrategy,
	logger log.Logger,
	pushService *service.PushService,
) *servers.Timer {
	interval := pushStrategy.GetInterval().AsDuration()
	if interval <= 0 {
		interval = time.Second * 10
	}
	ticker := time.NewTicker(interval)
	var count int64
	loggerHelper := log.NewHelper(log.With(logger, "module", "server/Timer"))

	srvList := make([]*pb.NodeServer, 0, len(pushStrategy.GetNodes()))
	for _, srv := range pushStrategy.GetNodes() {
		srvList = append(srvList, &pb.NodeServer{
			ServerName: srv.GetServerName(),
			Timeout:    srv.GetTimeout(),
			Network:    srv.GetNetwork(),
		})
	}

	call := func(ctx context.Context) {
		if !pushStrategy.GetEnable() {
			return
		}
		count++
		loggerHelper.Info("TimerCallFunc: ", count)
		//return
		pushed, err := pushService.Call(ctx, &pb.CallRequest{Servers: srvList})
		if err != nil {
			loggerHelper.Errorf("[Timer] call error: %v", err)
			return
		}

		log.Info("pushed: ", pushed)
	}

	return servers.NewTimer(call, ticker, logger)
}
