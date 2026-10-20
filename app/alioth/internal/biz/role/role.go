package role

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
	roles  repository.RoleRepo
	access access.Checker
}

func NewUsecase(roles repository.RoleRepo, checker access.Checker) *Usecase {
	return &Usecase{roles: roles, access: checker}
}

func (uc *Usecase) CreateRole(ctx context.Context, name string) (*bo.Role, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.roles.CreateRole(ctx, ns, name)
	return item, uc.access.Wrap(err)
}

// UpdateRole 只改自定义角色的名称。owner、admin、member、viewer 是系统角色，改名会破坏预置边界。
func (uc *Usecase) UpdateRole(ctx context.Context, uid, name string) (*bo.Role, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	current, err := uc.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if current.IsSystem {
		return nil, merr.Client("invalid request")
	}
	item, err := uc.roles.UpdateRole(ctx, ns, uid, name)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateRoleStatus(ctx context.Context, uid, status string) (*bo.Role, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	current, err := uc.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if current.IsSystem {
		return nil, merr.Client("invalid request")
	}
	item, err := uc.roles.UpdateRoleStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

// DeleteRole 不删除系统角色。删掉 owner 会让租户失去负责人角色。
func (uc *Usecase) DeleteRole(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return err
	}
	current, err := uc.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return uc.access.Wrap(err)
	}
	if current.IsSystem {
		return merr.Client("invalid request")
	}
	return uc.access.Wrap(uc.roles.DeleteRole(ctx, ns, uid))
}

func (uc *Usecase) GetRole(ctx context.Context, uid string) (*bo.Role, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.roles.GetRole(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListRoles(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Role, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.roles.ListRoles(ctx, ns, p, s, name)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) SelectRoles(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.roles.SelectRoles(ctx, ns, page.Limit(limit), lastUID)
	return items, uc.access.Wrap(err)
}

// UpdateRolePermissions 替换自定义角色的权限。
// namespace:delete 和 namespace:transfer 只属于 owner，不能授给自定义角色。
// 系统角色的权限由预置目录决定，这里直接拒绝。
func (uc *Usecase) UpdateRolePermissions(ctx context.Context, uid string, names []string) (*bo.Role, error) {
	ns, err := uc.access.Allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	current, err := uc.roles.GetRole(ctx, ns, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if current.IsSystem {
		return nil, merr.Client("invalid request")
	}
	for _, name := range names {
		if name == perm.NamespaceDelete || name == perm.NamespaceTransfer {
			return nil, merr.Client("invalid request")
		}
	}
	item, err := uc.roles.UpdateRolePermissions(ctx, ns, uid, names)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListPermissions(ctx context.Context) ([]bo.Permission, error) {
	if _, err := uc.access.Allow(ctx, perm.RoleRead); err != nil {
		return nil, err
	}
	items, err := uc.roles.ListPermissions(ctx)
	return items, uc.access.Wrap(err)
}

// Mine 返回当前用户在当前租户的权限名。没有选择租户时返回空列表，而不是报错。
func (uc *Usecase) Mine(ctx context.Context) ([]string, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	namespace, ok := actor.NamespaceID(ctx)
	if !ok {
		return []string{}, nil
	}
	perms, err := uc.roles.Permissions(ctx, namespace, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	names := make([]string, 0, len(perms))
	for name := range perms {
		names = append(names, name)
	}
	return names, nil
}
