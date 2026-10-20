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

	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/internal/mailer"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/app/alioth/pkg/perm"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

const inviteTTL = 7 * 24 * time.Hour

// Usecase 管理租户成员和平台邀请消息。
// 邀请不会直接产生成员：必须由被邀请人在平台点击同意。
type Usecase struct {
	members repository.MemberRepo
	invites repository.InviteRepo
	users   repository.UserRepo
	roles   repository.RoleRepo
	access  access.Checker
	mail    *mailer.Sender
	webURL  string
	log     *slog.Logger
}

func NewUsecase(members repository.MemberRepo, invites repository.InviteRepo, users repository.UserRepo, roles repository.RoleRepo, checker access.Checker, mail *mailer.Sender, auth *conf.Auth, logger *slog.Logger) *Usecase {
	webURL := "http://127.0.0.1:8080"
	if auth != nil && strings.TrimSpace(auth.GetWebUrl()) != "" {
		webURL = strings.TrimRight(strings.TrimSpace(auth.GetWebUrl()), "/")
	}
	return &Usecase{
		members: members,
		invites: invites,
		users:   users,
		roles:   roles,
		access:  checker,
		mail:    mail,
		webURL:  webURL,
		log:     logger,
	}
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

// InviteMember 创建一条邀请消息，并尝试发送邀请邮件。
// 邮件失败不回滚邀请：用户仍然可以在平台消息里看到并确认。
func (uc *Usecase) InviteMember(ctx context.Context, email, roleUID string) (*bo.MemberInvite, bool, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, false, err
	}
	role, err := uc.roles.GetRole(ctx, ns, roleUID)
	if err != nil {
		return nil, false, uc.access.Wrap(err)
	}
	if role.Name == perm.RoleOwner || role.Status != "ACTIVE" {
		return nil, false, merr.Client("invalid request")
	}
	inviterUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, false, merr.Unauthenticated("missing token")
	}
	token, err := inviteToken()
	if err != nil {
		return nil, false, merr.Server("request failed")
	}
	item, err := uc.invites.UpsertInvite(ctx, bo.MemberInvite{
		NamespaceUID: ns,
		Email:        normalizeEmail(email),
		RoleUID:      role.UID,
		InviterUID:   inviterUID,
		ExpiresAt:    time.Now().Add(inviteTTL),
	}, hashToken(token))
	if err != nil {
		return nil, false, uc.access.Wrap(err)
	}
	sent := uc.sendInviteMail(item, token)
	return item, sent, nil
}

// ListInvites 返回当前租户还在等待确认的邀请，供成员页展示和撤销。
func (uc *Usecase) ListInvites(ctx context.Context) ([]bo.MemberInvite, error) {
	ns, err := uc.access.Allow(ctx, perm.MemberRead)
	if err != nil {
		return nil, err
	}
	items, err := uc.invites.ListInvitesForNamespace(ctx, ns)
	return items, uc.access.Wrap(err)
}

func (uc *Usecase) RevokeInvite(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return err
	}
	return uc.access.Wrap(uc.invites.RevokeInvite(ctx, ns, uid))
}

// ListSelfInvites 是平台消息收件箱。未登录不能看，未注册邮箱在注册后按邮箱匹配。
func (uc *Usecase) ListSelfInvites(ctx context.Context) ([]bo.MemberInvite, error) {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := uc.users.UserByUID(ctx, uid)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	items, err := uc.invites.ListInvitesForUser(ctx, uid, account.Email)
	return items, uc.access.Wrap(err)
}

// GetInviteByToken 是邮件快捷链接的公开入口。没有令牌的人拿不到邀请。
func (uc *Usecase) GetInviteByToken(ctx context.Context, token string) (*bo.MemberInvite, error) {
	if strings.TrimSpace(token) == "" {
		return nil, merr.Client("invalid invitation")
	}
	item, err := uc.invites.InviteByToken(ctx, hashToken(token))
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.NotFound("invitation not found")
		}
		return nil, uc.access.Wrap(err)
	}
	return item, nil
}

// AcceptInvite 把当前登录用户变成租户成员。只有受邀邮箱本人可以确认。
func (uc *Usecase) AcceptInvite(ctx context.Context, token, inviteUID string) (*bo.Member, error) {
	userUID, ok := actor.UserID(ctx)
	if !ok {
		return nil, merr.Unauthenticated("missing token")
	}
	account, err := uc.users.UserByUID(ctx, userUID)
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	item, err := uc.invites.AcceptInvite(ctx, optionalTokenHash(token), inviteUID, userUID, normalizeEmail(account.Email))
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) DeclineInvite(ctx context.Context, token, inviteUID string) error {
	uid, ok := actor.UserID(ctx)
	if !ok {
		return merr.Unauthenticated("missing token")
	}
	account, err := uc.users.UserByUID(ctx, uid)
	if err != nil {
		return uc.access.Wrap(err)
	}
	return uc.access.Wrap(uc.invites.DeclineInvite(ctx, optionalTokenHash(token), inviteUID, uid, normalizeEmail(account.Email)))
}

func (uc *Usecase) DeleteMember(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return err
	}
	actorUID, _ := actor.UserID(ctx)
	return uc.access.Wrap(uc.members.DeleteMember(ctx, ns, uid, actorUID))
}

// UpdateMemberStatus 不允许管理员直接把 INVITED 改成 JOINED，成员关系只能由本人确认产生。
func (uc *Usecase) UpdateMemberStatus(ctx context.Context, uid, status string) (*bo.Member, error) {
	if status == "JOINED" {
		return nil, merr.Invalid("member must accept invitation")
	}
	ns, err := uc.access.Allow(ctx, perm.MemberWrite)
	if err != nil {
		return nil, err
	}
	item, err := uc.members.UpdateMemberStatus(ctx, ns, uid, status)
	return item, uc.access.Wrap(err)
}

// UpdateMemberRoles 替换成员的角色。
// 结果里如果包含 owner，调用者还必须持有 namespace:transfer。
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

func (uc *Usecase) sendInviteMail(invite *bo.MemberInvite, token string) bool {
	if uc.mail == nil {
		return false
	}
	link := fmt.Sprintf("%s/invite/%s", uc.webURL, token)
	subject := fmt.Sprintf("邀请你加入 %s", invite.NamespaceName)
	body := fmt.Sprintf(
		"%s 邀请你加入租户「%s」，角色：%s。\n\n请在 7 天内打开以下链接确认：\n%s\n\n如果还没有注册，链接会引导你完成注册；注册成功后会自动登录并打开邀请函。\n",
		invite.InviterName, invite.NamespaceName, invite.RoleName, link,
	)
	if err := uc.mail.Send(invite.Email, subject, body); err != nil {
		uc.log.Error(fmt.Sprintf("send invite mail: %v", err))
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
