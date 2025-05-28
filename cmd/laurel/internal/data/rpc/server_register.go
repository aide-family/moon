package rpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/cmd/laurel/internal/data"
	common "github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/plugin/server"
)

func NewServerRegisterRepo(bc *conf.Bootstrap, data *data.Data, logger log.Logger) (repository.ServerRegister, error) {
	palaceConfig := bc.GetPalace()
	s := &serverRegisterRepo{
		Data:    data,
		enable:  palaceConfig.GetEnable(),
		network: palaceConfig.GetNetwork(),
		helper:  log.NewHelper(log.With(logger, "module", "data.repo.server_register")),
	}

	initConfig := &server.InitConfig{
		MicroConfig: palaceConfig,
		Registry:    bc.GetRegistry(),
	}

	return s, s.initClient(initConfig)
}

type serverRegisterRepo struct {
	*data.Data
	enable     bool
	network    config.Network
	rpcClient  common.ServerClient
	httpClient common.ServerHTTPClient

	helper *log.Helper
}

func (r *serverRegisterRepo) initClient(initConfig *server.InitConfig) error {
	if !r.enable {
		return nil
	}
	switch r.network {
	case config.Network_GRPC:
		conn, err := server.InitGRPCClient(initConfig)
		if err != nil {
			return err
		}
		r.rpcClient = common.NewServerClient(conn)
	case config.Network_HTTP:
		client, err := server.InitHTTPClient(initConfig)
		if err != nil {
			return err
		}
		r.httpClient = common.NewServerHTTPClient(client)
	}
	return nil
}

func (r *serverRegisterRepo) Register(ctx context.Context, server *common.ServerRegisterRequest) error {
	if !r.enable {
		return nil
	}
	var (
		reply *common.ServerRegisterReply
		err   error
	)
	switch r.network {
	case config.Network_GRPC:
		reply, err = r.rpcClient.Register(ctx, server)
	case config.Network_HTTP:
		reply, err = r.httpClient.Register(ctx, server)
	}
	if err != nil {
		r.helper.WithContext(ctx).Errorw("msg", "Register failed", "error", err, "reply", reply)
		return err
	}
	return nil
}
