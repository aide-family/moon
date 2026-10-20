package data

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/member"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/memberinvite"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/memberrole"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/role"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/Alioth/app/alioth/internal/data/ent/user"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

type inviteRepo struct{ *repoBase }

func (r *inviteRepo) UpsertInvite(ctx context.Context, in bo.MemberInvite, tokenHash string) (*bo.MemberInvite, error) {
	email := normalizeEmail(in.Email)
	userUID := ""
	if account, err := r.client.User.Query().Where(user.EmailEQ(email)).Only(ctx); err == nil {
		userUID = account.UID
	} else if !ent.IsNotFound(err) {
		return nil, r.fail(err)
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	if userUID != "" {
		exists, err := tx.Member.Query().Where(
			member.NamespaceUIDEQ(in.NamespaceUID),
			member.UserUIDEQ(userUID),
			member.DeletedAtIsNil(),
		).Exist(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		if exists {
			return nil, merr.Conflict("member already exists")
		}
	}
	// 同一租户同一邮箱已有待确认邀请时，刷新令牌、角色和有效期，不新增消息。
	existing, err := tx.MemberInvite.Query().Where(
		memberinvite.NamespaceUIDEQ(in.NamespaceUID),
		memberinvite.EmailEQ(email),
		memberinvite.StatusEQ(memberinvite.StatusPENDING),
	).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, r.fail(err)
	}
	if err == nil {
		row, err := tx.MemberInvite.UpdateOne(existing).
			SetTokenHash(tokenHash).
			SetUserUID(userUID).
			SetRoleUID(in.RoleUID).
			SetInviterUID(in.InviterUID).
			SetExpiresAt(in.ExpiresAt).
			SetStatus(memberinvite.StatusPENDING).
			ClearRespondedAt().
			Save(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		if err := createInviteMessage(ctx, tx.Client(), row, bo.MessageKindInviteReceived, ""); err != nil {
			return nil, r.fail(err)
		}
		if err := tx.Commit(); err != nil {
			return nil, r.fail(err)
		}
		return r.inviteView(ctx, row)
	}
	row, err := tx.MemberInvite.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetEmail(email).
		SetUserUID(userUID).
		SetRoleUID(in.RoleUID).
		SetInviterUID(in.InviterUID).
		SetTokenHash(tokenHash).
		SetStatus(memberinvite.StatusPENDING).
		SetExpiresAt(in.ExpiresAt).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := createInviteMessage(ctx, tx.Client(), row, bo.MessageKindInviteReceived, ""); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.inviteView(ctx, row)
}

func (r *inviteRepo) InviteByToken(ctx context.Context, tokenHash string) (*bo.MemberInvite, error) {
	row, err := r.client.MemberInvite.Query().Where(memberinvite.TokenHashEQ(tokenHash)).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if row.Status == memberinvite.StatusPENDING && row.ExpiresAt.Before(time.Now()) {
		if _, err := r.client.MemberInvite.Update().
			Where(memberinvite.UIDEQ(row.UID), memberinvite.StatusEQ(memberinvite.StatusPENDING)).
			SetStatus(memberinvite.StatusEXPIRED).
			Save(ctx); err != nil {
			return nil, r.fail(err)
		}
		row.Status = memberinvite.StatusEXPIRED
	}
	return r.inviteView(ctx, row)
}

func (r *inviteRepo) ListInvitesForUser(ctx context.Context, userUID, email string) ([]bo.MemberInvite, error) {
	r.expirePending(ctx)
	query := r.client.MemberInvite.Query().Where(
		memberinvite.StatusNEQ(memberinvite.StatusREVOKED),
	)
	if userUID != "" && email != "" {
		query = query.Where(memberinvite.Or(
			memberinvite.UserUIDEQ(userUID),
			memberinvite.EmailEQ(normalizeEmail(email)),
		))
	} else if userUID != "" {
		query = query.Where(memberinvite.UserUIDEQ(userUID))
	} else {
		query = query.Where(memberinvite.EmailEQ(normalizeEmail(email)))
	}
	rows, err := query.Order(ent.Desc(memberinvite.FieldCreatedAt)).Limit(200).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.MemberInvite, 0, len(rows))
	for _, row := range rows {
		item, err := r.inviteView(ctx, row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (r *inviteRepo) ListInvitesForNamespace(ctx context.Context, namespaceUID string) ([]bo.MemberInvite, error) {
	r.expirePending(ctx)
	rows, err := r.client.MemberInvite.Query().Where(
		memberinvite.NamespaceUIDEQ(namespaceUID),
		memberinvite.StatusEQ(memberinvite.StatusPENDING),
	).Order(ent.Desc(memberinvite.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.MemberInvite, 0, len(rows))
	for _, row := range rows {
		item, err := r.inviteView(ctx, row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (r *inviteRepo) AcceptInvite(ctx context.Context, tokenHash, inviteUID, userUID, email string) (*bo.Member, error) {
	email = normalizeEmail(email)
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	invite, err := inviteQuery(ctx, tx.Client(), tokenHash, inviteUID)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := checkInvite(invite, userUID, email); err != nil {
		return nil, err
	}
	roleRow, err := tx.Role.Query().Where(
		role.NamespaceUIDEQ(invite.NamespaceUID),
		role.UIDEQ(invite.RoleUID),
		role.DeletedAtIsNil(),
		role.StatusEQ(role.StatusACTIVE),
	).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if roleRow.Name == "owner" {
		return nil, merr.Invalid("invalid request")
	}
	memberRow, err := tx.Member.Query().Where(
		member.NamespaceUIDEQ(invite.NamespaceUID),
		member.UserUIDEQ(userUID),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, r.fail(err)
	}
	if ent.IsNotFound(err) {
		memberRow, err = tx.Member.Create().
			SetUID(id.New()).
			SetNamespaceUID(invite.NamespaceUID).
			SetUserUID(userUID).
			SetStatus(member.StatusJOINED).
			Save(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
	} else if memberRow.Status != member.StatusJOINED {
		memberRow, err = tx.Member.UpdateOne(memberRow).SetStatus(member.StatusJOINED).Save(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
	}
	if _, err := tx.MemberRole.Delete().Where(
		memberrole.NamespaceUIDEQ(invite.NamespaceUID),
		memberrole.MemberUIDEQ(memberRow.UID),
	).Exec(ctx); err != nil {
		return nil, r.fail(err)
	}
	if _, err := tx.MemberRole.Create().
		SetUID(id.New()).
		SetNamespaceUID(invite.NamespaceUID).
		SetMemberUID(memberRow.UID).
		SetRoleUID(roleRow.UID).
		Save(ctx); err != nil {
		return nil, r.fail(err)
	}
	updated, err := tx.MemberInvite.UpdateOne(invite).
		SetUserUID(userUID).
		SetStatus(memberinvite.StatusACCEPTED).
		SetRespondedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := createInviteMessage(ctx, tx.Client(), updated, bo.MessageKindInviteAccepted, userUID); err != nil {
		return nil, r.fail(err)
	}
	if err := markInviteMessageRead(ctx, tx.Client(), invite.UID); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return (&memberRepo{r.repoBase}).GetMember(ctx, invite.NamespaceUID, memberRow.UID)
}

func (r *inviteRepo) DeclineInvite(ctx context.Context, tokenHash, inviteUID, userUID, email string) error {
	email = normalizeEmail(email)
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	invite, err := inviteQuery(ctx, tx.Client(), tokenHash, inviteUID)
	if err != nil {
		return r.fail(err)
	}
	if err := checkInvite(invite, userUID, email); err != nil {
		return err
	}
	n, err := tx.MemberInvite.Update().
		Where(memberinvite.UIDEQ(invite.UID), memberinvite.StatusEQ(memberinvite.StatusPENDING)).
		SetUserUID(userUID).
		SetStatus(memberinvite.StatusDECLINED).
		SetRespondedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	invite.Status = memberinvite.StatusDECLINED
	invite.UserUID = userUID
	if err := createInviteMessage(ctx, tx.Client(), invite, bo.MessageKindInviteDeclined, userUID); err != nil {
		return r.fail(err)
	}
	if err := markInviteMessageRead(ctx, tx.Client(), invite.UID); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *inviteRepo) RevokeInvite(ctx context.Context, namespaceUID, uid string) error {
	n, err := r.client.MemberInvite.Update().
		Where(
			memberinvite.NamespaceUIDEQ(namespaceUID),
			memberinvite.UIDEQ(uid),
			memberinvite.StatusEQ(memberinvite.StatusPENDING),
		).
		SetStatus(memberinvite.StatusREVOKED).
		SetRespondedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return r.fail(err)
	}
	if n == 0 {
		return merr.NotFound("not found")
	}
	return nil
}

// expirePending 把已过期的待确认邀请标成 EXPIRED，列表里不再显示成待处理。
func (r *inviteRepo) expirePending(ctx context.Context) {
	if _, err := r.client.MemberInvite.Update().
		Where(
			memberinvite.StatusEQ(memberinvite.StatusPENDING),
			memberinvite.ExpiresAtLT(time.Now()),
		).
		SetStatus(memberinvite.StatusEXPIRED).
		Save(ctx); err != nil {
		r.log.Error("expire member invites: " + err.Error())
	}
}

// inviteQuery 按邮件令牌或平台消息标识取邀请，二者至少给一个。
func inviteQuery(ctx context.Context, client *ent.Client, tokenHash, inviteUID string) (*ent.MemberInvite, error) {
	switch {
	case tokenHash != "":
		return client.MemberInvite.Query().Where(memberinvite.TokenHashEQ(tokenHash)).Only(ctx)
	case inviteUID != "":
		return client.MemberInvite.Query().Where(memberinvite.UIDEQ(inviteUID)).Only(ctx)
	default:
		return nil, merr.Client("invalid invitation")
	}
}

// checkInvite 校验邀请仍然有效，并且确实发给当前登录用户。
func checkInvite(invite *ent.MemberInvite, userUID, email string) error {
	if invite.Status != memberinvite.StatusPENDING {
		return merr.Invalid("invitation is not pending")
	}
	if invite.ExpiresAt.Before(time.Now()) {
		return merr.Invalid("invitation expired")
	}
	if invite.UserUID != "" {
		if invite.UserUID != userUID {
			return merr.Forbidden("invitation belongs to another user")
		}
	} else if normalizeEmail(invite.Email) != email {
		return merr.Forbidden("invitation email does not match")
	}
	return nil
}

func (r *inviteRepo) inviteView(ctx context.Context, row *ent.MemberInvite) (*bo.MemberInvite, error) {
	item := bo.MemberInvite{
		UID:          row.UID,
		NamespaceUID: row.NamespaceUID,
		Email:        row.Email,
		UserUID:      row.UserUID,
		RoleUID:      row.RoleUID,
		InviterUID:   row.InviterUID,
		Status:       string(row.Status),
		ExpiresAt:    row.ExpiresAt,
		CreatedAt:    row.CreatedAt,
		RespondedAt:  row.RespondedAt,
	}
	if ns, err := r.client.Namespace.Query().Where(namespace.UIDEQ(row.NamespaceUID)).Only(ctx); err == nil {
		item.NamespaceName = ns.Name
	}
	if roleRow, err := r.client.Role.Query().Where(
		role.UIDEQ(row.RoleUID),
		role.NamespaceUIDEQ(row.NamespaceUID),
	).Only(ctx); err == nil {
		item.RoleName = roleRow.Name
	}
	if inviter, err := r.client.User.Query().Where(user.UIDEQ(row.InviterUID)).Only(ctx); err == nil {
		item.InviterName = inviter.Name
		if item.InviterName == "" {
			item.InviterName = inviter.Email
		}
	}
	if item.InviterName == "" {
		item.InviterName = "Alioth"
	}
	return &item, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// createInviteMessage 写入邀请相关的系统消息。
// 收到的邀请发给被邀请人；同意/拒绝的结果发给邀请人。
func createInviteMessage(ctx context.Context, client *ent.Client, invite *ent.MemberInvite, kind, actorUID string) error {
	exists, err := client.SystemMessage.Query().Where(
		systemmessage.InviteUIDEQ(invite.UID),
		systemmessage.KindEQ(systemmessage.Kind(kind)),
	).Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	namespaceName := ""
	if ns, err := client.Namespace.Query().Where(namespace.UIDEQ(invite.NamespaceUID)).Only(ctx); err == nil {
		namespaceName = ns.Name
	}
	roleName := ""
	if roleRow, err := client.Role.Query().Where(role.UIDEQ(invite.RoleUID), role.NamespaceUIDEQ(invite.NamespaceUID)).Only(ctx); err == nil {
		roleName = roleRow.Name
	}
	recipientUID := invite.UserUID
	recipientEmail := normalizeEmail(invite.Email)
	actor := actorUID
	if kind == bo.MessageKindInviteReceived {
		recipientUID = invite.UserUID
		recipientEmail = normalizeEmail(invite.Email)
		actor = invite.InviterUID
	} else {
		recipientUID = invite.InviterUID
		recipientEmail = ""
		actor = actorUID
	}
	if recipientUID == "" && recipientEmail == "" {
		return nil
	}
	actorName := ""
	if actor != "" {
		if row, err := client.User.Query().Where(user.UIDEQ(actor)).Only(ctx); err == nil {
			actorName = row.Name
			if actorName == "" {
				actorName = row.Email
			}
		}
	}
	if actorName == "" {
		actorName = "Alioth"
	}
	if recipientEmail == "" && recipientUID != "" {
		if row, err := client.User.Query().Where(user.UIDEQ(recipientUID)).Only(ctx); err == nil {
			recipientEmail = normalizeEmail(row.Email)
		}
	}
	if _, err := client.SystemMessage.Create().
		SetUID(id.New()).
		SetUserUID(recipientUID).
		SetEmail(recipientEmail).
		SetKind(systemmessage.Kind(kind)).
		SetActorName(actorName).
		SetNamespaceUID(invite.NamespaceUID).
		SetNamespaceName(namespaceName).
		SetRoleName(roleName).
		SetInviteUID(invite.UID).
		Save(ctx); err != nil {
		return err
	}
	return nil
}

func markInviteMessageRead(ctx context.Context, client *ent.Client, inviteUID string) error {
	_, err := client.SystemMessage.Update().
		Where(
			systemmessage.InviteUIDEQ(inviteUID),
			systemmessage.KindEQ(systemmessage.KindINVITE_RECEIVED),
			systemmessage.ReadAtIsNil(),
		).
		SetReadAt(time.Now()).
		Save(ctx)
	return err
}
