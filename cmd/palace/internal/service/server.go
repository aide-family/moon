package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
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

func (s *ServerService) Sync(ctx context.Context, req *bo.SyncRequest) error {
	switch req.Type {
	case vobj.ChangedTypeMetricDatasource:
		return s.serverBiz.SyncMetricDatasource(ctx, bo.ChangedMetricDatasource(req.Rows))
	case vobj.ChangedTypeMetricStrategy:
		return s.serverBiz.SyncMetricStrategy(ctx, bo.ChangedMetricStrategy(req.Rows))
	case vobj.ChangedTypeNoticeGroup:
		return s.serverBiz.SyncNoticeGroup(ctx, bo.ChangedNoticeGroup(req.Rows))
	case vobj.ChangedTypeNoticeSMSConfig:
		return s.serverBiz.SyncNoticeSMSConfig(ctx, bo.ChangedNoticeSMSConfig(req.Rows))
	case vobj.ChangedTypeNoticeEmailConfig:
		return s.serverBiz.SyncNoticeEmailConfig(ctx, bo.ChangedNoticeEmailConfig(req.Rows))
	case vobj.ChangedTypeNoticeHookConfig:
		return s.serverBiz.SyncNoticeHookConfig(ctx, bo.ChangedNoticeHookConfig(req.Rows))
	default:
		s.helper.WithContext(ctx).Warnf("unknown sync type: %v", req.Type)
		return nil
	}
}
