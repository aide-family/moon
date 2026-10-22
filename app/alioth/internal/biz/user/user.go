package user

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewDirectory)

// Directory 管理平台用户。用户不属于租户，这里不读取 X-Namespace。
type Directory struct {
	users  repository.UserRepo
	access access.Checker
}

func NewDirectory(users repository.UserRepo, checker access.Checker) *Directory {
	return &Directory{users: users, access: checker}
}

func (directory *Directory) GetUser(ctx context.Context, uid string) (*bo.Account, error) {
	if _, err := directory.access.RequireSystem(ctx, perm.SystemUserRead); err != nil {
		return nil, err
	}
	account, err := directory.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, directory.access.Wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (directory *Directory) ListUsers(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Account, int, int, int, error) {
	if _, err := directory.access.RequireSystem(ctx, perm.SystemUserRead); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := directory.users.ListUsers(ctx, p, s, name)
	if err != nil {
		return nil, 0, 0, 0, directory.access.Wrap(err)
	}
	for i := range items {
		items[i] = publicAccount(items[i])
	}
	return items, total, p, s, nil
}

func (directory *Directory) SelectUsers(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	if _, err := directory.access.RequireSystem(ctx, perm.SystemUserRead); err != nil {
		return nil, err
	}
	items, err := directory.users.SelectUsers(ctx, page.Limit(limit), lastUID)
	return items, directory.access.Wrap(err)
}

// BanUser 封禁平台用户。封禁后登录和已有令牌都会被拒绝。
func (directory *Directory) BanUser(ctx context.Context, uid string) (*bo.Account, error) {
	return directory.setStatus(ctx, uid, "BANNED")
}

// PermitUser 恢复用户为可登录。
func (directory *Directory) PermitUser(ctx context.Context, uid string) (*bo.Account, error) {
	return directory.setStatus(ctx, uid, "ACTIVE")
}

func (directory *Directory) setStatus(ctx context.Context, uid, status string) (*bo.Account, error) {
	actorAccount, err := directory.access.RequireSystem(ctx, perm.SystemUserWrite)
	if err != nil {
		return nil, err
	}
	if actorAccount.UID == uid {
		return nil, merr.Client("cannot change your own status")
	}
	target, err := directory.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, directory.access.Wrap(err)
	}
	if target.SystemRole == perm.SystemRoleSuperAdmin {
		return nil, merr.Forbidden("cannot change super admin")
	}
	if actorAccount.SystemRole != perm.SystemRoleSuperAdmin && target.SystemRole == perm.SystemRoleSystemAdmin {
		return nil, merr.Forbidden("cannot change another system admin")
	}
	account, err := directory.users.UpdateUserStatus(ctx, uid, status)
	if err != nil {
		return nil, directory.access.Wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

// UpdateSystemRole 由超级管理员授权或撤销系统管理员。
// 超级管理员只能是最早注册的那个用户，不能通过接口再授予或转移。
func (directory *Directory) UpdateSystemRole(ctx context.Context, uid, role string) (*bo.Account, error) {
	actorAccount, err := directory.access.RequireSystem(ctx, perm.SystemRoleWrite)
	if err != nil {
		return nil, err
	}
	if actorAccount.UID == uid {
		return nil, merr.Client("cannot change your own system role")
	}
	if role != perm.SystemRoleSystemAdmin && role != perm.SystemRoleUser {
		return nil, merr.Invalid("invalid system role")
	}
	target, err := directory.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, directory.access.Wrap(err)
	}
	if target.SystemRole == perm.SystemRoleSuperAdmin {
		return nil, merr.Forbidden("cannot change super admin")
	}
	account, err := directory.users.UpdateUserSystemRole(ctx, uid, role, actorAccount.UID)
	if err != nil {
		return nil, directory.access.Wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func publicAccount(account bo.Account) bo.Account {
	account.PasswordHash = ""
	return account
}
