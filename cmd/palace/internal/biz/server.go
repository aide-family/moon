package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/validate"
)

func NewServerBiz(serverRepo repository.Server, logger log.Logger) *Server {
	return &Server{
		serverRepo: serverRepo,
		helper:     log.NewHelper(log.With(logger, "module", "biz.server")),
	}
}

type Server struct {
	serverRepo repository.Server
	helper     *log.Helper
}

func (b *Server) Register(ctx context.Context, req *bo.ServerRegisterReq) error {
	if validate.IsNil(req) {
		return merr.ErrorInvalidArgument("invalid request")
	}

	if !req.IsOnline {
		return b.Deregister(ctx, req)
	}

	if err := b.serverRepo.RegisterServer(ctx, req); err != nil {
		return merr.ErrorInternalServer("failed to register server: %v", err)
	}

	b.helper.WithContext(ctx).Debugf("registered server type: %v, uuid: %s", req.ServerType, req.Uuid)
	return nil
}

func (b *Server) Deregister(ctx context.Context, req *bo.ServerRegisterReq) error {
	if validate.IsNil(req) {
		return merr.ErrorInvalidArgument("invalid request")
	}
	if req.IsOnline {
		return b.Register(ctx, req)
	}
	if err := b.serverRepo.DeregisterServer(ctx, req); err != nil {
		return merr.ErrorInternalServer("failed to deregister server: %v", err)
	}
	b.helper.WithContext(ctx).Debugf("deregistered server type: %v, uuid: %s", req.ServerType, req.Uuid)
	return nil
}
