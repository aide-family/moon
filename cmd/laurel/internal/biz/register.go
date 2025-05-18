package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"

	"github.com/aide-family/moon/cmd/laurel/internal/biz/repository"
	"github.com/aide-family/moon/cmd/laurel/internal/conf"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/hello"
	"github.com/aide-family/moon/pkg/util/pointer"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewRegisterBiz(bc *conf.Bootstrap, serverRegisterRepo repository.ServerRegister, logger log.Logger) *RegisterBiz {
	r := &RegisterBiz{
		serverRegisterRepo: serverRegisterRepo,
		bc:                 bc,
		uuid:               uuid.New().String(),
		helper:             log.NewHelper(log.With(logger, "module", "biz.register")),
	}
	r.registerParams = r.register(false)
	return r
}

type RegisterBiz struct {
	uuid               string
	bc                 *conf.Bootstrap
	serverRegisterRepo repository.ServerRegister
	registerParams     *common.ServerRegisterRequest
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
		Discovery:  nil,
		TeamIds:    serverConfig.GetTeamIds(),
		IsOnline:   online,
		Uuid:       b.uuid,
		ServerType: common.ServerRegisterRequest_LAUREL,
	}
	switch serverConfig.GetNetwork() {
	case config.Network_GRPC:
		params.Server.Timeout = serverConfig.GetGrpc().GetTimeout()
	case config.Network_HTTP:
		params.Server.Timeout = serverConfig.GetHttp().GetTimeout()
	}
	register := b.bc.GetRegistry()
	if validate.IsNotNil(register) && register.GetEnable() {
		params.Discovery = &config.Discovery{
			Driver: register.GetDriver(),
			Enable: register.GetEnable(),
			Etcd:   register.GetEtcd(),
		}
	}
	return params
}

func (b *RegisterBiz) Online(ctx context.Context) error {
	b.registerParams.IsOnline = true
	return b.serverRegisterRepo.Register(ctx, b.registerParams)
}

func (b *RegisterBiz) Offline(ctx context.Context) error {
	b.registerParams.IsOnline = false
	return b.serverRegisterRepo.Register(ctx, b.registerParams)
}
