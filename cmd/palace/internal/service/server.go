package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/merr"
)

type ServerService struct {
	common.UnimplementedServerServer

	serverBiz *biz.Server

	helper *log.Helper
}

func NewServerService(serverBiz *biz.Server, logger log.Logger) *ServerService {
	return &ServerService{
		serverBiz: serverBiz,
		helper:    log.NewHelper(log.With(logger, "module", "service.server")),
	}
}

func (s *ServerService) Register(ctx context.Context, req *common.ServerRegisterRequest) (*common.ServerRegisterReply, error) {
	boReq := build.ToServerRegisterReq(req)
	if boReq == nil {
		return nil, merr.ErrorInvalidArgument("invalid request")
	}
	if err := s.serverBiz.Register(ctx, boReq); err != nil {
		return nil, err
	}
	return &common.ServerRegisterReply{}, nil
}
