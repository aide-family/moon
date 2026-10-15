package biz

import (
	"context"
	"errors"

	"github.com/aide-family/Alioth/app/alioth/internal/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

func (uc *IdentityUsecase) CreateNamespace(ctx context.Context, name, remark string) (*Namespace, error) {
	uid, ok := UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	item, err := uc.repo.CreateNamespace(ctx, name, remark, uid)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*Namespace, error) {
	if err := uc.sameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := uc.require(ctx, perm.NamespaceWrite); err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateNamespace(ctx, uid, name, remark, logo)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) UpdateNamespaceStatus(ctx context.Context, uid, status string) (*Namespace, error) {
	if err := uc.sameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := uc.require(ctx, perm.NamespaceWrite); err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateNamespaceStatus(ctx, uid, status)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) DeleteNamespace(ctx context.Context, uid string) error {
	if err := uc.sameNamespace(ctx, uid); err != nil {
		return err
	}
	if err := uc.require(ctx, perm.NamespaceDelete); err != nil {
		return err
	}
	return uc.wrap(uc.repo.DeleteNamespace(ctx, uid))
}

func (uc *IdentityUsecase) GetNamespace(ctx context.Context, uid string) (*Namespace, error) {
	if err := uc.sameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := uc.require(ctx, perm.NamespaceRead); err != nil {
		return nil, err
	}
	item, err := uc.repo.GetNamespace(ctx, uid)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) ListNamespaces(ctx context.Context, page, pageSize int32, name string) ([]Namespace, int, int, int, error) {
	if err := uc.require(ctx, perm.NamespaceRead); err != nil {
		return nil, 0, 0, 0, err
	}
	uid, _ := UserID(ctx)
	p, s := normalizePage(page, pageSize)
	items, total, err := uc.repo.ListNamespaces(ctx, uid, p, s, name)
	return items, total, p, s, uc.wrap(err)
}

func (uc *IdentityUsecase) SelectNamespaces(ctx context.Context, limit int32, lastUID string) ([]SelectItem, error) {
	if err := uc.require(ctx, perm.NamespaceRead); err != nil {
		return nil, err
	}
	uid, _ := UserID(ctx)
	items, err := uc.repo.SelectNamespaces(ctx, uid, normalizeLimit(limit), lastUID)
	return items, uc.wrap(err)
}

func (uc *IdentityUsecase) TransferOwner(ctx context.Context, uid, userUID string) (*Namespace, error) {
	if err := uc.sameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := uc.require(ctx, perm.NamespaceTransfer); err != nil {
		return nil, err
	}
	item, err := uc.repo.TransferOwner(ctx, uid, userUID)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) ListMembers(ctx context.Context, page, pageSize int32, name string) ([]Member, int, int, int, error) {
	ns, err := uc.allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := normalizePage(page, pageSize)
	items, total, err := uc.repo.ListMembers(ctx, ns, p, s, name)
	return items, total, p, s, uc.wrap(err)
}

func (uc *IdentityUsecase) GetMember(ctx context.Context, uid string) (*Member, error) {
	ns, err := uc.allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetMember(ctx, ns, uid)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) ListUserMembers(ctx context.Context, userUID string) ([]Member, error) {
	ns, err := uc.allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.MemberByUser(ctx, ns, userUID)
	if err != nil {
		if errorsIsNotFound(err) {
			return []Member{}, nil
		}
		return nil, uc.wrap(err)
	}
	return []Member{*item}, nil
}

func (uc *IdentityUsecase) SelectMembers(ctx context.Context, limit int32, lastUID string) ([]SelectItem, error) {
	ns, err := uc.allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.repo.SelectMembers(ctx, ns, normalizeLimit(limit), lastUID)
	return items, uc.wrap(err)
}

func (uc *IdentityUsecase) InviteMember(ctx context.Context, email, roleUID string) (*Member, error) {
	ns, err := uc.allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.InviteMember(ctx, ns, normalizeEmail(email), roleUID)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) DeleteMember(ctx context.Context, uid string) error {
	ns, err := uc.allow(ctx, perm.MemberWrite)
	if err != nil {
		return err
	}
	return uc.wrap(uc.repo.DeleteMember(ctx, ns, uid))
}

func (uc *IdentityUsecase) UpdateMemberStatus(ctx context.Context, uid, status string) (*Member, error) {
	ns, err := uc.allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateMemberStatus(ctx, ns, uid, status)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) UpdateMemberRoles(ctx context.Context, uid string, roleUIDs []string) (*Member, error) {
	ns, err := uc.allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	for _, roleUID := range roleUIDs {
		role, err := uc.repo.GetRole(ctx, ns, roleUID)
		if err != nil {
			return nil, uc.wrap(err)
		}
		if role.Name == perm.RoleOwner {
			if err := uc.require(ctx, perm.NamespaceTransfer); err != nil {
				return nil, err
			}
		}
	}
	item, err := uc.repo.UpdateMemberRoles(ctx, ns, uid, roleUIDs)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) GetUser(ctx context.Context, uid string) (*Account, error) {
	if _, err := uc.allow(ctx, perm.UserRead); err != nil {
		return nil, err
	}
	account, err := uc.repo.UserByUID(ctx, uid)
	if err != nil {
		return nil, uc.wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *IdentityUsecase) ListUsers(ctx context.Context, page, pageSize int32, name string) ([]Account, int, int, int, error) {
	if _, err := uc.allow(ctx, perm.UserRead); err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := normalizePage(page, pageSize)
	items, total, err := uc.repo.ListUsers(ctx, p, s, name)
	if err != nil {
		return nil, 0, 0, 0, uc.wrap(err)
	}
	for i := range items {
		items[i] = publicAccount(items[i])
	}
	return items, total, p, s, nil
}

func (uc *IdentityUsecase) SelectUsers(ctx context.Context, limit int32, lastUID string) ([]SelectItem, error) {
	if _, err := uc.allow(ctx, perm.UserRead); err != nil {
		return nil, err
	}
	items, err := uc.repo.SelectUsers(ctx, normalizeLimit(limit), lastUID)
	return items, uc.wrap(err)
}

func (uc *IdentityUsecase) BanUser(ctx context.Context, uid string) (*Account, error) {
	return uc.setUserStatus(ctx, uid, "BANNED")
}

func (uc *IdentityUsecase) PermitUser(ctx context.Context, uid string) (*Account, error) {
	return uc.setUserStatus(ctx, uid, "ACTIVE")
}

func (uc *IdentityUsecase) setUserStatus(ctx context.Context, uid, status string) (*Account, error) {
	if _, err := uc.allow(ctx, perm.UserWrite); err != nil {
		return nil, err
	}
	actor, _ := UserID(ctx)
	if actor == uid {
		return nil, merr.Client("cannot change your own status")
	}
	account, err := uc.repo.UpdateUserStatus(ctx, uid, status)
	if err != nil {
		return nil, uc.wrap(err)
	}
	cleaned := publicAccount(*account)
	return &cleaned, nil
}

func (uc *IdentityUsecase) CreateRole(ctx context.Context, name string) (*Role, error) {
	ns, err := uc.allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.CreateRole(ctx, ns, name)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) UpdateRole(ctx context.Context, uid, name string) (*Role, error) {
	ns, err := uc.allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateRole(ctx, ns, uid, name)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) UpdateRoleStatus(ctx context.Context, uid, status string) (*Role, error) {
	ns, err := uc.allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateRoleStatus(ctx, ns, uid, status)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) DeleteRole(ctx context.Context, uid string) error {
	ns, err := uc.allow(ctx, perm.RoleWrite)
	if err != nil {
		return err
	}
	return uc.wrap(uc.repo.DeleteRole(ctx, ns, uid))
}

func (uc *IdentityUsecase) GetRole(ctx context.Context, uid string) (*Role, error) {
	ns, err := uc.allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.GetRole(ctx, ns, uid)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) ListRoles(ctx context.Context, page, pageSize int32, name string) ([]Role, int, int, int, error) {
	ns, err := uc.allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := normalizePage(page, pageSize)
	items, total, err := uc.repo.ListRoles(ctx, ns, p, s, name)
	return items, total, p, s, uc.wrap(err)
}

func (uc *IdentityUsecase) SelectRoles(ctx context.Context, limit int32, lastUID string) ([]SelectItem, error) {
	ns, err := uc.allow(ctx, perm.RoleRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.repo.SelectRoles(ctx, ns, normalizeLimit(limit), lastUID)
	return items, uc.wrap(err)
}

func (uc *IdentityUsecase) UpdateRolePermissions(ctx context.Context, uid string, names []string) (*Role, error) {
	ns, err := uc.allow(ctx, perm.RoleWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.repo.UpdateRolePermissions(ctx, ns, uid, names)
	return item, uc.wrap(err)
}

func (uc *IdentityUsecase) ListPermissions(ctx context.Context) ([]Permission, error) {
	if _, err := uc.allow(ctx, perm.RoleRead); err != nil {
		return nil, err
	}
	items, err := uc.repo.ListPermissions(ctx)
	return items, uc.wrap(err)
}

func (uc *IdentityUsecase) allow(ctx context.Context, code string) (string, error) {
	if err := uc.require(ctx, code); err != nil {
		return "", err
	}
	ns, _ := NamespaceID(ctx)
	return ns, nil
}

func (uc *IdentityUsecase) require(ctx context.Context, code string) error {
	ns, err := uc.currentNamespace(ctx)
	if err != nil {
		return err
	}
	uid, ok := UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	perms, err := uc.repo.Permissions(ctx, ns, uid)
	if err != nil {
		return uc.wrap(err)
	}
	if _, ok := perms[code]; !ok {
		return merr.Forbidden("permission denied")
	}
	return nil
}

func (uc *IdentityUsecase) sameNamespace(ctx context.Context, uid string) error {
	ns, err := uc.currentNamespace(ctx)
	if err != nil {
		return err
	}
	if ns != uid {
		return merr.Client("namespace mismatch")
	}
	return nil
}

func (uc *IdentityUsecase) currentNamespace(ctx context.Context) (string, error) {
	ns, ok := NamespaceID(ctx)
	if !ok {
		return "", merr.Client("X-Namespace is required")
	}
	if _, err := uc.repo.GetNamespace(ctx, ns); err != nil {
		return "", uc.wrap(err)
	}
	return ns, nil
}

func normalizePage(page, pageSize int32) (int, int) {
	p := int(page)
	if p < 1 {
		p = 1
	}
	s := int(pageSize)
	if s < 1 {
		s = 20
	}
	if s > 100 {
		s = 100
	}
	return p, s
}

func normalizeLimit(limit int32) int {
	n := int(limit)
	if n < 1 {
		n = 20
	}
	if n > 100 {
		n = 100
	}
	return n
}

func errorsIsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}
