package message

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewInbox)

// Inbox 读取当前用户的系统消息。消息是平台级的，不要求 X-Namespace。
type Inbox struct {
	messages repository.MessageRepo
	users    repository.UserRepo
	access   access.Checker
}

func NewInbox(messages repository.MessageRepo, users repository.UserRepo, checker access.Checker) *Inbox {
	return &Inbox{messages: messages, users: users, access: checker}
}

func (inbox *Inbox) List(ctx context.Context, pageNo, pageSize int32, unreadOnly bool) ([]bo.SystemMessage, int, int, int, int, error) {
	uid, email, err := inbox.recipient(ctx)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := inbox.messages.ListMessages(ctx, uid, email, p, s, unreadOnly)
	if err != nil {
		return nil, 0, 0, 0, 0, inbox.access.Wrap(err)
	}
	unread, err := inbox.messages.UnreadCount(ctx, uid, email)
	if err != nil {
		return nil, 0, 0, 0, 0, inbox.access.Wrap(err)
	}
	return items, total, p, s, unread, nil
}

func (inbox *Inbox) UnreadCount(ctx context.Context) (int, error) {
	uid, email, err := inbox.recipient(ctx)
	if err != nil {
		return 0, err
	}
	count, err := inbox.messages.UnreadCount(ctx, uid, email)
	return count, inbox.access.Wrap(err)
}

func (inbox *Inbox) MarkRead(ctx context.Context, uid string) error {
	userUID, email, err := inbox.recipient(ctx)
	if err != nil {
		return err
	}
	return inbox.access.Wrap(inbox.messages.MarkRead(ctx, userUID, email, uid))
}

func (inbox *Inbox) MarkAllRead(ctx context.Context) error {
	userUID, email, err := inbox.recipient(ctx)
	if err != nil {
		return err
	}
	return inbox.access.Wrap(inbox.messages.MarkAllRead(ctx, userUID, email))
}

func (inbox *Inbox) recipient(ctx context.Context) (string, string, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return "", "", merr.Unauthenticated("missing token")
	}
	account, err := inbox.users.UserByUID(ctx, uid)
	if err != nil {
		return "", "", inbox.access.Wrap(err)
	}
	return uid, account.Email, nil
}
