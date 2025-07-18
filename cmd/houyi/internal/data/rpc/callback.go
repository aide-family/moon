package rpc

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/houyi/internal/biz/repository"
	"github.com/aide-family/moon/cmd/houyi/internal/conf"
	"github.com/aide-family/moon/cmd/houyi/internal/data"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/config"
)

func NewCallbackRepo(bc *conf.Bootstrap, d *data.Data, logger log.Logger) repository.Callback {
	return &callbackRepo{
		Data:         d,
		enable:       bc.GetPalace().GetEnable(),
		palaceServer: d.GetPlaceServer(),
		helper:       log.NewHelper(log.With(logger, "module", "data.repo.callback")),
	}
}

type callbackRepo struct {
	*data.Data
	enable       bool
	palaceServer *data.Server
	helper       *log.Helper
}

func (r *callbackRepo) SyncMetadata(ctx context.Context, req *palace.SyncMetadataRequest) error {
	if !r.enable {
		r.helper.WithContext(ctx).Infow("msg", "SyncMetadata disabled")
		return nil
	}
	var (
		reply *palace.SyncMetadataReply
		err   error
	)

	switch r.palaceServer.GetNetWork() {
	case config.Network_GRPC:
		reply, err = r.palaceServer.GetCallbackClient().SyncMetadata(ctx, req)
	case config.Network_HTTP:
		reply, err = r.palaceServer.GetCallbackHTTPClient().SyncMetadata(ctx, req)
	}
	if err != nil {
		r.helper.WithContext(ctx).Errorw("msg", "SyncMetadata failed", "error", err, "reply", reply)
		return err
	}
	return nil
}
