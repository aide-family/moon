package data

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/predicate"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/systemmessage"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *messageRepo) ListMessages(ctx context.Context, userUID, email string, page, pageSize int, unreadOnly bool) ([]bo.SystemMessage, int, error) {
	preds := messagePredicates(userUID, email)
	if len(preds) == 0 {
		return []bo.SystemMessage{}, 0, nil
	}
	if unreadOnly {
		preds = append(preds, systemmessage.ReadAtIsNil())
	}
	query := r.client.SystemMessage.Query().Where(preds...)
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(systemmessage.FieldCreatedAt)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	items := make([]bo.SystemMessage, 0, len(rows))
	for _, row := range rows {
		items = append(items, toSystemMessage(row))
	}
	return items, total, nil
}

func (r *messageRepo) UnreadCount(ctx context.Context, userUID, email string) (int, error) {
	preds := messagePredicates(userUID, email)
	if len(preds) == 0 {
		return 0, nil
	}
	preds = append(preds, systemmessage.ReadAtIsNil())
	count, err := r.client.SystemMessage.Query().Where(preds...).Count(ctx)
	if err != nil {
		return 0, r.fail(err)
	}
	return count, nil
}

func (r *messageRepo) MarkRead(ctx context.Context, userUID, email, uid string) error {
	preds := messagePredicates(userUID, email)
	if len(preds) == 0 {
		return merr.NotFound("not found")
	}
	exists, err := r.client.SystemMessage.Query().Where(append(preds, systemmessage.UIDEQ(uid))...).Exist(ctx)
	if err != nil {
		return r.fail(err)
	}
	if !exists {
		return merr.NotFound("not found")
	}
	if _, err := r.client.SystemMessage.Update().
		Where(append(preds, systemmessage.UIDEQ(uid), systemmessage.ReadAtIsNil())...).
		SetReadAt(time.Now()).
		Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *messageRepo) MarkAllRead(ctx context.Context, userUID, email string) error {
	preds := messagePredicates(userUID, email)
	if len(preds) == 0 {
		return nil
	}
	preds = append(preds, systemmessage.ReadAtIsNil())
	if _, err := r.client.SystemMessage.Update().Where(preds...).SetReadAt(time.Now()).Save(ctx); err != nil {
		return r.fail(err)
	}
	return nil
}

// messagePredicates 按平台用户或邮箱匹配收件人。未注册邮箱的 user_uid 为空。
func messagePredicates(userUID, email string) []predicate.SystemMessage {
	email = strings.ToLower(strings.TrimSpace(email))
	switch {
	case userUID != "" && email != "":
		return []predicate.SystemMessage{systemmessage.Or(
			systemmessage.UserUIDEQ(userUID),
			systemmessage.EmailEQ(email),
		)}
	case userUID != "":
		return []predicate.SystemMessage{systemmessage.UserUIDEQ(userUID)}
	case email != "":
		return []predicate.SystemMessage{systemmessage.EmailEQ(email)}
	default:
		return nil
	}
}

func toSystemMessage(row *ent.SystemMessage) bo.SystemMessage {
	return bo.SystemMessage{
		UID:           row.UID,
		UserUID:       row.UserUID,
		Email:         row.Email,
		Kind:          string(row.Kind),
		ActorName:     row.ActorName,
		NamespaceUID:  row.NamespaceUID,
		NamespaceName: row.NamespaceName,
		RoleName:      row.RoleName,
		InviteUID:     row.InviteUID,
		ReadAt:        row.ReadAt,
		CreatedAt:     row.CreatedAt,
	}
}
