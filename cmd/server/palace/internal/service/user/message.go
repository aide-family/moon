package user

import (
	"context"

	userapi "github.com/aide-family/moon/api/admin/user"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/service/builder"
	"github.com/aide-family/moon/pkg/util/types"
)

// MessageService 用户消息操作服务
type MessageService struct {
	userapi.UnimplementedMessageServer

	userMessageBiz *biz.UserMessageBiz
}

// NewMessageService 创建用户消息操作服务
func NewMessageService(userMessageBiz *biz.UserMessageBiz) *MessageService {
	return &MessageService{
		userMessageBiz: userMessageBiz,
	}
}

// DeleteMessages 删除用户消息
func (s *MessageService) DeleteMessages(ctx context.Context, req *userapi.DeleteMessagesRequest) (*userapi.DeleteMessagesReply, error) {
	if req.GetAll() {
		return &userapi.DeleteMessagesReply{}, s.userMessageBiz.DeleteAllUserMessage(ctx)
	}
	if err := s.userMessageBiz.DeleteUserMessage(ctx, req.GetIds()); err != nil {
		return nil, err
	}
	return &userapi.DeleteMessagesReply{}, nil
}

// ListMessage 获取用户消息列表
func (s *MessageService) ListMessage(ctx context.Context, req *userapi.ListMessageRequest) (*userapi.ListMessageReply, error) {
	params := &bo.QueryUserMessageListParams{
		Keyword: req.GetKeyword(),
		Page:    types.NewPagination(req.GetPagination()),
	}
	list, err := s.userMessageBiz.ListUserMessage(ctx, params)
	if err != nil {
		return nil, err
	}
	build := builder.NewParamsBuild(ctx)
	return &userapi.ListMessageReply{
		List:       build.UserModuleBuilder().NoticeUserMessageBuilder().ToAPIs(list),
		Pagination: build.PaginationModuleBuilder().ToAPI(params.Page),
	}, nil
}

// ConfirmMessage 确认用户消息
func (s *MessageService) ConfirmMessage(ctx context.Context, req *userapi.ConfirmMessageRequest) (*userapi.ConfirmMessageReply, error) {
	if err := s.userMessageBiz.ConfirmUserMessage(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &userapi.ConfirmMessageReply{}, nil
}

// CancelMessage 取消用户消息
func (s *MessageService) CancelMessage(ctx context.Context, req *userapi.CancelMessageRequest) (*userapi.CancelMessageReply, error) {
	if err := s.userMessageBiz.CancelUserMessage(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &userapi.CancelMessageReply{}, nil
}
