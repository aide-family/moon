package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/merr"
)

type ServerService struct {
	common.UnimplementedServerServer

	serverBiz *biz.ServerBiz

	helper *log.Helper
}

func NewServerService(serverBiz *biz.ServerBiz, logger log.Logger) *ServerService {
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
