package biz

import (
	"context"
	stdjson "encoding/json"
	"slices"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func NewAlert(
	alertRecordRepo repository.AlertRecord,
	alertSubscriptionRepo repository.AlertSubscription,
	memberRepo repository.Member,
	recipientGroupBiz *RecipientGroup,
	emailBiz *Email,
	webhookBiz *Webhook,
	helper *klog.Helper,
) *Alert {
	return &Alert{
		alertRecordRepo:       alertRecordRepo,
		alertSubscriptionRepo: alertSubscriptionRepo,
		memberRepo:            memberRepo,
		recipientGroupBiz:     recipientGroupBiz,
		emailBiz:              emailBiz,
		webhookBiz:            webhookBiz,
		helper:                klog.NewHelper(klog.With(helper.Logger(), "biz", "alert")),
	}
}

type Alert struct {
	alertRecordRepo       repository.AlertRecord
	alertSubscriptionRepo repository.AlertSubscription
	memberRepo            repository.Member
	recipientGroupBiz     *RecipientGroup
	emailBiz              *Email
	webhookBiz            *Webhook
	helper                *klog.Helper
}

func (b *Alert) ReceivePrometheusWebhook(ctx context.Context, req *bo.ReceivePrometheusWebhookBo) ([]snowflake.ID, error) {
	namespaceUID, err := req.NamespaceUID()
	if err != nil {
		return nil, err
	}
	ctx = contextx.WithNamespace(ctx, namespaceUID)

	uids := make([]snowflake.ID, 0, len(req.Alerts))
	for _, alert := range req.Alerts {
		if alert == nil {
			continue
		}
		labels := mergeStringMap(req.CommonLabels, alert.Labels)
		annotations := mergeStringMap(req.CommonAnnotations, alert.Annotations)
		rawBytes, err := stdjson.Marshal(map[string]any{
			"version":           req.Version,
			"groupKey":          req.GroupKey,
			"status":            req.Status,
			"receiver":          req.Receiver,
			"externalURL":       req.ExternalURL,
			"groupLabels":       req.GroupLabels,
			"commonLabels":      req.CommonLabels,
			"commonAnnotations": req.CommonAnnotations,
			"source":            req.Source,
			"alert":             alert,
		})
		if err != nil {
			return nil, merr.ErrorInternalServer("marshal alert raw payload failed").WithCause(err)
		}
		recordUID, err := b.alertRecordRepo.CreateAlertRecord(ctx, &bo.CreateAlertRecordBo{
			Source:       req.Source,
			Receiver:     req.Receiver,
			Status:       alert.Status,
			Fingerprint:  alert.Fingerprint,
			GroupKey:     req.GroupKey,
			StartsAt:     alert.StartsAt,
			EndsAt:       alert.EndsAt,
			GeneratorURL: alert.GeneratorURL,
			Labels:       labels,
			Annotations:  annotations,
			Raw:          strEncrypt(rawBytes),
		})
		if err != nil {
			b.helper.WithContext(ctx).Errorw("msg", "create alert record failed", "error", err, "fingerprint", alert.Fingerprint)
			return nil, err
		}
		uids = append(uids, recordUID)
		record, err := b.alertRecordRepo.GetAlertRecord(ctx, recordUID)
		if err != nil {
			b.helper.WithContext(ctx).Errorw("msg", "get alert record failed", "error", err, "uid", recordUID)
			return nil, err
		}
		if err := b.dispatchAlertRecord(ctx, record); err != nil {
			b.helper.WithContext(ctx).Errorw("msg", "dispatch alert record failed", "error", err, "uid", recordUID)
		}
	}
	return uids, nil
}

func (b *Alert) GetAlertRecord(ctx context.Context, uid snowflake.ID) (*bo.AlertRecordItemBo, error) {
	return b.alertRecordRepo.GetAlertRecord(ctx, uid)
}

func (b *Alert) ListAlertRecord(ctx context.Context, req *bo.ListAlertRecordBo) (*bo.PageResponseBo[*bo.AlertRecordItemBo], error) {
	return b.alertRecordRepo.ListAlertRecord(ctx, req)
}

func (b *Alert) CreateAlertSubscription(ctx context.Context, req *bo.CreateAlertSubscriptionBo) (snowflake.ID, error) {
	if len(req.Labels) == 0 {
		return 0, merr.ErrorParams("labels are required")
	}
	if subscription, err := b.alertSubscriptionRepo.GetAlertSubscriptionByName(ctx, req.Name); err == nil {
		return 0, merr.ErrorParams("alert subscription %s already exists, uid: %d", req.Name, subscription.UID.Int64())
	} else if !merr.IsNotFound(err) {
		return 0, err
	}
	return b.alertSubscriptionRepo.CreateAlertSubscription(ctx, req)
}

func (b *Alert) UpdateAlertSubscription(ctx context.Context, req *bo.UpdateAlertSubscriptionBo) error {
	if len(req.Labels) == 0 {
		return merr.ErrorParams("labels are required")
	}
	return b.alertSubscriptionRepo.UpdateAlertSubscription(ctx, req)
}

func (b *Alert) DeleteAlertSubscription(ctx context.Context, uid snowflake.ID) error {
	return b.alertSubscriptionRepo.DeleteAlertSubscription(ctx, uid)
}

func (b *Alert) GetAlertSubscription(ctx context.Context, uid snowflake.ID) (*bo.AlertSubscriptionItemBo, error) {
	subscription, err := b.alertSubscriptionRepo.GetAlertSubscription(ctx, uid)
	if err != nil {
		return nil, err
	}
	if err := b.fillSubscriptionMembers(ctx, []*bo.AlertSubscriptionItemBo{subscription}); err != nil {
		return nil, err
	}
	return subscription, nil
}

func (b *Alert) ListAlertSubscription(ctx context.Context, req *bo.ListAlertSubscriptionBo) (*bo.PageResponseBo[*bo.AlertSubscriptionItemBo], error) {
	page, err := b.alertSubscriptionRepo.ListAlertSubscription(ctx, req)
	if err != nil {
		return nil, err
	}
	if err := b.fillSubscriptionMembers(ctx, page.GetItems()); err != nil {
		return nil, err
	}
	return page, nil
}

func (b *Alert) UpdateAlertSubscriptionStatus(ctx context.Context, req *bo.UpdateAlertSubscriptionStatusBo) error {
	return b.alertSubscriptionRepo.UpdateAlertSubscriptionStatus(ctx, req)
}

func (b *Alert) dispatchAlertRecord(ctx context.Context, record *bo.AlertRecordItemBo) error {
	subscriptions, err := b.alertSubscriptionRepo.ListEnabledAlertSubscriptions(ctx)
	if err != nil {
		return err
	}
	if err := b.fillSubscriptionMembers(ctx, subscriptions); err != nil {
		return err
	}
	for _, subscription := range subscriptions {
		if subscription == nil || !subscription.MatchesLabels(record.Labels) {
			continue
		}
		for _, recipientGroupUID := range subscription.RecipientGroupUIDs {
			if err := b.dispatchRecipientGroup(ctx, snowflake.ID(recipientGroupUID), record); err != nil {
				b.helper.WithContext(ctx).Warnw("msg", "dispatch recipient group failed", "error", err, "recipientGroupUID", recipientGroupUID)
			}
		}
		if err := b.dispatchSubscriptionMembers(ctx, subscription, record); err != nil {
			b.helper.WithContext(ctx).Warnw("msg", "dispatch subscription members failed", "error", err, "subscriptionUID", subscription.UID.Int64())
		}
	}
	return nil
}

func (b *Alert) dispatchRecipientGroup(ctx context.Context, uid snowflake.ID, record *bo.AlertRecordItemBo) error {
	group, err := b.recipientGroupBiz.GetRecipientGroup(ctx, uid)
	if err != nil {
		return err
	}
	groupItem := &group.RecipientGroupItemBo
	to := make([]string, 0, len(groupItem.Members))
	for _, member := range groupItem.Members {
		if member != nil && member.Email != "" {
			to = appendUnique(to, member.Email)
		}
	}
	for _, emailConfig := range groupItem.EmailConfigs {
		if emailConfig == nil || len(to) == 0 {
			continue
		}
		_, err := b.emailBiz.AppendEmailMessage(ctx, &bo.SendEmailBo{
			UID:         emailConfig.UID,
			Subject:     bo.BuildDefaultAlertSubject(record),
			Body:        bo.BuildDefaultAlertBody(record),
			To:          to,
			ContentType: "text/plain",
		})
		if err != nil {
			return err
		}
	}
	for _, webhookConfig := range groupItem.WebhookConfigs {
		if webhookConfig == nil {
			continue
		}
		_, err := b.webhookBiz.AppendWebhookMessage(ctx, &bo.SendWebhookBo{
			UID:  webhookConfig.UID,
			Data: bo.BuildAlertTemplateData(record),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Alert) dispatchSubscriptionMembers(ctx context.Context, subscription *bo.AlertSubscriptionItemBo, record *bo.AlertRecordItemBo) error {
	if subscription == nil || !subscription.DirectEmailEnabled() {
		return nil
	}
	to := make([]string, 0, len(subscription.Members))
	for _, member := range subscription.Members {
		if member != nil && member.IsEmail && member.MemberEmail != "" {
			to = appendUnique(to, member.MemberEmail)
		}
	}
	if len(to) == 0 {
		return nil
	}
	if subscription.DirectMemberTemplateUID > 0 {
		_, err := b.emailBiz.AppendEmailMessageWithTemplate(ctx, &bo.SendEmailWithTemplateBo{
			UID:         subscription.DirectMemberEmailConfigUID,
			TemplateUID: subscription.DirectMemberTemplateUID,
			JSONData:    []byte(bo.BuildAlertTemplateData(record)),
			To:          to,
		})
		return err
	}
	_, err := b.emailBiz.AppendEmailMessage(ctx, &bo.SendEmailBo{
		UID:         subscription.DirectMemberEmailConfigUID,
		Subject:     bo.BuildDefaultAlertSubject(record),
		Body:        bo.BuildDefaultAlertBody(record),
		To:          to,
		ContentType: "text/plain",
	})
	return err
}

func (b *Alert) fillSubscriptionMembers(ctx context.Context, subscriptions []*bo.AlertSubscriptionItemBo) error {
	memberUIDSet := make(map[int64]struct{})
	for _, subscription := range subscriptions {
		if subscription == nil {
			continue
		}
		for _, member := range subscription.Members {
			if member != nil && member.MemberUID > 0 {
				memberUIDSet[member.MemberUID] = struct{}{}
			}
		}
	}
	if len(memberUIDSet) == 0 {
		return nil
	}
	memberUIDs := make([]int64, 0, len(memberUIDSet))
	for uid := range memberUIDSet {
		memberUIDs = append(memberUIDs, uid)
	}
	slices.Sort(memberUIDs)
	memberMap := make(map[int64]*goddessv1.MemberItem, len(memberUIDs))
	const chunkSize = 200
	for i := 0; i < len(memberUIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(memberUIDs) {
			end = len(memberUIDs)
		}
		reply, err := b.memberRepo.ListMember(ctx, &goddessv1.ListMemberRequest{
			Page:     1,
			PageSize: chunkSize,
			Uids:     memberUIDs[i:end],
		})
		if err != nil {
			return err
		}
		for _, item := range reply.GetItems() {
			if item != nil {
				memberMap[item.GetUid()] = item
			}
		}
	}
	for _, subscription := range subscriptions {
		if subscription == nil {
			continue
		}
		for _, member := range subscription.Members {
			if member == nil {
				continue
			}
			item := memberMap[member.MemberUID]
			if item == nil {
				continue
			}
			member.MemberName = item.GetName()
			member.MemberAvatar = item.GetAvatar()
			member.MemberEmail = item.GetEmail()
			member.MemberPhone = item.GetPhone()
		}
	}
	return nil
}

func mergeStringMap(base map[string]string, extra map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(extra))
	for key, value := range base {
		result[key] = value
	}
	for key, value := range extra {
		result[key] = value
	}
	return result
}

func appendUnique(items []string, value string) []string {
	if value == "" {
		return items
	}
	if slices.Contains(items, value) {
		return items
	}
	return append(items, value)
}

func strEncrypt(value []byte) strutil.EncryptString {
	return strutil.EncryptString(value)
}
