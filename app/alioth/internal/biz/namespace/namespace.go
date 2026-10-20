package namespace

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
	namespaces repository.NamespaceRepo
	access     access.Checker
}

func NewUsecase(namespaces repository.NamespaceRepo, checker access.Checker) *Usecase {
	return &Usecase{namespaces: namespaces, access: checker}
}

// CreateNamespace 用当前登录用户作为负责人。创建租户不要求已有租户上下文。
func (uc *Usecase) CreateNamespace(ctx context.Context, name, remark string) (*bo.Namespace, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	item, err := uc.namespaces.CreateNamespace(ctx, name, remark, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*bo.Namespace, error) {
	if err := uc.access.SameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := uc.access.Require(ctx, perm.NamespaceWrite); err != nil {
		return nil, err
	}
	item, err := uc.namespaces.UpdateNamespace(ctx, uid, name, remark, logo)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) UpdateNamespaceStatus(ctx context.Context, uid, status string) (*bo.Namespace, error) {
	ns, err := uc.access.AllowAnyStatus(ctx, perm.NamespaceWrite)
	if err != nil {
		return nil, err
	}
	if ns != uid {
		return nil, merr.Client("namespace mismatch")
	}
	item, err := uc.namespaces.UpdateNamespaceStatus(ctx, uid, status)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) DeleteNamespace(ctx context.Context, uid string) error {
	ns, err := uc.access.AllowAnyStatus(ctx, perm.NamespaceDelete)
	if err != nil {
		return err
	}
	if ns != uid {
		return merr.Client("namespace mismatch")
	}
	return uc.access.Wrap(uc.namespaces.DeleteNamespace(ctx, uid))
}

func (uc *Usecase) GetNamespace(ctx context.Context, uid string) (*bo.Namespace, error) {
	ns, err := uc.access.AllowAnyStatus(ctx, perm.NamespaceRead)
	if err != nil {
		return nil, err
	}
	if ns != uid {
		return nil, merr.Client("namespace mismatch")
	}
	item, err := uc.namespaces.GetNamespace(ctx, uid)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) ListNamespaces(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Namespace, int, int, int, error) {
	if err := uc.access.Require(ctx, perm.NamespaceRead); err != nil {
		return nil, 0, 0, 0, err
	}
	uid, _ := actor.UserID(ctx)
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.namespaces.ListNamespaces(ctx, uid, p, s, name)
	return items, total, p, s, uc.access.Wrap(err)
}

func (uc *Usecase) SelectNamespaces(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	if err := uc.access.Require(ctx, perm.NamespaceRead); err != nil {
		return nil, err
	}
	uid, _ := actor.UserID(ctx)
	items, err := uc.namespaces.SelectNamespaces(ctx, uid, page.Limit(limit), lastUID)
	return items, uc.access.Wrap(err)
}

// TransferOwner 把租户负责人换成另一名已加入成员。
// 只有 namespace:transfer 可以做这件事。原负责人会留下管理权限，目标成员成为 owner。
// 不能移交给自己。仓储在同一事务里改负责人和两边的角色，避免出现两个 owner 或没有 owner。
func (uc *Usecase) TransferOwner(ctx context.Context, uid, userUID string) (*bo.Namespace, error) {
	if err := uc.access.SameNamespace(ctx, uid); err != nil {
		return nil, err
	}
	if err := uc.access.Require(ctx, perm.NamespaceTransfer); err != nil {
		return nil, err
	}
	item, err := uc.namespaces.TransferOwner(ctx, uid, userUID)
	return item, uc.access.Wrap(err)
}

// ListMine 返回当前用户加入的租户，供个人资料使用，不要求已经选定租户。
func (uc *Usecase) ListMine(ctx context.Context) ([]bo.Namespace, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	items, err := uc.namespaces.ListUserNamespaces(ctx, uid)
	return items, uc.access.Wrap(err)
}
