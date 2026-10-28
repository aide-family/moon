package user

import (
	"context"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewDirectory)

const usernameMaxLength = 8

// usernamePattern 是注册用户名的规则：只允许中英文和数字，长度 1 到 8 个字符。
// 与 api/alioth/v1/identity.proto 上 RegisterRequest.name 的 buf.validate 注解保持一致。
var usernamePattern = regexp.MustCompile(`^[\p{L}\p{N}]{1,8}$`)

// ValidateUsername 校验注册用户名。只用于新建用户，已存在的用户名不回溯校验，
// 避免历史账号因为规则收紧而无法登录。
func ValidateUsername(name string) error {
	if !usernamePattern.MatchString(name) {
		return merr.Client("invalid username")
	}
	return nil
}

// DefaultUsername 在用户没填用户名时从邮箱前缀派生一个合法用户名：
// 只保留中英文和数字，截断到 8 个字符。保留旧行为（未填名称用邮箱前缀），
// 同时保证派生结果也符合注册规则。前缀没有可用字符时退回 user。
func DefaultUsername(email string) string {
	prefix := email
	if at := strings.IndexByte(email, '@'); at >= 0 {
		prefix = email[:at]
	}
	name := make([]rune, 0, usernameMaxLength)
	for _, letter := range prefix {
		if !unicode.IsLetter(letter) && !unicode.IsNumber(letter) {
			continue
		}
		name = append(name, letter)
		if len(name) == usernameMaxLength {
			break
		}
	}
	if len(name) == 0 {
		return "user"
	}
	return string(name)
}

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
	return account, nil
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
	return account, nil
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
	return account, nil
}
