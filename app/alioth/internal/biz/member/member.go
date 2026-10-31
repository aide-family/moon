package member

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/wire"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/internal/mailer"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/page"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

var ProviderSet = wire.NewSet(NewMembership)

const inviteTTL = 7 * 24 * time.Hour

// Membership 管理租户成员和平台邀请消息。
// 邀请不会直接产生成员：必须由被邀请人在平台点击同意。
type Membership struct {
	members  repository.MemberRepo
	invites  repository.InviteRepo
	users    repository.UserRepo
	roles    repository.RoleRepo
	features repository.FeatureRepo
	access   access.Checker
	mail     *mailer.Sender
	webURL   string
	log      *slog.Logger
}

func NewMembership(members repository.MemberRepo, invites repository.InviteRepo, users repository.UserRepo, roles repository.RoleRepo, features repository.FeatureRepo, checker access.Checker, mail *mailer.Sender, auth *conf.Auth, logger *slog.Logger) *Membership {
	webURL := "http://127.0.0.1:8080"
	if auth != nil && strings.TrimSpace(auth.GetWebUrl()) != "" {
		webURL = strings.TrimRight(strings.TrimSpace(auth.GetWebUrl()), "/")
	}
	return &Membership{
		members:  members,
		invites:  invites,
		users:    users,
		roles:    roles,
		features: features,
		access:   checker,
		mail:     mail,
		webURL:   webURL,
		log:      logger,
	}
}

func (membership *Membership) ListMembers(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Member, int, int, int, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := membership.members.ListMembers(ctx, ns, p, s, name)
	return items, total, p, s, membership.access.Wrap(err)
}

func (membership *Membership) GetMember(ctx context.Context, uid string) (*bo.Member, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	item, err := membership.members.GetMember(ctx, ns, uid)
	return item, membership.access.Wrap(err)
}

func (membership *Membership) ListUserMembers(ctx context.Context, userUID string) ([]bo.Member, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	item, err := membership.members.MemberByUser(ctx, ns, userUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return []bo.Member{}, nil
		}
		return nil, membership.access.Wrap(err)
	}
	return []bo.Member{*item}, nil
}

func (membership *Membership) SelectMembers(ctx context.Context, limit int32, lastUID string) ([]bo.SelectItem, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	items, err := membership.members.SelectMembers(ctx, ns, page.Limit(limit), lastUID)
	return items, membership.access.Wrap(err)
}

// InviteMember 创建一条邀请消息，并尝试发送邀请邮件。
// 邮件失败不回滚邀请：用户仍然可以在平台消息里看到并确认。
// 非负责人只能邀请持有「权限集合 ⊆ 自己权限」的角色，避免借邀请越权。
func (membership *Membership) InviteMember(ctx context.Context, email, roleUID string) (*bo.MemberInvite, bool, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, false, err
	}
	role, err := membership.roles.GetRole(ctx, ns, roleUID)
	if err != nil {
		return nil, false, membership.access.Wrap(err)
	}
	if role.Status != "ACTIVE" {
		return nil, false, merr.Client("invalid request")
	}
	if err := membership.ensureRolesGrantable(ctx, []bo.Role{*role}); err != nil {
		return nil, false, err
	}
	// 邀请前先按当前套餐容量拦住明显超限的操作，给出比「确认时才发现」更早的提示。
	// 真正防止并发放大的检查在确认邀请的事务里，以那里的结果为准。
	if err := membership.checkMemberLimit(ctx, ns); err != nil {
		return nil, false, err
	}
	inviterUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, false, merr.Unauthenticated("missing token")
	}
	token, err := inviteToken()
	if err != nil {
		return nil, false, merr.Server("request failed")
	}
	item, err := membership.invites.UpsertInvite(ctx, bo.MemberInvite{
		NamespaceUID: ns,
		Email:        normalizeEmail(email),
		RoleUID:      role.UID,
		InviterUID:   inviterUID,
		ExpiresAt:    time.Now().Add(inviteTTL),
	}, hashToken(token))
	if err != nil {
		return nil, false, membership.access.Wrap(err)
	}
	sent := membership.sendInviteMail(item, token)
	return item, sent, nil
}

// checkMemberLimit 校验租户当前 JOINED 成员数没有达到套餐上限。
// 上限 0 表示不限制；邀请中的成员不占名额，只有确认后成为 JOINED 成员才计入。
func (membership *Membership) checkMemberLimit(ctx context.Context, namespaceUID string) error {
	limit, err := membership.features.NamespaceMemberLimit(ctx, namespaceUID)
	if err != nil {
		return membership.access.Wrap(err)
	}
	if limit == 0 {
		return nil
	}
	joined, err := membership.members.CountJoined(ctx, namespaceUID)
	if err != nil {
		return membership.access.Wrap(err)
	}
	if int32(joined) >= limit {
		return merr.Client("member limit reached")
	}
	return nil
}

// ListInvites 返回当前租户还在等待确认的邀请，供成员页展示和撤销。
func (membership *Membership) ListInvites(ctx context.Context) ([]bo.MemberInvite, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	items, err := membership.invites.ListInvitesForNamespace(ctx, ns)
	return items, membership.access.Wrap(err)
}

func (membership *Membership) RevokeInvite(ctx context.Context, uid string) error {
	ns, err := membership.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return err
	}
	return membership.access.Wrap(membership.invites.RevokeInvite(ctx, ns, uid))
}

// ListSelfInvites 是平台消息收件箱。未登录不能看，未注册邮箱在注册后按邮箱匹配。
func (membership *Membership) ListSelfInvites(ctx context.Context) ([]bo.MemberInvite, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := membership.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, membership.access.Wrap(err)
	}
	items, err := membership.invites.ListInvitesForUser(ctx, uid, account.Email)
	return items, membership.access.Wrap(err)
}

// GetInviteByToken 是邮件快捷链接的公开入口。没有令牌的人拿不到邀请。
func (membership *Membership) GetInviteByToken(ctx context.Context, token string) (*bo.MemberInvite, error) {
	if strings.TrimSpace(token) == "" {
		return nil, merr.Client("invalid invitation")
	}
	item, err := membership.invites.InviteByToken(ctx, hashToken(token))
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.NotFound("invitation not found")
		}
		return nil, membership.access.Wrap(err)
	}
	return item, nil
}

// AcceptInvite 把当前登录用户变成租户成员。只有受邀邮箱本人可以确认。
func (membership *Membership) AcceptInvite(ctx context.Context, token, inviteUID string) (*bo.Member, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := membership.users.UserByUID(ctx, userUID)
	if err != nil {
		return nil, membership.access.Wrap(err)
	}
	item, err := membership.invites.AcceptInvite(ctx, optionalTokenHash(token), inviteUID, userUID, normalizeEmail(account.Email))
	return item, membership.access.Wrap(err)
}

func (membership *Membership) DeclineInvite(ctx context.Context, token, inviteUID string) error {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	account, err := membership.users.UserByUID(ctx, uid)
	if err != nil {
		return membership.access.Wrap(err)
	}
	return membership.access.Wrap(membership.invites.DeclineInvite(ctx, optionalTokenHash(token), inviteUID, uid, normalizeEmail(account.Email)))
}

func (membership *Membership) DeleteMember(ctx context.Context, uid string) error {
	ns, err := membership.access.Allow(ctx, perm.MemberDelete)
	if err != nil {
		return err
	}
	actorUID, _ := actor.UserID(ctx)
	return membership.access.Wrap(membership.members.DeleteMember(ctx, ns, uid, actorUID))
}

// UpdateMemberStatus 不允许管理员直接把 INVITED 改成 JOINED，成员关系只能由本人确认产生。
func (membership *Membership) UpdateMemberStatus(ctx context.Context, uid, status string) (*bo.Member, error) {
	if status == "JOINED" {
		return nil, merr.Invalid("member must accept invitation")
	}
	ns, err := membership.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	item, err := membership.members.UpdateMemberStatus(ctx, ns, uid, status)
	return item, membership.access.Wrap(err)
}

// UpdateMemberRoles 替换成员的角色。
// 非负责人只能分配「权限集合 ⊆ 自己权限」的角色；负责人不受子集限制。
// 负责人身份只由 namespace.owner_uid 决定，不再依赖名为 owner 的角色。
func (membership *Membership) UpdateMemberRoles(ctx context.Context, uid string, roleUIDs []string) (*bo.Member, error) {
	ns, err := membership.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	roles := make([]bo.Role, 0, len(roleUIDs))
	for _, roleUID := range roleUIDs {
		role, err := membership.roles.GetRole(ctx, ns, roleUID)
		if err != nil {
			return nil, membership.access.Wrap(err)
		}
		roles = append(roles, *role)
	}
	if err := membership.ensureRolesGrantable(ctx, roles); err != nil {
		return nil, err
	}
	item, err := membership.members.UpdateMemberRoles(ctx, ns, uid, roleUIDs)
	return item, membership.access.Wrap(err)
}

// ensureRolesGrantable 校验调用者能否把这些角色分配给成员：
// 负责人拥有租户全部权限，不受子集限制；非负责人只能分配权限完全包含在自己
// 当前权限里的角色。角色本身的权限范围由角色授权接口按租户能力限制。
func (membership *Membership) ensureRolesGrantable(ctx context.Context, roles []bo.Role) error {
	held, owner, err := membership.access.HeldPermissions(ctx)
	if err != nil {
		return err
	}
	if owner {
		return nil
	}
	for _, role := range roles {
		for _, name := range role.Permissions {
			if _, ok := held[name]; !ok {
				return merr.Client("cannot grant permissions beyond your own")
			}
		}
	}
	return nil
}

func (membership *Membership) sendInviteMail(invite *bo.MemberInvite, token string) bool {
	if membership.mail == nil {
		return false
	}
	link := fmt.Sprintf("%s/invite/%s", membership.webURL, token)
	subject := fmt.Sprintf("邀请你加入 %s", invite.NamespaceName)
	body := fmt.Sprintf(
		"%s 邀请你加入租户「%s」，角色：%s。\n\n请在 7 天内打开以下链接确认：\n%s\n\n如果还没有注册，链接会引导你完成注册；注册成功后会自动登录并打开邀请函。\n",
		invite.InviterName, invite.NamespaceName, invite.RoleName, link,
	)
	if err := membership.mail.Send(invite.Email, subject, body); err != nil {
		membership.log.Error(fmt.Sprintf("send invite mail: %v", err))
		return false
	}
	return true
}

func inviteToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

// optionalTokenHash 让平台消息用 uid 确认时不会退化成“空令牌的摘要”。
func optionalTokenHash(token string) string {
	if strings.TrimSpace(token) == "" {
		return ""
	}
	return hashToken(token)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
