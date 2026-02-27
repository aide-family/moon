package service

import (
	"context"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz"
	"github.com/aide-family/rabbit/internal/biz/bo"
	apiv1 "github.com/aide-family/rabbit/pkg/api/v1"
)

func NewMessageLogService(messageLogBiz *biz.MessageLog) *MessageLogService {
	return &MessageLogService{
		messageLogBiz: messageLogBiz,
	}
}

type MessageLogService struct {
	apiv1.UnimplementedMessageLogServer
	messageLogBiz *biz.MessageLog
}

func (s *MessageLogService) RetryMessage(ctx context.Context, req *apiv1.RetryMessageLogRequest) (*apiv1.RetryMessageLogReply, error) {
	err := s.messageLogBiz.RetryMessage(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return &apiv1.RetryMessageLogReply{}, nil
}

func (s *MessageLogService) CancelMessage(ctx context.Context, req *apiv1.CancelMessageLogRequest) (*apiv1.CancelMessageLogReply, error) {
	err := s.messageLogBiz.CancelMessage(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return &apiv1.CancelMessageLogReply{}, nil
}

func (s *MessageLogService) GetMessageLog(ctx context.Context, req *apiv1.GetMessageLogRequest) (*apiv1.MessageLogItem, error) {
	messageLogBo, err := s.messageLogBiz.GetMessageLog(ctx, snowflake.ParseInt64(req.Uid))
	if err != nil {
		return nil, err
	}
	return messageLogBo.ToAPIV1MessageLogItem(), nil
}

func (s *MessageLogService) ListMessageLog(ctx context.Context, req *apiv1.ListMessageLogRequest) (*apiv1.ListMessageLogReply, error) {
	listBo := bo.NewListMessageLogBo(req)
	pageResponseBo, err := s.messageLogBiz.ListMessageLog(ctx, listBo)
	if err != nil {
		return nil, err
	}
	return bo.ToAPIV1ListMessageLogReply(pageResponseBo), nil
}
