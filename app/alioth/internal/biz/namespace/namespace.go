package namespace

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/internal/mailer"
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
	features   repository.FeatureRepo
	access     access.Checker
	mail       *mailer.Sender
	log        *slog.Logger
}

func NewTenancy(namespaces repository.NamespaceRepo, features repository.FeatureRepo, checker access.Checker, mail *mailer.Sender, logger *slog.Logger) *Tenancy {
	return &Tenancy{namespaces: namespaces, features: features, access: checker, mail: mail, log: logger}
}

// CreateNamespace 用当前登录用户作为负责人，并绑定选择的功能套餐。
// 创建租户不要求已有租户上下文；只有公开且启用的套餐可以被用户选择，
// 内测套餐只能由平台在租户功能授权里分配。
func (tenancy *Tenancy) CreateNamespace(ctx context.Context, name, remark, packageUID string) (*bo.Namespace, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	if err := tenancy.checkPublicPackage(ctx, packageUID); err != nil {
		return nil, err
	}
	item, err := tenancy.namespaces.CreateNamespace(ctx, name, remark, uid, packageUID)
	return item, tenancy.access.Wrap(err)
}

// checkPublicPackage 确认套餐仍公开可售：停用、删除或内测套餐都会拒绝创建/下单。
func (tenancy *Tenancy) checkPublicPackage(ctx context.Context, packageUID string) error {
	pkg, err := tenancy.features.GetPackage(ctx, packageUID)
	if err != nil {
		return tenancy.access.Wrap(err)
	}
	if pkg.Status != "ACTIVE" || pkg.Visibility != perm.PackageVisibilityPublic {
		return merr.Client("package is not available")
	}
	return nil
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

// UpdateNamespaceStatus 租户侧不能再自行启停：停用和启用由平台管理，
// 否则租户可以用自己的令牌直接解封。平台入口是 SetSystemNamespaceStatus。
func (tenancy *Tenancy) UpdateNamespaceStatus(_ context.Context, _ string, _ string) (*bo.Namespace, error) {
	return nil, merr.Forbidden("namespace status is managed by platform")
}

// SetSystemNamespaceStatus 由平台停用或启用租户，不读 X-Namespace。
// 停用必须填写原因（1-500 字），成功后邮件通知租户负责人；邮件失败只记日志。
// 启用会清空停用原因且不发邮件。
func (tenancy *Tenancy) SetSystemNamespaceStatus(ctx context.Context, uid, status, reason string) (*bo.Namespace, error) {
	if _, err := tenancy.access.RequireSystem(ctx, perm.SystemNamespaceWrite); err != nil {
		return nil, err
	}
	reason = strings.TrimSpace(reason)
	switch status {
	case "DISABLED":
		if reason == "" {
			return nil, merr.Invalid("reason is required")
		}
		if len([]rune(reason)) > 500 {
			return nil, merr.Invalid("reason is too long")
		}
	case "ACTIVE":
		// 启用后租户不再有停用原因，避免下一次停用时展示旧原因。
		reason = ""
	default:
		return nil, merr.Invalid("invalid status")
	}
	current, err := tenancy.namespaces.GetNamespace(ctx, uid)
	if err != nil {
		return nil, tenancy.access.Wrap(err)
	}
	item, err := tenancy.namespaces.SetNamespaceStatus(ctx, uid, status, reason)
	if err != nil {
		return nil, tenancy.access.Wrap(err)
	}
	// 只在真正从启用切到停用时通知，重复停用不重复发信。
	if status == "DISABLED" && current.Status != "DISABLED" && item != nil {
		tenancy.notifyNamespaceDisabled(current, item, reason)
	}
	return item, nil
}

// ArchiveNamespace 归档租户：软删除租户和它的全部业务数据。不读 X-Namespace。
// 归档不可恢复，没有取消归档的接口。
func (tenancy *Tenancy) ArchiveNamespace(ctx context.Context, uid string) error {
	if _, err := tenancy.access.RequireSystem(ctx, perm.SystemNamespaceWrite); err != nil {
		return err
	}
	return tenancy.access.Wrap(tenancy.features.ArchiveNamespace(ctx, uid))
}

// notifyNamespaceDisabled 通知租户负责人租户已停用。邮件是旁路动作：
// SMTP 未配置、负责人没有邮箱或发送失败都只记日志，不影响停用结果。
func (tenancy *Tenancy) notifyNamespaceDisabled(before, after *bo.Namespace, reason string) {
	name := before.Name
	if name == "" {
		name = after.Name
	}
	subject := "Alioth 租户已停用：" + name
	body := fmt.Sprintf("您好，\n\n租户「%s」已被平台停用。\n\n停用原因：%s\n停用时间：%s\n\n如需恢复，请联系平台管理员。\n",
		name, reason, time.Now().Format("2006-01-02 15:04:05"))
	recipient := strings.TrimSpace(before.OwnerEmail)
	if tenancy.mail == nil || !tenancy.mail.Configured() {
		tenancy.log.Warn("namespace disabled: smtp is not configured, skip notifying owner",
			"namespace", after.UID, "owner", before.OwnerUID)
		return
	}
	if recipient == "" {
		tenancy.log.Warn("namespace disabled: owner has no email, skip notifying",
			"namespace", after.UID, "owner", before.OwnerUID)
		return
	}
	if err := tenancy.mail.Send(recipient, subject, body); err != nil {
		tenancy.log.Error(fmt.Sprintf("namespace disabled: notify %s failed: %v", recipient, err),
			"namespace", after.UID)
	}
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
// 只有 namespace:transfer 可以做这件事。负责人身份由 namespace.owner_uid 决定，
// 权限隐式生效，因此这里只改该字段，不再增删 owner / admin 角色授权。
// 不能移交给自己。仓储在事务里校验目标成员是 JOINED 成员后再改负责人。
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

// ListMine 返回当前用户加入的租户，供切换器使用，不要求已经选定租户。
// includeDisabled 只用于切换失败时解释原因；正常切换列表不包含停用租户。
func (tenancy *Tenancy) ListMine(ctx context.Context, includeDisabled bool) ([]bo.Namespace, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	var (
		items []bo.Namespace
		err   error
	)
	if includeDisabled {
		items, err = tenancy.namespaces.ListUserNamespacesIncludingDisabled(ctx, uid)
	} else {
		items, err = tenancy.namespaces.ListUserNamespaces(ctx, uid)
	}
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
