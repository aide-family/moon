package member

import (
	"context"
	"strings"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type Usecase struct {
	members repository.MemberRepo
	roles   repository.RoleRepo
	access  access.Checker
}

func NewUsecase(members repository.MemberRepo, roles repository.RoleRepo, checker access.Checker) *Usecase {
	return &Usecase{members: members, roles: roles, access: checker}
}

func (uc *Usecase) ListMembers(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Member, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.members.ListMembers(ctx, ns, p, s, name)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) GetMember(ctx context.Context, uid string) (*bo.Member, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.members.GetMember(ctx, ns, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListUserMembers(ctx context.Context, userUID string) ([]bo.Member, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	item, err := uc.members.MemberByUser(ctx, ns, userUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return []bo.Member{}, nil
		}
		return nil, uc.access.Wrap(err)
	}
	return []bo.Member{*item}, nil
}

func (uc *Usecase) SelectMembers(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.members.SelectMembers(ctx, ns, page.Limit(limit), lastUID)
	return items, uc.access.Wrap(err)
}

// InviteMember 把已注册用户加进当前租户。
// 邀请不能指定 owner。负责人只能通过移交产生，避免出现第二个 owner。
func (uc *Usecase) InviteMember(ctx context.Context, email, roleUID string) (*bo.Member, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	role, err := uc.roles.GetRole(ctx, ns, roleUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if role.Name == perm.RoleOwner {
		return nil, merr.Client("invalid request")
	}
	item, err := uc.members.InviteMember(ctx, ns, normalizeEmail(email), roleUID)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) DeleteMember(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.members.DeleteMember(ctx, ns, uid))
}

func (uc *Usecase) UpdateMemberStatus(ctx context.Context, uid, status string) (*bo.Member, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.members.UpdateMemberStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

// UpdateMemberRoles 替换成员的角色。
// 结果里如果包含 owner，调用者还必须持有 namespace:transfer。
// 普通成员管理员不能借分配角色把负责人交出去。
func (uc *Usecase) UpdateMemberRoles(ctx context.Context, uid string, roleUIDs []string) (*bo.Member, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	for _, roleUID := range roleUIDs {
		role, err := uc.roles.GetRole(ctx, ns, roleUID)
		if err != nil {
			return nil, uc.access.Wrap(err)
		}
		if role.Name == perm.RoleOwner {
			if err := uc.access.Require(ctx, perm.NamespaceTransfer); err != nil {
				return nil, err
			}
		}
	}
	item, err := uc.members.UpdateMemberRoles(ctx, ns, uid, roleUIDs)
	return item, uc.access.Wrap(err)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
