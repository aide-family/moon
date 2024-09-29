package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model"
)

// NewUserMessageBiz .
func NewUserMessageBiz(userMessageRepository repository.UserMessage) *UserMessageBiz {
	return &UserMessageBiz{
		userMessageRepository: userMessageRepository,
	}
}

// UserMessageBiz .
type UserMessageBiz struct {
	userMessageRepository repository.UserMessage
}

// DeleteUserMessage 删除用户通知消息
func (b *UserMessageBiz) DeleteUserMessage(ctx context.Context, ids []uint32) error {
	return b.userMessageRepository.Delete(ctx, ids)
}

// DeleteAllUserMessage 删除用户通知消息-所有
func (b *UserMessageBiz) DeleteAllUserMessage(ctx context.Context) error {
	return b.userMessageRepository.DeleteAll(ctx)
}

// ListUserMessage 获取用户通知消息列表
func (b *UserMessageBiz) ListUserMessage(ctx context.Context, params *bo.QueryUserMessageListParams) ([]*model.SysUserMessage, error) {
	return b.userMessageRepository.List(ctx, params)
}
