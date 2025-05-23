package service

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz"
	"github.com/aide-family/moon/cmd/palace/internal/service/build"
	"github.com/aide-family/moon/pkg/api/palace"
	"github.com/aide-family/moon/pkg/api/palace/common"
)

type TeamLogService struct {
	palace.UnimplementedTeamLogServer
	logsBiz *biz.Logs
}

func NewTeamLogService(logsBiz *biz.Logs) *TeamLogService {
	return &TeamLogService{
		logsBiz: logsBiz,
	}
}

func (s *TeamLogService) GetSendMessageLogs(ctx context.Context, req *palace.GetTeamSendMessageLogsRequest) (*palace.GetTeamSendMessageLogsReply, error) {
	listParams, err := build.ToListSendMessageLogParams(req)
	if err != nil {
		return nil, err
	}
	params, err := listParams.WithTeamID(ctx)
	if err != nil {
		return nil, err
	}
	logsReply, err := s.logsBiz.GetSendMessageLogs(ctx, params)
	if err != nil {
		return nil, err
	}
	return &palace.GetTeamSendMessageLogsReply{
		Items:      build.ToSendMessageLogs(logsReply.Items),
		Pagination: build.ToPaginationReply(logsReply.PaginationReply),
	}, nil
}

func (s *TeamLogService) GetSendMessageLog(ctx context.Context, req *palace.OperateOneTeamSendMessageRequest) (*common.SendMessageLogItem, error) {
	params, err := build.ToGetSendMessageLogParams(req.GetRequestId()).WithTeamID(ctx)
	if err != nil {
		return nil, err
	}
	logDo, err := s.logsBiz.GetSendMessageLog(ctx, params)
	if err != nil {
		return nil, err
	}
	return build.ToSendMessageLog(logDo), nil
}

func (s *TeamLogService) RetrySendMessage(ctx context.Context, req *palace.OperateOneTeamSendMessageRequest) (*common.EmptyReply, error) {
	reTryParams, err := build.ToRetrySendMessageParams(req)
	if err != nil {
		return nil, err
	}
	params, err := reTryParams.WithTeamID(ctx)
	if err != nil {
		return nil, err
	}
	if err = s.logsBiz.RetrySendMessage(ctx, params); err != nil {
		return nil, err
	}
	return &common.EmptyReply{}, nil
}
