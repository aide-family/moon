package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/message"
)

// SystemMessageService 是平台系统消息入口。
type SystemMessageService struct {
	v1.UnimplementedSystemMessageServiceServer
	messages *message.Usecase
}

func NewSystemMessageService(uc *message.Usecase) *SystemMessageService {
	return &SystemMessageService{messages: uc}
}

func (s *SystemMessageService) ListSelfMessages(ctx context.Context, req *v1.ListSelfMessagesRequest) (*v1.ListSelfMessagesResponse, error) {
	items, total, page, pageSize, unread, err := s.messages.List(ctx, req.GetPage(), req.GetPageSize(), req.GetUnreadOnly())
	if err != nil {
		return nil, err
	}
	return &v1.ListSelfMessagesResponse{
		Items: systemMessageItems(items), Total: int32(total), Page: int32(page),
		PageSize: int32(pageSize), Unread: int32(unread),
	}, nil
}

func (s *SystemMessageService) GetUnreadMessageCount(ctx context.Context, _ *v1.GetUnreadMessageCountRequest) (*v1.GetUnreadMessageCountResponse, error) {
	count, err := s.messages.UnreadCount(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetUnreadMessageCountResponse{Count: int32(count)}, nil
}

func (s *SystemMessageService) MarkMessageRead(ctx context.Context, req *v1.MarkMessageReadRequest) (*v1.MarkMessageReadResponse, error) {
	if err := s.messages.MarkRead(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.MarkMessageReadResponse{}, nil
}

func (s *SystemMessageService) MarkAllMessagesRead(ctx context.Context, _ *v1.MarkAllMessagesReadRequest) (*v1.MarkAllMessagesReadResponse, error) {
	if err := s.messages.MarkAllRead(ctx); err != nil {
		return nil, err
	}
	return &v1.MarkAllMessagesReadResponse{}, nil
}
