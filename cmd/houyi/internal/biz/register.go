package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/houyi/internal/conf"
	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/config"
	"github.com/moon-monitor/moon/pkg/hello"
	"github.com/moon-monitor/moon/pkg/util/pointer"
)

func NewRegisterBiz(bc *conf.Bootstrap, serverRegisterRepo repository.ServerRegister, logger log.Logger) *RegisterBiz {
	return &RegisterBiz{
		serverRegisterRepo: serverRegisterRepo,
		bc:                 bc,
		uuid:               uuid.New().String(),
		helper:             log.NewHelper(log.With(logger, "module", "biz.register")),
	}
}

type RegisterBiz struct {
	uuid               string
	bc                 *conf.Bootstrap
	serverRegisterRepo repository.ServerRegister
	helper             *log.Helper
}

func (b *RegisterBiz) register(online bool) *common.ServerRegisterRequest {
	serverConfig := b.bc.GetServer()
	jwtConf := b.bc.GetAuth().GetJwt()
	params := &common.ServerRegisterRequest{
		Server: &config.MicroServer{
			Endpoint: serverConfig.GetOutEndpoint(),
			Secret:   pointer.Of(jwtConf.GetSignKey()),
			Timeout:  nil,
			Network:  serverConfig.GetNetwork(),
			Version:  hello.GetEnv().Version(),
			Name:     serverConfig.GetName(),
		},
		Discovery: nil,
		TeamIds:   serverConfig.GetTeamIds(),
		IsOnline:  online,
		Uuid:      b.uuid,
	}
	switch serverConfig.GetNetwork() {
	case config.Network_GRPC:
		params.Server.Timeout = serverConfig.GetGrpc().GetTimeout()
	case config.Network_HTTP:
		params.Server.Timeout = serverConfig.GetHttp().GetTimeout()
	}
	register := b.bc.GetRegistry()
	if register != nil {
		params.Discovery = &config.Discovery{
			Driver: register.GetDriver(),
			Enable: register.GetEnable(),
			Etcd:   register.GetEtcd(),
		}
	}
	return params
}

func (b *RegisterBiz) Online(ctx context.Context) error {
	return b.serverRegisterRepo.Register(ctx, b.register(true))
}

func (b *RegisterBiz) Offline(ctx context.Context) error {
	return b.serverRegisterRepo.Register(ctx, b.register(false))
}
