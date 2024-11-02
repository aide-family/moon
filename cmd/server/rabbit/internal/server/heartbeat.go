package server

import (
	"context"
	"time"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/service"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/env"
	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*HeartbeatServer)(nil)

func newHeartbeatServer(bc *rabbitconf.Bootstrap, healthService *service.HealthService) *HeartbeatServer {
	server := bc.GetServer()
	network := vobj.ToNetwork(server.GetNetwork())
	return &HeartbeatServer{
		healthService: healthService,
		tick:          time.NewTicker(time.Second * 10),
		srv: &conf.MicroServer{
			Endpoint:    types.Ternary(network.IsRpc(), server.GetGrpcEndpoint(), server.GetHttpEndpoint()),
			Secret:      types.Of(server.GetSecret()),
			Timeout:     types.Ternary(network.IsRpc(), bc.GetGrpc().GetTimeout(), bc.GetHttp().GetTimeout()),
			Network:     server.GetNetwork(),
			NodeVersion: env.Version(),
			Name:        server.GetName(),
		},
		teamIds:      bc.GetTeams(),
		stopCh:       make(chan struct{}),
		dependPalace: bc.GetDependPalace(),
	}
}

type HeartbeatServer struct {
	healthService *service.HealthService

	tick         *time.Ticker
	srv          *conf.MicroServer
	teamIds      []uint32
	stopCh       chan struct{}
	dependPalace bool
}

func (h *HeartbeatServer) Start(ctx context.Context) error {
	go func() {
		defer after.RecoverX()
		if !h.dependPalace {
			return
		}
		log.Infof("[HeartbeatServer] server started")
		for {
			select {
			case <-h.stopCh:
				if err := h.healthService.Heartbeat(ctx, &api.HeartbeatRequest{Server: h.srv, TeamIds: h.teamIds, Online: false}); err != nil {
					log.Errorw("heartbeat error", err)
				}
				log.Infof("[HeartbeatServer] server stopped")
				return
			case <-h.tick.C:
				if err := h.healthService.Heartbeat(ctx, &api.HeartbeatRequest{Server: h.srv, TeamIds: h.teamIds, Online: true}); err != nil {
					log.Errorw("heartbeat error", err)
				}
			}
		}
	}()
	return nil
}

func (h *HeartbeatServer) Stop(_ context.Context) error {
	h.tick.Stop()
	close(h.stopCh)
	return nil
}
