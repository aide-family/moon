package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewUserMessageBiz .
func NewUserMessageBiz(userMessageRepository repository.UserMessage, inviteBiz *InviteBiz) *UserMessageBiz {
	return &UserMessageBiz{
		userMessageRepository: userMessageRepository,
		inviteBiz:             inviteBiz,
	}
}

// UserMessageBiz .
type UserMessageBiz struct {
	userMessageRepository repository.UserMessage

	inviteBiz *InviteBiz
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

// ConfirmUserMessage 确认用户通知消息
//
//	根据不同的业务类型，完成不同的业务动作
func (b *UserMessageBiz) ConfirmUserMessage(ctx context.Context, id uint32) error {
	// 查询用户消息
	userMessage, err := b.userMessageRepository.GetById(ctx, id)
	if err != nil {
		return err
	}
	userMessage.Biz = vobj.BizTypeInvitationAccepted
	if err := b.confirmCall(ctx, userMessage); err != nil {
		return err
	}
	return b.userMessageRepository.Delete(ctx, []uint32{id})
}

// CancelUserMessage 取消用户通知消息
func (b *UserMessageBiz) CancelUserMessage(ctx context.Context, id uint32) error {
	// 查询用户消息
	userMessage, err := b.userMessageRepository.GetById(ctx, id)
	if err != nil {
		return err
	}
	userMessage.Biz = vobj.BizTypeInvitationRejected
	if err := b.cancelCall(ctx, userMessage); err != nil {
		return err
	}
	return b.userMessageRepository.Delete(ctx, []uint32{id})
}

func (b *UserMessageBiz) confirmCall(ctx context.Context, msg *model.SysUserMessage) error {
	switch msg.Biz {
	case vobj.BizTypeInvitationAccepted:
		return b.userAgreeJoinTeam(ctx, msg.BizID)
	default:
		return nil
	}
}

func (b *UserMessageBiz) cancelCall(ctx context.Context, msg *model.SysUserMessage) error {
	switch msg.Biz {
	case vobj.BizTypeInvitationAccepted:
		return b.userRefuseJoinTeam(ctx, msg.BizID)
	default:
		return nil
	}
}

// 用户同意加入团队
func (b *UserMessageBiz) userAgreeJoinTeam(ctx context.Context, id uint32) error {
	return b.inviteBiz.UpdateInviteStatus(ctx, &bo.UpdateInviteStatusParams{
		InviteID:   id,
		InviteType: vobj.InviteTypeJoined,
	})
}

// 用户拒绝加入团队
func (b *UserMessageBiz) userRefuseJoinTeam(ctx context.Context, id uint32) error {
	return b.inviteBiz.UpdateInviteStatus(ctx, &bo.UpdateInviteStatusParams{
		InviteID:   id,
		InviteType: vobj.InviteTypeRejected,
	})
}
