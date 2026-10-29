package namespace

import (
	"context"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewTenancy)

// Tenancy 管理租户的创建、状态和所有者移交。
type Tenancy struct {
	namespaces repository.NamespaceRepo
	access     access.Checker
}

func NewTenancy(namespaces repository.NamespaceRepo, checker access.Checker) *Tenancy {
	return &Tenancy{namespaces: namespaces, access: checker}
}

// CreateNamespace 用当前登录用户作为负责人。创建租户不要求已有租户上下文。
func (tenancy *Tenancy) CreateNamespace(ctx context.Context, name, remark string) (*bo.Namespace, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	item, err := tenancy.namespaces.CreateNamespace(ctx, name, remark, uid)
	return item, tenancy.access.Wrap(err)
}

func (tenancy *Tenancy) UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*bo.Namespace, error) {
	if err := tenancy.access.SameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := tenancy.access.Require(ctx, perm.NamespaceWrite); err != nil {
		return nil, err
	}
	item, err := tenancy.namespaces.UpdateNamespace(ctx, uid, name, remark, logo)
	return item, tenancy.access.Wrap(err)
}

func (tenancy *Tenancy) UpdateNamespaceStatus(ctx context.Context, uid, status string) (*bo.Namespace, error) {
	ns, err := tenancy.access.AllowAnyStatus(ctx, perm.NamespaceWrite)
	if err != nil {
		return nil, err
	}
	if ns != uid {
		return nil, merr.Client("namespace mismatch")
	}
	item, err := tenancy.namespaces.UpdateNamespaceStatus(ctx, uid, status)
	return item, tenancy.access.Wrap(err)
}

func (tenancy *Tenancy) DeleteNamespace(ctx context.Context, uid string) error {
	ns, err := tenancy.access.AllowAnyStatus(ctx, perm.NamespaceDelete)
	if err != nil {
		return err
	}
	if ns != uid {
		return merr.Client("namespace mismatch")
	}
	return tenancy.access.Wrap(tenancy.namespaces.DeleteNamespace(ctx, uid))
}

func (tenancy *Tenancy) GetNamespace(ctx context.Context, uid string) (*bo.Namespace, error) {
	ns, err := tenancy.access.AllowAnyStatus(ctx, perm.NamespaceRead)
	if err != nil {
		return nil, err
	}
	if ns != uid {
		return nil, merr.Client("namespace mismatch")
	}
	item, err := tenancy.namespaces.GetNamespace(ctx, uid)
	return item, tenancy.access.Wrap(err)
}

func (tenancy *Tenancy) ListNamespaces(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Namespace, int, int, int, error) {
	// 列表属于租户自身的查看：当前租户停用后仍要能列出来，否则页面上没有入口再启用它。
	if _, err := tenancy.access.AllowAnyStatus(ctx, perm.NamespaceRead); err != nil {
		return nil, 0, 0, 0, err
	}
	uid, _ := actor.UserID(ctx)
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := tenancy.namespaces.ListNamespaces(ctx, uid, p, s, name)
	return items, total, p, s, tenancy.access.Wrap(err)
}

func (tenancy *Tenancy) SelectNamespaces(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	if err := tenancy.access.Require(ctx, perm.NamespaceRead); err != nil {
		return nil, err
	}
	uid, _ := actor.UserID(ctx)
	items, err := tenancy.namespaces.SelectNamespaces(ctx, uid, page.Limit(limit), lastUID)
	return items, tenancy.access.Wrap(err)
}

// TransferOwner 把租户负责人换成另一名已加入成员。
// 只有 namespace:transfer 可以做这件事。原负责人会留下管理权限，目标成员成为 owner。
// 不能移交给自己。仓储在同一事务里改负责人和两边的角色，避免出现两个 owner 或没有 owner。
func (tenancy *Tenancy) TransferOwner(ctx context.Context, uid, userUID string) (*bo.Namespace, error) {
	if err := tenancy.access.SameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := tenancy.access.Require(ctx, perm.NamespaceTransfer); err != nil {
		return nil, err
	}
	item, err := tenancy.namespaces.TransferOwner(ctx, uid, userUID)
	return item, tenancy.access.Wrap(err)
}

// ListAllNamespaces 列出平台全部租户，只给系统管理员使用。
// 它走 RequireSystem，不读 X-Namespace，也不要求调用者是任何租户的成员。
func (tenancy *Tenancy) ListAllNamespaces(ctx context.Context, pageNo, pageSize int32, name, status string) ([]bo.Namespace, int, int, int, error) {
	if _, err := tenancy.access.RequireSystem(ctx, perm.SystemNamespaceRead); err != nil {
		return nil, 0, 0, 0, err
	}
	if status != "" && status != "ACTIVE" && status != "DISABLED" {
		return nil, 0, 0, 0, merr.Invalid("invalid request")
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := tenancy.namespaces.ListAllNamespaces(ctx, p, s, name, status)
	return items, total, p, s, tenancy.access.Wrap(err)
}

// ListMine 返回当前用户加入的租户，供个人资料使用，不要求已经选定租户。
func (tenancy *Tenancy) ListMine(ctx context.Context) ([]bo.Namespace, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	items, err := tenancy.namespaces.ListUserNamespaces(ctx, uid)
	return items, tenancy.access.Wrap(err)
}

// Current 返回 X-Namespace 指定的租户，不校验成员身份。
// /self/permissions 用它返回当前视角下的租户信息；没有选择租户或租户已删除时返回 nil，
// 保持该接口在租户失效时仍能返回空权限，由前端重新选择。
func (tenancy *Tenancy) Current(ctx context.Context) (*bo.Namespace, error) {
	uid, ok := actor.NamespaceID(ctx)
	if !ok {
		return nil, nil
	}
	item, err := tenancy.namespaces.GetNamespace(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, nil
		}
		return nil, tenancy.access.Wrap(err)
	}
	return item, nil
}
