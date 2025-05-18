package rpc

import (
	"context"

	"github.com/aide-family/moon/cmd/houyi/internal/data"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/config"
)

func NewServerRegisterRepo(d *data.Data, logger log.Logger) repository.ServerRegister {
	return &serverRegisterRepo{
		Data:         d,
		palaceServer: d.GetPlaceServer(),
		helper:       log.NewHelper(log.With(logger, "module", "data.repo.server_register")),
	}
}

type serverRegisterRepo struct {
	*data.Data
	palaceServer *data.Server

	helper *log.Helper
}

func (r *serverRegisterRepo) Register(ctx context.Context, server *common.ServerRegisterRequest) error {
	var (
		reply *common.ServerRegisterReply
		err   error
	)
	switch r.palaceServer.GetNetWork() {
	case config.Network_GRPC:
		reply, err = r.palaceServer.GetServerClient().Register(ctx, server)
	case config.Network_HTTP:
		reply, err = r.palaceServer.GetServerHTTPClient().Register(ctx, server)
	}
	if err != nil {
		r.helper.WithContext(ctx).Errorw("msg", "Register failed", "error", err, "reply", reply)
		return err
	}
	return nil
}
