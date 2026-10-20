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

// GetUser 读取当前租户内的用户。User 是全局实体，但租户管理员只能看到与自己同租户的用户。
func (uc *Usecase) GetUser(ctx context.Context, uid string) (*bo.Account, error) {
	ns, err := uc.access.Allow(ctx, perm.UserRead)
	if err != nil {
		return nil, err
	}
	ok, err := uc.users.UserInNamespace(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if !ok {
		return nil, merr.NotFound("not found")
	}
	account, err := uc.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *Usecase) ListUsers(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Account, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.UserRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.users.ListUsers(ctx, ns, p, s, name)
	if err != nil {
		return nil, 0, 0, 0, uc.access.Wrap(err)
	}
	for i := range items {
		items[i] = publicAccount(items[i])
	}
	return items, total, p, s, nil
}

func (uc *Usecase) SelectUsers(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.UserRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.users.SelectUsers(ctx, ns, page.Limit(limit), lastUID)
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

// setStatus 不允许用户改自己的状态，也不允许跨租户封禁/解封。
// 全局 User 的管理权限来自当前租户，因此目标用户必须已加入当前租户。
func (uc *Usecase) setStatus(ctx context.Context, uid, status string) (*bo.Account, error) {
	ns, err := uc.access.Allow(ctx, perm.UserWrite)
	if err != nil {
		return nil, err
	}
	actorUID, _ := actor.UserID(ctx)
	if actorUID == uid {
		return nil, merr.Client("cannot change your own status")
	}
	ok, err := uc.users.UserInNamespace(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if !ok {
		return nil, merr.Forbidden("user is not in current namespace")
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
