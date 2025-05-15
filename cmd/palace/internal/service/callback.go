package service

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz"
	"github.com/moon-monitor/moon/cmd/palace/internal/helper/permission"
	"github.com/moon-monitor/moon/cmd/palace/internal/service/build"
	"github.com/moon-monitor/moon/pkg/api/palace"
)

func NewCallbackService(logsBiz *biz.Logs, teamDatasource *biz.TeamDatasource) *CallbackService {
	return &CallbackService{
		logsBiz:        logsBiz,
		teamDatasource: teamDatasource,
	}
}

type CallbackService struct {
	palace.UnimplementedCallbackServer
	logsBiz        *biz.Logs
	teamDatasource *biz.TeamDatasource
}

func (s *CallbackService) SendMsgCallback(ctx context.Context, req *palace.SendMsgCallbackRequest) (*palace.SendMsgCallbackReply, error) {
	params := build.ToUpdateSendMessageLogStatusParams(req)
	if err := s.logsBiz.UpdateSendMessageLogStatus(ctx, params); err != nil {
		return nil, err
	}
	return &palace.SendMsgCallbackReply{
		Code: 0,
		Msg:  "success",
	}, nil
}

func (s *CallbackService) SyncMetadata(ctx context.Context, req *palace.SyncMetadataRequest) (*palace.SyncMetadataReply, error) {
	ctx = permission.WithTeamIDContext(ctx, req.GetTeamId())
	ctx = permission.WithUserIDContext(ctx, req.GetOperatorId())
	batchSaveMetadata := build.ToBatchSaveTeamMetricDatasourceMetadataRequest(req)
	if err := s.teamDatasource.BatchSaveMetricDatasourceMetadata(ctx, batchSaveMetadata); err != nil {
		return nil, err
	}

	return &palace.SyncMetadataReply{
		Code: 0,
		Msg:  "success",
	}, nil
}
