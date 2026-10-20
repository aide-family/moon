package message

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

// Usecase 读取当前用户的系统消息。消息是平台级的，不要求 X-Namespace。
type Usecase struct {
	messages repository.MessageRepo
	users    repository.UserRepo
	access   access.Checker
}

func NewUsecase(messages repository.MessageRepo, users repository.UserRepo, checker access.Checker) *Usecase {
	return &Usecase{messages: messages, users: users, access: checker}
}

func (uc *Usecase) List(ctx context.Context, pageNo, pageSize int32, unreadOnly bool) ([]bo.SystemMessage, int, int, int, int, error) {
	uid, email, err := uc.recipient(ctx)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.messages.ListMessages(ctx, uid, email, p, s, unreadOnly)
	if err != nil {
		return nil, 0, 0, 0, 0, uc.access.Wrap(err)
	}
	unread, err := uc.messages.UnreadCount(ctx, uid, email)
	if err != nil {
		return nil, 0, 0, 0, 0, uc.access.Wrap(err)
	}
	return items, total, p, s, unread, nil
}

func (uc *Usecase) UnreadCount(ctx context.Context) (int, error) {
	uid, email, err := uc.recipient(ctx)
	if err != nil {
		return 0, err
	}
	count, err := uc.messages.UnreadCount(ctx, uid, email)
	return count, uc.access.Wrap(err)
}

func (uc *Usecase) MarkRead(ctx context.Context, uid string) error {
	userUID, email, err := uc.recipient(ctx)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.messages.MarkRead(ctx, userUID, email, uid))
}

func (uc *Usecase) MarkAllRead(ctx context.Context) error {
	userUID, email, err := uc.recipient(ctx)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.messages.MarkAllRead(ctx, userUID, email))
}

func (uc *Usecase) recipient(ctx context.Context) (string, string, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return "", "", merr.Unauthenticated("missing token")
	}
	account, err := uc.users.UserByUID(ctx, uid)
	if err != nil {
		return "", "", uc.access.Wrap(err)
	}
	return uid, account.Email, nil
}
