package user

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	users  repository.UserRepo
	access access.Checker
}

func NewUsecase(users repository.UserRepo, checker access.Checker) *Usecase {
	return &Usecase{users: users, access: checker}
}

func (uc *Usecase) GetUser(ctx context.Context, uid string) (*bo.Account, error) {
	if _, err := uc.access.Allow(ctx, perm.UserRead); err != nil {
		return nil, err
	}
	account, err := uc.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *Usecase) ListUsers(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Account, int, int, int, error) {
	if _, err := uc.access.Allow(ctx, perm.UserRead); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.users.ListUsers(ctx, p, s, name)
	if err != nil {
		return nil, 0, 0, 0, uc.access.Wrap(err)
	}
	for i := range items {
		items[i] = publicAccount(items[i])
	}
	return items, total, p, s, nil
}

func (uc *Usecase) SelectUsers(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	if _, err := uc.access.Allow(ctx, perm.UserRead); err != nil {
		return nil, err
	}
	items, err := uc.users.SelectUsers(ctx, page.Limit(limit), lastUID)
	return items, uc.access.Wrap(err)
}

// BanUser 把用户标为封禁。封禁后 Authenticate 和登录发令牌都会拒绝该用户。
func (uc *Usecase) BanUser(ctx context.Context, uid string) (*bo.Account, error) {
	return uc.setStatus(ctx, uid, "BANNED")
}

// PermitUser 恢复用户为可登录。这不是危险确认流程的一部分，调用即生效。
func (uc *Usecase) PermitUser(ctx context.Context, uid string) (*bo.Account, error) {
	return uc.setStatus(ctx, uid, "ACTIVE")
}

// setStatus 不允许用户改自己的状态，避免管理员把自己封死后无法恢复。
func (uc *Usecase) setStatus(ctx context.Context, uid, status string) (*bo.Account, error) {
	if _, err := uc.access.Allow(ctx, perm.UserWrite); err != nil {
		return nil, err
	}
	actorUID, _ := actor.UserID(ctx)
	if actorUID == uid {
		return nil, merr.Client("cannot change your own status")
	}
	account, err := uc.users.UpdateUserStatus(ctx, uid, status)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func publicAccount(account bo.Account) bo.Account {
	account.PasswordHash = ""
	return account
}
