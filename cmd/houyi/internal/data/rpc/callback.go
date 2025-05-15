package rpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/houyi/internal/conf"
	"github.com/moon-monitor/moon/pkg/api/palace"
	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/plugin/server"
)

func NewCallbackRepo(bc *conf.Bootstrap, logger log.Logger) (repository.Callback, error) {
	palaceConfig := bc.GetPalace()
	s := &callbackRepo{
		network: palaceConfig.GetNetwork(),
		helper:  log.NewHelper(log.With(logger, "module", "data.repo.callback")),
	}

	initConfig := &server.InitConfig{
		MicroConfig: palaceConfig,
		Registry:    bc.GetRegistry(),
	}

	return s, s.initClient(initConfig)
}

type callbackRepo struct {
	network    config.Network
	rpcClient  palace.CallbackClient
	httpClient palace.CallbackHTTPClient

	helper *log.Helper
}

func (r *callbackRepo) initClient(initConfig *server.InitConfig) error {
	switch r.network {
	case config.Network_GRPC:
		conn, err := server.InitGRPCClient(initConfig)
		if err != nil {
			return err
		}
		r.rpcClient = palace.NewCallbackClient(conn)
	case config.Network_HTTP:
		client, err := server.InitHTTPClient(initConfig)
		if err != nil {
			return err
		}
		r.httpClient = palace.NewCallbackHTTPClient(client)
	}
	return nil
}

func (r *callbackRepo) SyncMetadata(ctx context.Context, req *palace.SyncMetadataRequest) error {
	var (
		reply *palace.SyncMetadataReply
		err   error
	)
	switch r.network {
	case config.Network_GRPC:
		reply, err = r.rpcClient.SyncMetadata(ctx, req)
	case config.Network_HTTP:
		reply, err = r.httpClient.SyncMetadata(ctx, req)
	}
	if err != nil {
		r.helper.WithContext(ctx).Errorw("msg", "SyncMetadata failed", "error", err, "reply", reply)
		return err
	}
	return nil
}
